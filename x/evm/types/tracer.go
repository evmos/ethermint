// Copyright 2021 Evmos Foundation
// This file is part of Evmos' Ethermint library.
//
// The Ethermint library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Ethermint library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Ethermint library. If not, see https://github.com/evmos/ethermint/blob/main/LICENSE
package types

import (
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethlogger "github.com/ethereum/go-ethereum/eth/tracers/logger"

	"github.com/evmos/ethermint/ethereum/tracers/logger"
	evm "github.com/evmos/ethermint/x/evm/vm"
)

const (
	TracerAccessList = "access_list"
	TracerJSON       = "json"
	TracerStruct     = "struct"
	TracerMarkdown   = "markdown"
)

// NewTracer creates a new Logger tracer to collect execution traces from an
// EVM transaction.
func NewTracer(
	tracer string,
	msg core.Message,
	precompiles ...common.Address,
) evm.Logger {
	// TODO: enable additional log configuration
	logCfg := &ethlogger.Config{
		Debug: true,
	}

	switch tracer {
	case TracerAccessList:
		return logger.NewAccessListTracer(msg.AccessList(), msg.From(), *msg.To(), precompiles)
	case TracerJSON:
		return logger.NewJSONLogger(logCfg, os.Stderr)
	case TracerMarkdown:
		return logger.NewMarkdownLogger(logCfg, os.Stdout) // TODO: Stderr ?
	case TracerStruct:
		return logger.NewStructLogger(logCfg)
	default:
		return logger.NewNoOpTracer()
	}
}

// TxTraceResult is the result of a single transaction trace during a block trace.
type TxTraceResult struct {
	Result interface{} `json:"result,omitempty"` // Trace results produced by the tracer
	Error  string      `json:"error,omitempty"`  // Trace failure produced by the tracer
}
