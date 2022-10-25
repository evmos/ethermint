package ethermint

import (
	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/ethermint/x/evm/statedb"
)

var _ statedb.JournalEntry = ics20TransferChange{}

type balanceCoinsChange struct {
	account *common.Address
	prev    sdk.Coins
}

func (ch balanceCoinsChange) Revert(s *statedb.StateDB) {
	// FIXME: we don't have the state object exposed
	// s.getStateObject(*ch.account).setBalance(ch.prev)
}

func (ch balanceCoinsChange) Dirtied() *common.Address {
	return ch.account
}

type ics20TransferChange struct {
	precompile *ICS20Precompile
	caller     common.Address
	sender     *common.Address
	msg        *ics20Transfer
}

func (tc ics20TransferChange) Revert(*statedb.StateDB) {
	// FIXME: this needs to be sequential balance changes
}

func (tc ics20TransferChange) Dirtied() *common.Address {
	return nil
}
