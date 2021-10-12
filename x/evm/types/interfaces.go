package types

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	feemarkettypes "github.com/tharsis/ethermint/x/feemarket/types"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// AccountKeeper defines the expected account keeper interface
type AccountKeeper interface {
	NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
	GetModuleAddress(moduleName string) sdk.AccAddress
	GetAllAccounts(ctx sdk.Context) (accounts []authtypes.AccountI)
	IterateAccounts(ctx sdk.Context, cb func(account authtypes.AccountI) bool)
	GetSequence(sdk.Context, sdk.AccAddress) (uint64, error)
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
	SetAccount(ctx sdk.Context, account authtypes.AccountI)
	RemoveAccount(ctx sdk.Context, account authtypes.AccountI)
	GetParams(ctx sdk.Context) (params authtypes.Params)
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	// SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
}

// StakingKeeper returns the historical headers kept in store.
type StakingKeeper interface {
	GetHistoricalInfo(ctx sdk.Context, height int64) (stakingtypes.HistoricalInfo, bool)
	GetValidatorByConsAddr(ctx sdk.Context, consAddr sdk.ConsAddress) (validator stakingtypes.Validator, found bool)
}

// FeeMarketKeeper
type FeeMarketKeeper interface {
	GetBaseFee(ctx sdk.Context) *big.Int
	GetParams(ctx sdk.Context) feemarkettypes.Params
}

// Event Hooks
// These can be utilized to customize evm transaction processing.

// EvmHooks event hooks for evm tx processing
type EvmHooks interface {
	// Must be called after tx is processed successfully, if return an error, the whole transaction is reverted.
	PostTxProcessing(ctx sdk.Context, txHash common.Hash, logs []*ethtypes.Log) error
}

// StateDBKeeper is the keeper interface underlying StateDB implementation
// Differences from `vm.StateDB`:
// - Pass `sdk.Context` explicitly.
// - No following methods:
//   - `GetCommittedState`, it's just `GetState` with a different `sdk.Context`.
//   - `Snapshot`/`RevertToSnapshot`, it's implemented in the `StateDB`.
// - Some methods return `error`, which is set to the `stateErr` in `StateDB`.
type StateDBKeeper interface {
	CreateAccount(sdk.Context, common.Address)

	SubBalance(sdk.Context, common.Address, *big.Int) error
	AddBalance(sdk.Context, common.Address, *big.Int) error
	GetBalance(sdk.Context, common.Address) *big.Int

	GetNonce(sdk.Context, common.Address) uint64
	SetNonce(sdk.Context, common.Address, uint64) error

	GetCodeHash(sdk.Context, common.Address) common.Hash
	GetCode(sdk.Context, common.Address) []byte
	SetCode(sdk.Context, common.Address, []byte) error
	GetCodeSize(sdk.Context, common.Address) int

	AddRefund(sdk.Context, uint64)
	SubRefund(sdk.Context, uint64)
	GetRefund(sdk.Context) uint64

	GetState(sdk.Context, common.Address, common.Hash) common.Hash
	SetState(sdk.Context, common.Address, common.Hash, common.Hash)

	Suicide(sdk.Context, common.Address) (bool, error)
	HasSuicided(sdk.Context, common.Address) bool

	// Exist reports whether the given account exists in state.
	// Notably this should also return true for suicided accounts.
	Exist(sdk.Context, common.Address) bool
	// Empty returns whether the given account is empty. Empty
	// is defined according to EIP161 (balance = nonce = code = 0).
	Empty(sdk.Context, common.Address) bool

	PrepareAccessList(ctx sdk.Context, sender common.Address, dest *common.Address, precompiles []common.Address, txAccesses ethtypes.AccessList)
	AddressInAccessList(ctx sdk.Context, addr common.Address) bool
	SlotInAccessList(ctx sdk.Context, addr common.Address, slot common.Hash) (addressOk bool, slotOk bool)
	// AddAddressToAccessList adds the given address to the access list. This operation is safe to perform
	// even if the feature/fork is not active yet
	AddAddressToAccessList(ctx sdk.Context, addr common.Address)
	// AddSlotToAccessList adds the given (address,slot) to the access list. This operation is safe to perform
	// even if the feature/fork is not active yet
	AddSlotToAccessList(ctx sdk.Context, addr common.Address, slot common.Hash)

	AddLog(sdk.Context, *ethtypes.Log)

	ForEachStorage(sdk.Context, common.Address, func(common.Hash, common.Hash) bool) error
}
