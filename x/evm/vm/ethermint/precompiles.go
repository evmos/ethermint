package ethermint

// TODO: add

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"github.com/evmos/ethermint/x/evm/statedb"
	"github.com/evmos/ethermint/x/evm/types"
	evm "github.com/evmos/ethermint/x/evm/vm"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transferkeeper "github.com/cosmos/ibc-go/v5/modules/apps/transfer/keeper"
	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v5/modules/core/02-client/types"
	channelkeeper "github.com/cosmos/ibc-go/v5/modules/core/04-channel/keeper"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
)

var (
	TransferMethod abi.Method

	_ evm.StatefulPrecompiledContract = (*ICS20Precompile)(nil)
)

func init() {
	addressType, _ := abi.NewType("address", "", nil)
	stringType, _ := abi.NewType("string", "", nil)
	uint256Type, _ := abi.NewType("uint256", "", nil)

	// TransferMethod defines the ABI method signature for the ICS20 `transfer` function.
	// It rep
	TransferMethod = abi.NewMethod(
		"transfer", // name
		"transfer", // raw name
		abi.Function,
		"",
		false,
		false,
		abi.Arguments{
			{
				Name: "denom",
				Type: stringType,
			},
			{
				Name: "amount",
				Type: uint256Type,
			},
			{
				Name: "sender",
				// NOTE: sender is always local so we use common.Address instead of the string type
				Type: addressType,
			},
			{
				Name: "receiver",
				Type: stringType,
			},
			{
				Name: "sourcePort",
				Type: stringType,
			},
			{
				Name: "sourceChannel",
				Type: stringType,
			},
			{
				Name: "timeoutHeight",
				Type: uint256Type,
			},
			{
				Name: "timeoutTimestamp",
				Type: uint256Type,
			},
		},
		abi.Arguments{
			{
				Name: "sequence",
				Type: uint256Type,
			},
		},
	)
}

type ICS20Precompile struct {
	channelKeeper  *channelkeeper.Keeper
	transferKeeper *transferkeeper.Keeper
	bankKeeper     types.BankKeeper
}

func NewICS20Precompile(
	channelKeeper *channelkeeper.Keeper,
	transferKeeper *transferkeeper.Keeper,
	bankKeeper types.BankKeeper,
) evm.StatefulPrecompiledContract {
	return &ICS20Precompile{
		channelKeeper:  channelKeeper,
		transferKeeper: transferKeeper,
		bankKeeper:     bankKeeper,
	}
}

// RequiredGas calculates the contract gas use
func (ic *ICS20Precompile) RequiredGas(input []byte) uint64 {
	// TODO estimate required gas
	return 0
}

func (ic *ICS20Precompile) Run(_ []byte) ([]byte, error) {
	return nil, errors.New("should run with RunStateful")
}

func (ic *ICS20Precompile) RunStateful(evm evm.EVM, caller common.Address, input []byte, value *big.Int) ([]byte, error) {
	stateDB, ok := evm.StateDB().(statedb.ExtStateDB)
	if !ok {
		return nil, errors.New("not run in ethermint")
	}

	ctx := stateDB.Context()

	methodID := string(input[:4])
	argsBz := input[4:]

	switch methodID {
	case string(TransferMethod.ID):
		return ic.Transfer(ctx, argsBz, stateDB)
	default:
		return nil, fmt.Errorf("unknown method '%s'", methodID)
	}
}

func (ic *ICS20Precompile) Transfer(ctx sdk.Context, argsBz []byte, stateDB statedb.ExtStateDB) ([]byte, error) {
	args, err := TransferMethod.Inputs.Unpack(argsBz)
	if err != nil {
		return nil, errors.New("fail to unpack input arguments")
	}

	msg, err := ic.checkArgs(args, ctx.BlockTime())
	if err != nil {
		return nil, err
	}

	packet, err := ic.buildPacket(ctx,
		msg.SourcePort, msg.SourceChannel,
		msg.Token,
		msg.Sender, msg.Receiver,
		msg.TimeoutHeight, msg.TimeoutTimestamp,
	)
	if err != nil {
		return nil, err
	}

	cacheCtx, writeFn := ctx.CacheContext()

	if _, err := ic.transferKeeper.Transfer(sdk.WrapSDKContext(cacheCtx), msg); err != nil {
		return nil, err
	}

	writeFn()

	// Add state changes to the journal so we can revert in case of error
	entry := ics20TransferChange{
		packet:     packet,
		bankKeeper: ic.bankKeeper,
	}

	stateDB.AppendJournalEntry(entry)

	return TransferMethod.Outputs.Pack(new(big.Int).SetUint64(packet.Sequence))
}

