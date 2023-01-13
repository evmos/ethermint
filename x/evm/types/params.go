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
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/params"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/ethermint/types"
)

var (
	// DefaultEVMDenom defines the default EVM denomination on Ethermint
	DefaultEVMDenom = types.AttoPhoton
	// DefaultAllowUnprotectedTxs rejects all unprotected txs (i.e false)
	DefaultAllowUnprotectedTxs = false
	// DefaultEnableCreate enables contract creation (i.e true)
	DefaultEnableCreate = true
	// DefaultEnableCall enables contract calls (i.e true)
	DefaultEnableCall = true
	// DefaultExtraEIPs  defines the set of activateable Ethereum Improvement Proposals
	DefaultExtraEIPs = ExtraEIPs{AvailableExtraEIPs}
)

// Parameter keys
var (
	ParamStoreKeyEVMDenom            = []byte("EVMDenom")
	ParamStoreKeyEnableCreate        = []byte("EnableCreate")
	ParamStoreKeyEnableCall          = []byte("EnableCall")
	ParamStoreKeyExtraEIPs           = []byte("EnableExtraEIPs")
	ParamStoreKeyChainConfig         = []byte("ChainConfig")
	ParamStoreKeyAllowUnprotectedTxs = []byte("AllowUnprotectedTxs")

	// AvailableExtraEIPs define the list of all EIPs that can be enabled by the
	// EVM interpreter. These EIPs are applied in order and can override the
	// instruction sets from the latest hard fork enabled by the ChainConfig. For
	// more info check:
	// https://github.com/ethereum/go-ethereum/blob/master/core/vm/interpreter.go#L97
	AvailableExtraEIPs = []int64{1344, 1884, 2200, 2929, 3198, 3529}
)

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(evmDenom string, allowUnprotectedTxs, enableCreate, enableCall bool, config ChainConfig, extraEIPs ExtraEIPs) Params {
	return Params{
		EvmDenom:            evmDenom,
		AllowUnprotectedTxs: allowUnprotectedTxs,
		EnableCreate:        enableCreate,
		EnableCall:          enableCall,
		ExtraEIPs:           extraEIPs,
		ChainConfig:         config,
	}
}

// DefaultParams returns default evm parameters
// ExtraEIPs is empty to prevent overriding the latest hard fork instruction set
func DefaultParams() Params {
	return Params{
		EvmDenom:            DefaultEVMDenom,
		EnableCreate:        DefaultEnableCreate,
		EnableCall:          DefaultEnableCall,
		ChainConfig:         DefaultChainConfig(),
		ExtraEIPs:           DefaultExtraEIPs,
		AllowUnprotectedTxs: DefaultAllowUnprotectedTxs,
	}
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyEVMDenom, &p.EvmDenom, validateEVMDenom),
		paramtypes.NewParamSetPair(ParamStoreKeyEnableCreate, &p.EnableCreate, validateBool),
		paramtypes.NewParamSetPair(ParamStoreKeyEnableCall, &p.EnableCall, validateBool),
		paramtypes.NewParamSetPair(ParamStoreKeyExtraEIPs, &p.ExtraEIPs.EIPs, validateEIPs),
		paramtypes.NewParamSetPair(ParamStoreKeyChainConfig, &p.ChainConfig, validateChainConfig),
		paramtypes.NewParamSetPair(ParamStoreKeyAllowUnprotectedTxs, &p.AllowUnprotectedTxs, validateBool),
	}
}

// Validate performs basic validation on evm parameters.
func (p Params) Validate() error {
	if err := validateEVMDenom(p.EvmDenom); err != nil {
		return err
	}

	if err := validateEIPs(p.ExtraEIPs.EIPs); err != nil {
		return err
	}

	if err := validateBool(p.EnableCall); err != nil {
		return err
	}

	if err := validateBool(p.EnableCreate); err != nil {
		return err
	}

	if err := validateBool(p.AllowUnprotectedTxs); err != nil {
		return err
	}

	return validateChainConfig(p.ChainConfig)
}

// EIPs returns the ExtraEIPS as a int slice
func (p Params) EIPs() []int {
	eips := make([]int, len(p.ExtraEIPs.EIPs))
	for i, eip := range p.ExtraEIPs.EIPs {
		eips[i] = int(eip)
	}
	return eips
}

func validateEVMDenom(i interface{}) error {
	denom, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter EVM denom type: %T", i)
	}

	return sdk.ValidateDenom(denom)
}

func validateBool(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateEIPs(i interface{}) error {
	eips, ok := i.([]int64)
	if !ok {
		return fmt.Errorf("invalid EIP slice type: %T", i)
	}

	for _, eip := range eips {
		if !vm.ValidEip(int(eip)) {
			return fmt.Errorf("EIP %d is not activateable, valid EIPS are: %s", eip, vm.ActivateableEips())
		}
	}

	return nil
}

func validateChainConfig(i interface{}) error {
	cfg, ok := i.(ChainConfig)
	if !ok {
		return fmt.Errorf("invalid chain config type: %T", i)
	}

	return cfg.Validate()
}

// IsLondon returns if london hardfork is enabled.
func IsLondon(ethConfig *params.ChainConfig, height int64) bool {
	return ethConfig.IsLondon(big.NewInt(height))
}
