package core

import (
	"math/big"

	ethcommon "github.com/ethereum/go-ethereum/common"
	ethconsensus "github.com/ethereum/go-ethereum/consensus"
	ethstate "github.com/ethereum/go-ethereum/core/state"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethrpc "github.com/ethereum/go-ethereum/rpc"
)

// ChainContext implements Ethereum's core.ChainContext and consensus.Engine
// interfaces. It is needed in order to apply and process Ethereum
// transactions. There should only be a single implementation in Ethermint. For
// the purposes of Ethermint, it should be support retrieving headers and
// consensus parameters from  the current blockchain to be used during
// transaction processing.
//
// NOTE: Ethermint will distribute the fees out to validators, so the structure
// and functionality of this is a WIP and subject to change.
type ChainContext struct {
	coinbase ethcommon.Address
}

// Engine implements Ethereum's core.ChainContext interface. As a ChainContext
// implements the consensus.Engine interface, it is simply returned.
func (cc *ChainContext) Engine() ethconsensus.Engine {
	return cc
}

// GetHeader implements Ethereum's core.ChainContext interface. It currently
// performs a no-op.
//
// TODO: The Cosmos SDK supports retreiving such information in contexts and
// multi-store, so this will be need to be integrated.
func (cc *ChainContext) GetHeader(ethcommon.Hash, uint64) *ethtypes.Header {
	return nil
}

// Author implements Ethereum's consensus.Engine interface. It is responsible
// for returned the address of the validtor to receive any fees. This function
// is only invoked if the given author in the ApplyTransaction call is nil.
//
// NOTE: Ethermint will distribute the fees out to validators, so the structure
// and functionality of this is a WIP and subject to change.
func (cc *ChainContext) Author(_ *ethtypes.Header) (ethcommon.Address, error) {
	return cc.coinbase, nil
}

// APIs implements Ethereum's core.ChainContext interface. It currently
// performs a no-op.
//
// TODO: Do we need to support such RPC APIs? This will tie into a bigger
// discussion on if we want to support web3.
func (cc *ChainContext) APIs(_ ethconsensus.ChainReader) []ethrpc.API {
	return nil
}

// CalcDifficulty implements Ethereum's core.ChainContext interface. It
// currently performs a no-op.
func (cc *ChainContext) CalcDifficulty(_ ethconsensus.ChainReader, _ uint64, _ *ethtypes.Header) *big.Int {
	return nil
}

// Finalize implements Ethereum's core.ChainContext interface. It currently
// performs a no-op.
//
// TODO: Figure out if this needs to be hooked up to any part of the ABCI?
func (cc *ChainContext) Finalize(
	_ ethconsensus.ChainReader, _ *ethtypes.Header, _ *ethstate.StateDB,
	_ []*ethtypes.Transaction, _ []*ethtypes.Header, _ []*ethtypes.Receipt,
) (*ethtypes.Block, error) {
	return nil, nil
}

// Prepare implements Ethereum's core.ChainContext interface. It currently
// performs a no-op.
//
// TODO: Figure out if this needs to be hooked up to any part of the ABCI?
func (cc *ChainContext) Prepare(_ ethconsensus.ChainReader, _ *ethtypes.Header) error {
	return nil
}

// Seal implements Ethereum's core.ChainContext interface. It currently
// performs a no-op.
//
// TODO: Figure out if this needs to be hooked up to any part of the ABCI?
func (cc *ChainContext) Seal(_ ethconsensus.ChainReader, _ *ethtypes.Block, _ <-chan struct{}) (*ethtypes.Block, error) {
	return nil, nil
}

// VerifyHeader implements Ethereum's core.ChainContext interface. It currently
// performs a no-op.
//
// TODO: Figure out if this needs to be hooked up to any part of the Cosmos SDK
// handlers?
func (cc *ChainContext) VerifyHeader(_ ethconsensus.ChainReader, _ *ethtypes.Header, _ bool) error {
	return nil
}

// VerifyHeaders implements Ethereum's core.ChainContext interface. It
// currently performs a no-op.
//
// TODO: Figure out if this needs to be hooked up to any part of the Cosmos SDK
// handlers?
func (cc *ChainContext) VerifyHeaders(_ ethconsensus.ChainReader, _ []*ethtypes.Header, _ []bool) (chan<- struct{}, <-chan error) {
	return nil, nil
}

// VerifySeal implements Ethereum's core.ChainContext interface. It currently
// performs a no-op.
//
// TODO: Figure out if this needs to be hooked up to any part of the Cosmos SDK
// handlers?
func (cc *ChainContext) VerifySeal(_ ethconsensus.ChainReader, _ *ethtypes.Header) error {
	return nil
}

// VerifyUncles implements Ethereum's core.ChainContext interface. It currently
// performs a no-op.
func (cc *ChainContext) VerifyUncles(_ ethconsensus.ChainReader, _ *ethtypes.Block) error {
	return nil
}