func (ic ICS20Precompile) buildPacket(
	ctx sdk.Context,
	sourcePort,
	sourceChannel string,
	token sdk.Coin,
	sender string,
	receiver string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
) (channeltypes.Packet, error) {
	sequence, found := ic.channelKeeper.GetNextSequenceSend(ctx, sourcePort, sourceChannel)
	if !found {
		return channeltypes.Packet{}, sdkerrors.Wrapf(
			channeltypes.ErrSequenceSendNotFound, "source port %s, source channel",
			sourcePort, sourceChannel,
		)
	}

	// NOTE: denomination and hex hash correctness checked during msg.ValidateBasic
	fullDenomPath := token.Denom
	var err error

	// deconstruct the token denomination into the denomination trace info
	// to determine if the sender is the source chain
	if strings.HasPrefix(token.Denom, "ibc/") {
		fullDenomPath, err = ic.transferKeeper.DenomPathFromHash(ctx, token.Denom)
		if err != nil {
			return channeltypes.Packet{}, err
		}
	}

	sourceChannelEnd, found := ic.channelKeeper.GetChannel(ctx, sourcePort, sourceChannel)
	if !found {
		return channeltypes.Packet{}, sdkerrors.Wrapf(channeltypes.ErrChannelNotFound, "port ID (%s) channel ID (%s)", sourcePort, sourceChannel)
	}

	destinationPort := sourceChannelEnd.GetCounterparty().GetPortID()
	destinationChannel := sourceChannelEnd.GetCounterparty().GetChannelID()

	packetData := transfertypes.NewFungibleTokenPacketData(
		fullDenomPath, token.Amount.String(), sender, receiver,
	)

	packet := channeltypes.NewPacket(
		packetData.GetBytes(),
		sequence,
		sourcePort,
		sourceChannel,
		destinationPort,
		destinationChannel,
		timeoutHeight,
		timeoutTimestamp,
	)

	return packet, nil
}

func (ic *ICS20Precompile) checkArgs(args []interface{}, blockTime time.Time) (*transfertypes.MsgTransfer, error) {
	if len(args) != 8 {
		return nil, fmt.Errorf("invalid input arguments. Expected 8, got %d", len(args))
	}

	sourcePort, _ := args[0].(string)
	sourceChannel, _ := args[1].(string)
	denom, _ := args[2].(string)

	amount, ok := args[3].(*big.Int)
	if !ok || amount == nil {
		amount = big.NewInt(0)
	}

	sender, _ := args[4].(common.Address)
	receiver, ok := args[5].(string)

	timeoutHeightRevisionNumber, ok := args[6].(*big.Int)
	if !ok || timeoutHeightRevisionNumber == nil {
		timeoutHeightRevisionNumber = big.NewInt(0)
	}

	timeoutHeightRevisionHeight, ok := args[7].(*big.Int)
	if !ok || timeoutHeightRevisionHeight == nil {
		timeoutHeightRevisionHeight = big.NewInt(0)
	}

	timeoutDuration, ok := args[8].(*big.Int)
	if !ok || timeoutDuration == nil {
		timeoutDuration = big.NewInt(0)
	}

	timeoutTimestamp := uint64(blockTime.UnixNano()) + timeoutDuration.Uint64()
	timeoutHeight := clienttypes.Height{
		RevisionNumber: timeoutHeightRevisionNumber.Uint64(),
		RevisionHeight: timeoutHeightRevisionHeight.Uint64(),
	}

	// Use instance to prevent errors on denom or
	token := sdk.Coin{
		Denom:  denom,
		Amount: sdk.NewIntFromBigInt(amount),
	}

	src := sdk.AccAddress(sender.Bytes())

	msg := &transfertypes.MsgTransfer{
		SourcePort:       sourcePort,
		SourceChannel:    sourceChannel,
		Token:            token,
		Sender:           src.String(), // convert to bech32 format
		Receiver:         receiver,
		TimeoutHeight:    timeoutHeight,
		TimeoutTimestamp: timeoutTimestamp,
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	return msg, nil
}
