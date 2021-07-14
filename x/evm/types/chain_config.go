package types

import (
	"math/big"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

// EthereumConfig returns an Ethereum ChainConfig for EVM state transitions.
// All the negative or nil values are converted to nil
func (cc ChainConfig) EthereumConfig(chainID *big.Int) *params.ChainConfig {
	return &params.ChainConfig{
		ChainID:             chainID,
		HomesteadBlock:      getBlockValue(cc.HomesteadBlock),
		DAOForkBlock:        getBlockValue(cc.DAOForkBlock),
		DAOForkSupport:      cc.DAOForkSupport,
		EIP150Block:         getBlockValue(cc.EIP150Block),
		EIP150Hash:          common.HexToHash(cc.EIP150Hash),
		EIP155Block:         getBlockValue(cc.EIP155Block),
		EIP158Block:         getBlockValue(cc.EIP158Block),
		ByzantiumBlock:      getBlockValue(cc.ByzantiumBlock),
		ConstantinopleBlock: getBlockValue(cc.ConstantinopleBlock),
		PetersburgBlock:     getBlockValue(cc.PetersburgBlock),
		IstanbulBlock:       getBlockValue(cc.IstanbulBlock),
		MuirGlacierBlock:    getBlockValue(cc.MuirGlacierBlock),
		BerlinBlock:         getBlockValue(cc.BerlinBlock),
		YoloV3Block:         getBlockValue(cc.YoloV3Block),
		EWASMBlock:          getBlockValue(cc.EWASMBlock),
		CatalystBlock:       getBlockValue(cc.CatalystBlock),
	}
}

// DefaultChainConfig returns default evm parameters.
func DefaultChainConfig() ChainConfig {
	homesteadBlock := sdk.ZeroInt()
	daoForkBlock := sdk.ZeroInt()
	eip150Block := sdk.ZeroInt()
	eip155Block := sdk.ZeroInt()
	eip158Block := sdk.ZeroInt()
	byzantiumBlock := sdk.ZeroInt()
	constantinopleBlock := sdk.ZeroInt()
	petersburgBlock := sdk.ZeroInt()
	istanbulBlock := sdk.ZeroInt()
	muirGlacierBlock := sdk.ZeroInt()
	berlinBlock := sdk.ZeroInt()
	yoloV3Block := sdk.ZeroInt()

	return ChainConfig{
		HomesteadBlock:      &homesteadBlock,
		DAOForkBlock:        &daoForkBlock,
		DAOForkSupport:      true,
		EIP150Block:         &eip150Block,
		EIP150Hash:          common.Hash{}.String(),
		EIP155Block:         &eip155Block,
		EIP158Block:         &eip158Block,
		ByzantiumBlock:      &byzantiumBlock,
		ConstantinopleBlock: &constantinopleBlock,
		PetersburgBlock:     &petersburgBlock,
		IstanbulBlock:       &istanbulBlock,
		MuirGlacierBlock:    &muirGlacierBlock,
		BerlinBlock:         &berlinBlock,
		YoloV3Block:         &yoloV3Block,
		EWASMBlock:          nil,
		CatalystBlock:       nil,
	}
}

func getBlockValue(block *sdk.Int) *big.Int {
	if block == nil || block.IsNegative() {
		return nil
	}

	return block.BigInt()
}

// Validate performs a basic validation of the ChainConfig params. The function will return an error
// if any of the block values is uninitialized (i.e nil) or if the EIP150Hash is an invalid hash.
func (cc ChainConfig) Validate() error {
	if err := validateBlock(cc.HomesteadBlock); err != nil {
		return sdkerrors.Wrap(err, "homesteadBlock")
	}
	if err := validateBlock(cc.DAOForkBlock); err != nil {
		return sdkerrors.Wrap(err, "daoForkBlock")
	}
	if err := validateBlock(cc.EIP150Block); err != nil {
		return sdkerrors.Wrap(err, "eip150Block")
	}
	if err := validateHash(cc.EIP150Hash); err != nil {
		return err
	}
	if err := validateBlock(cc.EIP155Block); err != nil {
		return sdkerrors.Wrap(err, "eip155Block")
	}
	if err := validateBlock(cc.EIP158Block); err != nil {
		return sdkerrors.Wrap(err, "eip158Block")
	}
	if err := validateBlock(cc.ByzantiumBlock); err != nil {
		return sdkerrors.Wrap(err, "byzantiumBlock")
	}
	if err := validateBlock(cc.ConstantinopleBlock); err != nil {
		return sdkerrors.Wrap(err, "constantinopleBlock")
	}
	if err := validateBlock(cc.PetersburgBlock); err != nil {
		return sdkerrors.Wrap(err, "petersburgBlock")
	}
	if err := validateBlock(cc.IstanbulBlock); err != nil {
		return sdkerrors.Wrap(err, "istanbulBlock")
	}
	if err := validateBlock(cc.MuirGlacierBlock); err != nil {
		return sdkerrors.Wrap(err, "muirGlacierBlock")
	}
	if err := validateBlock(cc.BerlinBlock); err != nil {
		return sdkerrors.Wrap(err, "berlinBlock")
	}
	if err := validateBlock(cc.YoloV3Block); err != nil {
		return sdkerrors.Wrap(err, "yoloV3Block")
	}
	if err := validateBlock(cc.EWASMBlock); err != nil {
		return sdkerrors.Wrap(err, "eWASMBlock")
	}
	if err := validateBlock(cc.CatalystBlock); err != nil {
		return sdkerrors.Wrap(err, "calalystBlock")
	}

	return nil
}

func validateHash(hex string) error {
	if hex != "" && strings.TrimSpace(hex) == "" {
		return sdkerrors.Wrapf(ErrInvalidChainConfig, "hash cannot be blank")
	}

	return nil
}

func validateBlock(block *sdk.Int) error {
	// nil value means that the fork has not yet been applied
	if block == nil {
		return nil
	}

	if block.IsNegative() {
		return sdkerrors.Wrapf(
			ErrInvalidChainConfig, "block value cannot be negative: %s", block,
		)
	}

	return nil
}

// IsIstanbul returns whether the Istanbul version is enabled.
func (cc ChainConfig) IsIstanbul() bool {
	return getBlockValue(cc.IstanbulBlock) != nil
}

// IsHomestead returns whether the Homestead version is enabled.
func (cc ChainConfig) IsHomestead() bool {
	return getBlockValue(cc.HomesteadBlock) != nil
}
