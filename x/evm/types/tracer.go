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
package types

import (
	"encoding/json"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/eth/tracers/logger"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"

	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth/tracers"
)

const (
	TracerAccessList = "access_list"
	TracerJSON       = "json"
	TracerStruct     = "struct"
	TracerMarkdown   = "markdown"
)

// TxTraceResult is the result of a single transaction trace during a block trace.
type TxTraceResult struct {
	Result interface{} `json:"result,omitempty"` // Trace results produced by the tracer
	Error  string      `json:"error,omitempty"`  // Trace failure produced by the tracer
}

// noOpTracer is a go implementation of the Tracer interface which
// performs no action. It's mostly useful for testing purposes.
type NoOpTracer struct{}

// NewTracer creates a new Logger tracer to collect execution traces from an
// EVM transaction.
func NewTracer(tracer string, msg core.Message, cfg *params.ChainConfig, height int64) *tracers.Tracer {
	// TODO: enable additional log configuration
	logCfg := &logger.Config{
		Debug: true,
	}

	switch tracer {
	case TracerAccessList:
		preCompiles := vm.DefaultActivePrecompiles(cfg.Rules(big.NewInt(height), cfg.MergeNetsplitBlock != nil, 0))
		lgr := logger.NewAccessListTracer(msg.AccessList, msg.From, *msg.To, preCompiles)
		tracer := &tracers.Tracer{
			Hooks: lgr.Hooks(),
		}
		return tracer
	case TracerJSON:
		tracer := &tracers.Tracer{
			Hooks: logger.NewJSONLogger(logCfg, os.Stderr),
		}
		return tracer
	case TracerMarkdown:
		lgr := logger.NewMarkdownLogger(logCfg, os.Stdout)
		tracer := &tracers.Tracer{
			Hooks: lgr.Hooks(),
		}
		return tracer
	case TracerStruct:
		lgr := logger.NewStructLogger(logCfg)
		tracer := &tracers.Tracer{
			Hooks:     lgr.Hooks(),
			GetResult: lgr.GetResult,
			Stop:      lgr.Stop,
		}
		return tracer
	default:
		tracer, _ := NewNoopTracer()
		return tracer
	}
}

// newNoopTracer returns a new noop tracer.
func NewNoopTracer() (*tracers.Tracer, error) {
	t := &NoOpTracer{}
	return &tracers.Tracer{
		Hooks: &tracing.Hooks{
			OnTxStart:       t.OnTxStart,
			OnTxEnd:         t.OnTxEnd,
			OnEnter:         t.OnEnter,
			OnExit:          t.OnExit,
			OnOpcode:        t.OnOpcode,
			OnFault:         t.OnFault,
			OnGasChange:     t.OnGasChange,
			OnBalanceChange: t.OnBalanceChange,
			OnNonceChange:   t.OnNonceChange,
			OnCodeChange:    t.OnCodeChange,
			OnStorageChange: t.OnStorageChange,
			OnLog:           t.OnLog,
		},
		GetResult: t.GetResult,
		Stop:      t.Stop,
	}, nil
}

func (t *NoOpTracer) OnOpcode(_ uint64, _ byte, _, _ uint64, _ tracing.OpContext, _ []byte, _ int, _ error) {
}

func (t *NoOpTracer) OnFault(_ uint64, _ byte, _, _ uint64, _ tracing.OpContext, _ int, _ error) {
}

func (t *NoOpTracer) OnGasChange(_, _ uint64, _ tracing.GasChangeReason) {}

func (t *NoOpTracer) OnEnter(_ int, _ byte, _ common.Address, _ common.Address, _ []byte, _ uint64, _ *big.Int) {
}

func (t *NoOpTracer) OnExit(_ int, _ []byte, _ uint64, _ error, _ bool) {
}

func (*NoOpTracer) OnTxStart(_ *tracing.VMContext, _ *types.Transaction, _ common.Address) {
}

func (*NoOpTracer) OnTxEnd(_ *types.Receipt, _ error) {}

func (*NoOpTracer) OnBalanceChange(_ common.Address, _, _ *big.Int, _ tracing.BalanceChangeReason) {
}

func (*NoOpTracer) OnNonceChange(_ common.Address, _, _ uint64) {}

func (*NoOpTracer) OnCodeChange(_ common.Address, _ common.Hash, _ []byte, _ common.Hash, _ []byte) {
}

func (*NoOpTracer) OnStorageChange(_ common.Address, _, _, _ common.Hash) {}

func (*NoOpTracer) OnLog(_ *types.Log) {}

// GetResult returns an empty json object.
func (t *NoOpTracer) GetResult() (json.RawMessage, error) {
	return json.RawMessage(`{}`), nil
}

// Stop terminates execution of the tracer at the first opportune moment.
func (t *NoOpTracer) Stop(_ error) {
}
