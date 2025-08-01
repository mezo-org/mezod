// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE
package filters

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"

	"cosmossdk.io/log"
	tmjson "github.com/cometbft/cometbft/libs/json"
	tmquery "github.com/cometbft/cometbft/libs/pubsub/query"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	rpcclient "github.com/cometbft/cometbft/rpc/jsonrpc/client"
	tmtypes "github.com/cometbft/cometbft/types"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth/filters"
	"github.com/ethereum/go-ethereum/rpc"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/mezo-org/mezod/rpc/ethereum/pubsub"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

var (
	txEvents  = tmtypes.QueryForEvent(tmtypes.EventTx).String()
	evmEvents = tmquery.MustCompile(fmt.Sprintf("%s='%s' AND %s.%s='%s'",
		tmtypes.EventTypeKey,
		tmtypes.EventTx,
		sdk.EventTypeMessage,
		sdk.AttributeKeyModule, evmtypes.ModuleName)).String()
	headerEvents = tmtypes.QueryForEvent(tmtypes.EventNewBlockHeader).String()
)

// EventSystem creates subscriptions, processes events and broadcasts them to the
// subscription which match the subscription criteria using the Tendermint's RPC client.
type EventSystem struct {
	logger     log.Logger
	ctx        context.Context
	tmWSClient *rpcclient.WSClient

	// light client mode
	lightMode bool

	index      filterIndex
	topicChans map[string]chan<- coretypes.ResultEvent
	indexMux   *sync.RWMutex

	// Channels
	install   chan *Subscription // install filter for event notification
	uninstall chan *Subscription // remove filter for event notification
	eventBus  pubsub.EventBus
}

// NewEventSystem creates a new manager that listens for event on the given mux,
// parses and filters them. It uses the all map to retrieve filter changes. The
// work loop holds its own index that is used to forward events to filters.
//
// The returned manager has a loop that needs to be stopped with the Stop function
// or by stopping the given mux.
func NewEventSystem(logger log.Logger, tmWSClient *rpcclient.WSClient) *EventSystem {
	index := make(filterIndex)
	for i := filters.UnknownSubscription; i < filters.LastIndexSubscription; i++ {
		index[i] = make(map[rpc.ID]*Subscription)
	}

	es := &EventSystem{
		logger:     logger,
		ctx:        context.Background(),
		tmWSClient: tmWSClient,
		lightMode:  false,
		index:      index,
		topicChans: make(map[string]chan<- coretypes.ResultEvent, len(index)),
		indexMux:   new(sync.RWMutex),
		install:    make(chan *Subscription),
		uninstall:  make(chan *Subscription),
		eventBus:   pubsub.NewEventBus(),
	}

	go es.eventLoop()
	go es.consumeEvents()
	return es
}

// WithContext sets a new context to the EventSystem. This is required to set a timeout context when
// a new filter is intantiated.
func (es *EventSystem) WithContext(ctx context.Context) {
	es.ctx = ctx
}

// subscribe performs a new event subscription to a given Tendermint event.
// The subscription creates a unidirectional receive event channel to receive the ResultEvent.
func (es *EventSystem) subscribe(sub *Subscription) (*Subscription, pubsub.UnsubscribeFunc, error) {
	var (
		err      error
		cancelFn context.CancelFunc
	)

	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	existingSubs := es.eventBus.Topics()
	for _, topic := range existingSubs {
		if topic == sub.event {
			eventCh, unsubFn, err := es.eventBus.Subscribe(sub.event)
			if err != nil {
				err := errors.Wrapf(err, "failed to subscribe to topic: %s", sub.event)
				sub.err <- err
				return nil, nil, err
			}

			// wrap events in a go routine to prevent blocking
			es.install <- sub
			<-sub.installed

			sub.eventCh = eventCh
			return sub, unsubFn, nil
		}
	}

	switch sub.typ {
	case filters.LogsSubscription:
		err = es.tmWSClient.Subscribe(ctx, sub.event)
	case filters.BlocksSubscription:
		err = es.tmWSClient.Subscribe(ctx, sub.event)
	case filters.PendingTransactionsSubscription:
		err = es.tmWSClient.Subscribe(ctx, sub.event)
	default:
		err = fmt.Errorf("invalid filter subscription type %d", sub.typ)
	}

	if err != nil {
		sub.err <- err
		return nil, nil, err
	}

	// wrap events in a go routine to prevent blocking
	es.install <- sub
	<-sub.installed

	eventCh, unsubFn, err := es.eventBus.Subscribe(sub.event)
	if err != nil {
		sub.err <- err
		return nil, nil, errors.Wrapf(err, "failed to subscribe to topic after installed: %s", sub.event)
	}

	sub.eventCh = eventCh
	return sub, unsubFn, nil
}

// SubscribeLogs creates a subscription that will write all logs matching the
// given criteria to the given logs channel. Default value for the from and to
// block is "latest". If the fromBlock > toBlock an error is returned.

