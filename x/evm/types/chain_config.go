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
		//TODO(xlab): after upgrading ethereum to newer version, this should be set to YoloV2Block
		YoloV2Block: getBlockValue(cc.YoloV2Block),
		EWASMBlock:  getBlockValue(cc.EWASMBlock),
	}
}

// DefaultChainConfig returns default evm parameters. Th
func DefaultChainConfig() ChainConfig {
	return ChainConfig{
		HomesteadBlock:      sdk.ZeroInt(),
		DAOForkBlock:        sdk.ZeroInt(),
		DAOForkSupport:      true,
		EIP150Block:         sdk.ZeroInt(),
		EIP150Hash:          common.Hash{}.String(),
		EIP155Block:         sdk.ZeroInt(),
		EIP158Block:         sdk.ZeroInt(),
		ByzantiumBlock:      sdk.ZeroInt(),
		ConstantinopleBlock: sdk.ZeroInt(),
		PetersburgBlock:     sdk.ZeroInt(),
		IstanbulBlock:       sdk.NewInt(-1),
		MuirGlacierBlock:    sdk.NewInt(-1),
		YoloV2Block:         sdk.NewInt(-1),
		EWASMBlock:          sdk.NewInt(-1),
	}
}

func getBlockValue(block sdk.Int) *big.Int {
	if block.IsNegative() {
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
	if err := validateBlock(cc.YoloV2Block); err != nil {
		return sdkerrors.Wrap(err, "yoloV2Block")
	}
	if err := validateBlock(cc.EWASMBlock); err != nil {
		return sdkerrors.Wrap(err, "eWASMBlock")
	}

	return nil
}

func validateHash(hex string) error {
	if hex != "" && strings.TrimSpace(hex) == "" {
		return sdkerrors.Wrapf(ErrInvalidChainConfig, "hash cannot be blank")
	}

	return nil
}

func validateBlock(block sdk.Int) error {
	if block == (sdk.Int{}) || block.BigInt() == nil {
		return sdkerrors.Wrapf(
			ErrInvalidChainConfig,
			"cannot use uninitialized or nil values for Int, set a negative Int value if you want to define a nil Ethereum block",
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
