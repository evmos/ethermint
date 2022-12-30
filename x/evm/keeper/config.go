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
package keeper

import (
	"math/big"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"

	"github.com/evmos/ethermint/ethereum/tracers/logger"
	"github.com/evmos/ethermint/x/evm/statedb"
	"github.com/evmos/ethermint/x/evm/types"
	evm "github.com/evmos/ethermint/x/evm/vm"
)

// Tracer returns an EVM Tracer (Logger) based on current Keeper state.
func (k Keeper) Tracer(ctx sdk.Context, msg core.Message, ethCfg *params.ChainConfig) evm.Logger {
	rules := ethCfg.Rules(big.NewInt(ctx.BlockHeight()), ethCfg.MergeNetsplitBlock != nil)
	precompiles := vm.ActivePrecompiles(rules)
	return types.NewTracer(k.tracer, msg, precompiles...)
}

// EVMConfig creates the EVMConfig based on current state
func (k *Keeper) EVMConfig(ctx sdk.Context, proposerAddress sdk.ConsAddress, chainID *big.Int) (*statedb.EVMConfig, error) {
	params := k.GetParams(ctx)
	ethCfg := params.ChainConfig.EthereumConfig(chainID)

	// get the coinbase address from the block proposer
	coinbase, err := k.GetCoinbaseAddress(ctx, proposerAddress)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to obtain coinbase address")
	}

	baseFee := k.GetBaseFee(ctx, ethCfg)
	return &statedb.EVMConfig{
		Params:      params,
		ChainConfig: ethCfg,
		CoinBase:    coinbase,
		BaseFee:     baseFee,
	}, nil
}

// TxConfig loads creates a new transaction configuration using the block
// transient storage and the context.
func (k *Keeper) TxConfig(ctx sdk.Context, txHash common.Hash) statedb.TxConfig {
	return statedb.NewTxConfig(
		common.BytesToHash(ctx.HeaderHash()), // BlockHash
		txHash,                               // TxHash
		uint(k.GetTxIndexTransient(ctx)),     // TxIndex
		uint(k.GetLogSizeTransient(ctx)),     // LogIndex
	)
}

// VMConfig creates an EVM configuration from the debug setting and the extra EIPs enabled on the
// module parameters. The config generated uses the default JumpTable from the EVM.
func (k Keeper) VMConfig(ctx sdk.Context, msg core.Message, cfg *statedb.EVMConfig, tracer evm.Logger) evm.Config {
	noBaseFee := true
	if types.IsLondon(cfg.ChainConfig, ctx.BlockHeight()) {
		noBaseFee = k.feeMarketKeeper.GetParams(ctx).NoBaseFee
	}

	var debug bool
	if _, ok := tracer.(logger.NoOpTracer); !ok {
		debug = true
	}

	return evm.NewConfig(debug, tracer, noBaseFee, false, nil, cfg.Params.EIPs()...)
}