func (es *EventSystem) SubscribeLogs(crit filters.FilterCriteria) (*Subscription, pubsub.UnsubscribeFunc, error) {
	if len(crit.Addresses) > defaultMaxAddressesFilter {
		return nil, nil, fmt.Errorf("max number of addresses exceeded (max allowed %v)", defaultMaxAddressesFilter)
	}

	var from, to rpc.BlockNumber
	if crit.FromBlock == nil {
		from = rpc.LatestBlockNumber
	} else {
		from = rpc.BlockNumber(crit.FromBlock.Int64())
	}
	if crit.ToBlock == nil {
		to = rpc.LatestBlockNumber
	} else {
		to = rpc.BlockNumber(crit.ToBlock.Int64())
	}

	switch {
	// only interested in new mined logs, mined logs within a specific block range, or
	// logs from a specific block number to new mined blocks
	case (from == rpc.LatestBlockNumber && to == rpc.LatestBlockNumber),
		(from >= 0 && to >= 0 && to >= from),
		(from >= 0 && to == rpc.LatestBlockNumber):
		return es.subscribeLogs(crit)

	default:
		return nil, nil, fmt.Errorf("invalid from and to block combination: from > to (%d > %d)", from, to)
	}
}

// subscribeLogs creates a subscription that will write all logs matching the
// given criteria to the given logs channel.
func (es *EventSystem) subscribeLogs(crit filters.FilterCriteria) (*Subscription, pubsub.UnsubscribeFunc, error) {
	sub := &Subscription{
		id:        rpc.NewID(),
		typ:       filters.LogsSubscription,
		event:     evmEvents,
		logsCrit:  crit,
		created:   time.Now().UTC(),
		logs:      make(chan []*ethtypes.Log),
		installed: make(chan struct{}, 1),
		err:       make(chan error, 1),
	}

	return es.subscribe(sub)
}

// SubscribeNewHeads subscribes to new block headers events.
func (es EventSystem) SubscribeNewHeads() (*Subscription, pubsub.UnsubscribeFunc, error) {
	sub := &Subscription{
		id:        rpc.NewID(),
		typ:       filters.BlocksSubscription,
		event:     headerEvents,
		created:   time.Now().UTC(),
		headers:   make(chan *ethtypes.Header),
		installed: make(chan struct{}, 1),
		err:       make(chan error, 1),
	}
	return es.subscribe(sub)
}

// SubscribePendingTxs subscribes to new pending transactions events from the mempool.
func (es EventSystem) SubscribePendingTxs() (*Subscription, pubsub.UnsubscribeFunc, error) {
	sub := &Subscription{
		id:        rpc.NewID(),
		typ:       filters.PendingTransactionsSubscription,
		event:     txEvents,
		created:   time.Now().UTC(),
		hashes:    make(chan []common.Hash),
		installed: make(chan struct{}, 1),
		err:       make(chan error, 1),
	}
	return es.subscribe(sub)
}

type filterIndex map[filters.Type]map[rpc.ID]*Subscription

// eventLoop (un)installs filters and processes mux events.
func (es *EventSystem) eventLoop() {
	for {
		select {
		case f := <-es.install:
			es.logger.Debug("installing subscription", "subId", f.ID())
			es.indexMux.Lock()
			es.index[f.typ][f.id] = f
			ch := make(chan coretypes.ResultEvent)
			if err := es.eventBus.AddTopic(f.event, ch); err != nil {
				// Just a log here, error can be that we already have created
				// the topic
				es.logger.Debug("failed to add event topic to event bus", "topic", f.event, "error", err.Error())
			} else {
				// topic didn't exists, add it to the map
				es.topicChans[f.event] = ch
			}
			es.indexMux.Unlock()
			close(f.installed)
		case f := <-es.uninstall:
			es.logger.Debug("uninstalling subscription", "subId", f.ID())
			es.indexMux.Lock()
			delete(es.index[f.typ], f.id)

			var channelInUse bool
			// #nosec G705
			for _, sub := range es.index[f.typ] {
				if sub.event == f.event {
					channelInUse = true
					break
				}
			}

			// remove topic only when channel is not used by other subscriptions
			if !channelInUse {
				es.logger.Debug("topic not used by any channel", "query", f.event)

				if err := es.tmWSClient.Unsubscribe(es.ctx, f.event); err != nil {
					es.logger.Error("failed to unsubscribe from query", "query", f.event, "error", err.Error())
				}

				ch, ok := es.topicChans[f.event]
				if ok {
					es.eventBus.RemoveTopic(f.event)
					close(ch)
					delete(es.topicChans, f.event)
				}
			}

			es.indexMux.Unlock()
			close(f.err)
		}
	}
}

func (es *EventSystem) consumeEvents() {
	for {
		for rpcResp := range es.tmWSClient.ResponsesCh {
			var ev coretypes.ResultEvent

			if rpcResp.Error != nil {
				time.Sleep(5 * time.Second)
				continue
			} else if err := tmjson.Unmarshal(rpcResp.Result, &ev); err != nil {
				es.logger.Error("failed to JSON unmarshal ResponsesCh result event", "error", err.Error())
				continue
			}

			if len(ev.Query) == 0 {
				// skip empty responses
				continue
			}

			es.indexMux.RLock()
			ch, ok := es.topicChans[ev.Query]
			es.indexMux.RUnlock()
			if !ok {
				es.logger.Debug("channel for subscription not found", "topic", ev.Query)
				es.logger.Debug("list of available channels", "channels", es.eventBus.Topics())
				continue
			}

			// gracefully handle lagging subscribers
			t := time.NewTimer(time.Second)
			select {
			case <-t.C:
				es.logger.Debug("dropped event during lagging subscription", "topic", ev.Query)
			case ch <- ev:
			}
		}

		time.Sleep(time.Second)
	}
}
