package evm

import (
	"github.com/cosmos/ethermint/x/evm/keeper"
	"github.com/cosmos/ethermint/x/evm/types"
)

// nolint
const (
	ModuleName           = types.ModuleName
	StoreKey             = types.StoreKey
	CodeKey              = types.StoreKey
	BlockKey             = types.BlockKey
	RouterKey            = types.RouterKey
	QueryProtocolVersion = types.QueryProtocolVersion
	QueryBalance         = types.QueryBalance
	QueryBlockNumber     = types.QueryBlockNumber
	QueryStorage         = types.QueryStorage
	QueryCode            = types.QueryCode
	QueryNonce           = types.QueryNonce
	QueryHashToHeight    = types.QueryHashToHeight
	QueryTransactionLogs = types.QueryTransactionLogs
	QueryLogsBloom       = types.QueryLogsBloom
	QueryLogs            = types.QueryLogs
	QueryAccount         = types.QueryAccount
)

// nolint
var (
	NewKeeper = keeper.NewKeeper
	TxDecoder = types.TxDecoder
)

//nolint
type (
	Keeper          = keeper.Keeper
	QueryResAccount = types.QueryResAccount
	GenesisState    = types.GenesisState
)
