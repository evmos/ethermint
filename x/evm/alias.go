package evm

import (
	"github.com/cosmos/ethermint/x/evm/keeper"
	"github.com/cosmos/ethermint/x/evm/types"
)

// nolint
const (
	ModuleName           = types.ModuleName
	StoreKey             = types.StoreKey
	RouterKey            = types.RouterKey
	QueryProtocolVersion = types.QueryProtocolVersion
	QueryBalance         = types.QueryBalance
	QueryBlockNumber     = types.QueryBlockNumber
	QueryStorage         = types.QueryStorage
	QueryCode            = types.QueryCode
	QueryNonce           = types.QueryNonce
	QueryHashToHeight    = types.QueryHashToHeight
	QueryTransactionLogs = types.QueryTransactionLogs
	QueryBloom           = types.QueryBloom
	QueryLogs            = types.QueryLogs
	QueryAccount         = types.QueryAccount
	QueryExportAccount   = types.QueryExportAccount
)

// nolint
var (
	NewKeeper         = keeper.NewKeeper
	TxDecoder         = types.TxDecoder
	NewGenesisStorage = types.NewGenesisStorage
)

//nolint
type (
	Keeper          = keeper.Keeper
	QueryResAccount = types.QueryResAccount
	GenesisState    = types.GenesisState
	GenesisAccount  = types.GenesisAccount
	GenesisStorage  = types.GenesisStorage
)
