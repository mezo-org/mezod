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
func NewTracer(
	tracer string,
	msg core.Message,
	cfg *params.ChainConfig,
	height int64,
	timestamp uint64,
	customPrecompiles func() []common.Address,
) *tracers.Tracer {
	logCfg := &logger.Config{}

	switch tracer {
	case TracerAccessList:
		addressesToExclude := accessListTracerExclusions(
			msg,
			cfg,
			height,
			timestamp,
			customPrecompiles(),
		)
		lgr := logger.NewAccessListTracer(msg.AccessList, addressesToExclude)
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

func accessListTracerExclusions(
	msg core.Message,
	cfg *params.ChainConfig,
	height int64,
	timestamp uint64,
	customPrecompiles []common.Address,
) map[common.Address]struct{} {
	precompiles := vm.ActivePrecompiles(
		cfg.Rules(big.NewInt(height), cfg.MergeNetsplitBlock != nil, timestamp),
	)
	addressesToExclude := map[common.Address]struct{}{msg.From: {}}
	if msg.To != nil {
		addressesToExclude[*msg.To] = struct{}{}
	}
	for _, addr := range precompiles {
		addressesToExclude[addr] = struct{}{}
	}
	for _, addr := range customPrecompiles {
		addressesToExclude[addr] = struct{}{}
	}
	return addressesToExclude
}

// newNoopTracer returns a new noop tracer.
func NewNoopTracer() (*tracers.Tracer, error) {
	t := &NoOpTracer{}
	return &tracers.Tracer{
		Hooks:     &tracing.Hooks{},
		GetResult: t.GetResult,
		Stop:      t.Stop,
	}, nil
}

// GetResult returns an empty json object.
func (t *NoOpTracer) GetResult() (json.RawMessage, error) {
	return json.RawMessage(`{}`), nil
}

// Stop terminates execution of the tracer at the first opportune moment.
func (t *NoOpTracer) Stop(_ error) {
}
