package ethermint

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	// transferkeeper "github.com/cosmos/ibc-go/v5/modules/apps/transfer/keeper"
	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"

	"github.com/evmos/ethermint/x/evm/statedb"
	"github.com/evmos/ethermint/x/evm/types"

	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
)

var _ statedb.JournalEntry = ics20TransferChange{}

type ics20TransferChange struct {
	packet     channeltypes.Packet
	bankKeeper types.BankKeeper
}

func (tc ics20TransferChange) Revert(stateDB *statedb.StateDB) {
	var data transfertypes.FungibleTokenPacketData

	if err := types.ModuleCdc.UnmarshalJSON(tc.packet.GetData(), &data); err != nil {
		err = sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "cannot unmarshal ICS-20 transfer packet data")
		panic(err)
	}

	if err := tc.refundPacketToken(stateDB.Context(), tc.packet, data); err != nil {
		panic(err)
	}
}

// Dirtied returns nil since the ethereum account information from the state object
// has not been modified.
func (tc ics20TransferChange) Dirtied() *common.Address {
	return nil
}

// TODO: use transfer keeper instead
// See: https://github.com/cosmos/ibc-go/pull/2660
func (tc ics20TransferChange) refundPacketToken(ctx sdk.Context, packet channeltypes.Packet, data transfertypes.FungibleTokenPacketData) error {
	// NOTE: packet data type already checked in handler.go

	// parse the denomination from the full denom path
	trace := transfertypes.ParseDenomTrace(data.Denom)

	// parse the transfer amount
	transferAmount, ok := sdk.NewIntFromString(data.Amount)
	if !ok {
		return sdkerrors.Wrapf(transfertypes.ErrInvalidAmount, "unable to parse transfer amount (%s) into math.Int", data.Amount)
	}
	token := sdk.NewCoin(trace.IBCDenom(), transferAmount)

	// decode the sender address
	sender, err := sdk.AccAddressFromBech32(data.Sender)
	if err != nil {
		return err
	}

	if transfertypes.SenderChainIsSource(packet.GetSourcePort(), packet.GetSourceChannel(), data.Denom) {
		// unescrow tokens back to sender
		escrowAddress := transfertypes.GetEscrowAddress(packet.GetSourcePort(), packet.GetSourceChannel())
		if err := tc.bankKeeper.SendCoins(ctx, escrowAddress, sender, sdk.NewCoins(token)); err != nil {
			// NOTE: this error is only expected to occur given an unexpected bug or a malicious
			// counterparty module. The bug may occur in bank or any part of the code that allows
			// the escrow address to be drained. A malicious counterparty module could drain the
			// escrow address by allowing more tokens to be sent back then were escrowed.
			return sdkerrors.Wrap(
				err,
				"unable to unescrow tokens, this may be caused by a malicious counterparty module or a bug: please open an issue on counterparty module",
			)
		}

		return nil
	}

	// mint vouchers back to sender
	if err := tc.bankKeeper.MintCoins(
		ctx, types.ModuleName, sdk.NewCoins(token),
	); err != nil {
		return err
	}

	if err := tc.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sender, sdk.NewCoins(token)); err != nil {
		panic(fmt.Sprintf("unable to send coins from module to account despite previously minting coins to module account: %v", err))
	}

	return nil
}
