package types

import (
	"math/big"
	"strings"

	"gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

// ChainConfig defines the Ethereum ChainConfig parameters using sdk.Int values instead of big.Int.
//
// NOTE 1: Since empty/uninitialized Ints (i.e with a nil big.Int value) are parsed to zero, we need to manually
// specify that negative Int values will be considered as nil. See getBlockValue for reference.
//
// NOTE 2: This type is not a configurable Param since the SDK does not allow for validation against
// a previous stored parameter values or the current block height (retrieved from context). If you
// want to update the config values, use an software upgrade procedure.
type ChainConfig struct {
	HomesteadBlock sdk.Int `json:"homestead_block" yaml:"homestead_block"` // Homestead switch block (< 0 no fork, 0 = already homestead)

	DAOForkBlock   sdk.Int `json:"dao_fork_block" yaml:"dao_fork_block"`     // TheDAO hard-fork switch block (< 0 no fork)
	DAOForkSupport bool    `json:"dao_fork_support" yaml:"dao_fork_support"` // Whether the nodes supports or opposes the DAO hard-fork

	// EIP150 implements the Gas price changes (https://github.com/ethereum/EIPs/issues/150)
	EIP150Block sdk.Int `json:"eip150_block" yaml:"eip150_block"` // EIP150 HF block (< 0 no fork)
	EIP150Hash  string  `json:"eip150_hash" yaml:"eip150_hash"`   // EIP150 HF hash (needed for header only clients as only gas pricing changed)

	EIP155Block sdk.Int `json:"eip155_block" yaml:"eip155_block"` // EIP155 HF block
	EIP158Block sdk.Int `json:"eip158_block" yaml:"eip158_block"` // EIP158 HF block

	ByzantiumBlock      sdk.Int `json:"byzantium_block" yaml:"byzantium_block"`           // Byzantium switch block (< 0 no fork, 0 = already on byzantium)
	ConstantinopleBlock sdk.Int `json:"constantinople_block" yaml:"constantinople_block"` // Constantinople switch block (< 0 no fork, 0 = already activated)
	PetersburgBlock     sdk.Int `json:"petersburg_block" yaml:"petersburg_block"`         // Petersburg switch block (< 0 same as Constantinople)
	IstanbulBlock       sdk.Int `json:"istanbul_block" yaml:"istanbul_block"`             // Istanbul switch block (< 0 no fork, 0 = already on istanbul)
	MuirGlacierBlock    sdk.Int `json:"muir_glacier_block" yaml:"muir_glacier_block"`     // Eip-2384 (bomb delay) switch block (< 0 no fork, 0 = already activated)

	YoloV1Block sdk.Int `json:"yoloV1_block" yaml:"yoloV1_block"` // YOLO v1: https://github.com/ethereum/EIPs/pull/2657 (Ephemeral testnet)
	EWASMBlock  sdk.Int `json:"ewasm_block" yaml:"ewasm_block"`   // EWASM switch block (< 0 no fork, 0 = already activated)
}

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
		YoloV1Block:         getBlockValue(cc.YoloV1Block),
		EWASMBlock:          getBlockValue(cc.EWASMBlock),
	}
}

// String implements the fmt.Stringer interface
func (cc ChainConfig) String() string {
	out, _ := yaml.Marshal(cc)
	return string(out)
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
		YoloV1Block:         sdk.NewInt(-1),
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
	if err := validateBlock(cc.YoloV1Block); err != nil {
		return sdkerrors.Wrap(err, "yoloV1Block")
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

	bz := common.FromHex(hex)
	lenHex := len(bz)
	if lenHex > 0 && lenHex != common.HashLength {
		return sdkerrors.Wrapf(ErrInvalidChainConfig, "invalid hash length, expected %d, got %d", common.HashLength, lenHex)
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
