package ethermint

// TODO: add

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	"github.com/evmos/ethermint/x/evm/statedb"
	"github.com/evmos/ethermint/x/evm/types"
	evm "github.com/evmos/ethermint/x/evm/vm"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransferkeeper "github.com/cosmos/ibc-go/v5/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v5/modules/core/02-client/types"
	ibcchannelkeeper "github.com/cosmos/ibc-go/v5/modules/core/04-channel/keeper"
	ibcchanneltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
)

var (
	TransferMethod     abi.Method
	HasCommitMethod    abi.Method
	QueryNextSeqMethod abi.Method

	_ evm.StatefulPrecompiledContract = (*ICS20Precompile)(nil)
)

func init() {
	addressType, _ := abi.NewType("address", "", nil)
	stringType, _ := abi.NewType("string", "", nil)
	uint256Type, _ := abi.NewType("uint256", "", nil)
	boolType, _ := abi.NewType("bool", "", nil)
	TransferMethod = abi.NewMethod(
		"transfer", // name
		"transfer", // raw name
		abi.Function,
		"",
		false,
		false,
		abi.Arguments{
			{
				Name: "portID",
				Type: stringType,
			},
			{
				Name: "channelID",
				Type: stringType,
			},
			{
				Name: "srcDenom",
				Type: stringType,
			},
			{
				Name: "ratio",
				Type: uint256Type,
			},
			{
				Name: "timeout",
				Type: uint256Type,
			},
			{
				Name: "sender",
				Type: addressType,
			},
			{
				Name: "receiver",
				Type: stringType,
			},
			{
				Name: "amount",
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
	HasCommitMethod = abi.NewMethod(
		"hasCommit",
		"hasCommit",
		abi.Function,
		"",
		false,
		false,
		abi.Arguments{
			{
				Name: "portID",
				Type: stringType,
			},
			{
				Name: "channelID",
				Type: stringType,
			},
			{
				Name: "sequence",
				Type: uint256Type,
			},
		},
		abi.Arguments{
			abi.Argument{
				Name: "status",
				Type: boolType,
			},
		},
	)
	QueryNextSeqMethod = abi.NewMethod(
		"queryNextSeq",
		"queryNextSeq",
		abi.Function,
		"",
		false,
		false,
		abi.Arguments{
			{
				Name: "portID",
				Type: stringType,
			},
			{
				Name: "channelID",
				Type: stringType,
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
	channelKeeper  *ibcchannelkeeper.Keeper
	transferKeeper *ibctransferkeeper.Keeper
	bankKeeper     types.BankKeeper
}

func NewICS20Precompile(
	channelKeeper *ibcchannelkeeper.Keeper,
	transferKeeper *ibctransferkeeper.Keeper,
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

func (ic *ICS20Precompile) RunStateful(evm *vm.EVM, caller common.Address, input []byte, value *big.Int) ([]byte, error) {
	stateDB, ok := evm.StateDB.(statedb.ExtStateDB)
	if !ok {
		return nil, errors.New("not run in ethermint")
	}

	ctx := stateDB.Context()

	methodID := string(input[:4])
	switch methodID {

	case string(TransferMethod.ID):

		args, err := TransferMethod.Inputs.Unpack(input[4:])
		if err != nil {
			return nil, errors.New("fail to unpack input arguments")
		}

		portID := args[0].(string)
		channelID := args[1].(string)
		srcDenom := args[2].(string)
		ratio := args[4].(*big.Int)
		timeout := args[5].(*big.Int)
		sender := args[6].(common.Address)
		receiver := args[7].(string)
		amount := args[8].(*big.Int)

		if amount.Sign() <= 0 {
			return nil, errors.New("invalid amount")
		}

		timeoutTimestamp := uint64(ctx.BlockTime().UnixNano()) + timeout.Uint64()
		timeoutHeight := clienttypes.ZeroHeight()

		// Use instance to prevent
		token := sdk.Coin{
			Denom:  srcDenom,
			Amount: sdk.NewIntFromBigInt(amount),
		}

		src := sdk.AccAddress(sender.Bytes())

		transfer := &ibctransfertypes.MsgTransfer{
			SourcePort:       portID,
			SourceChannel:    channelID,
			Token:            token,
			Sender:           src.String(), // convert to bech32 format
			Receiver:         receiver,
			TimeoutHeight:    timeoutHeight,
			TimeoutTimestamp: timeoutTimestamp,
		}

		if err := transfer.ValidateBasic(); err != nil {
			return nil, err
		}

		// TODO: create balance change journal entries for each ICS20 case
		// NOTE: this needs to be done via balance diff since we can't
		// set the balance

		if _, ok := ic.msgs[caller]; !ok {
			ic.msgs[caller] = make(map[common.Address]*ics20Transfer)
		}

		msgs := ic.msgs[caller]
		msg, ok := msgs[sender]
		if ok {
			msg.dirtyAmount = new(big.Int).Sub(msg.dirtyAmount, amount)
		} else {
			// query original amount
			addr := sdk.AccAddress(sender.Bytes())
			originAmount := ic.bankKeeper.GetBalance(ctx, addr, srcDenom).Amount.BigInt()
			dirtyAmount := new(big.Int).Sub(originAmount, amount)
			msg = newICS20Transfer(
				transfer,
				ratio,
				originAmount,
				dirtyAmount,
			)
			msgs[sender] = msg
		}

		stateDB.AppendJournalEntry(ics20TransferChange{ic, caller, &sender, msg})
		sequence, _ := ic.channelKeeper.GetNextSequenceSend(ctx, portID, channelID)

		status := ic.channelKeeper.HasPacketCommitment(ctx, portID, channelID, sequence)

		fmt.Printf("TransferMethod sequence: %d, %+v\n", sequence, status)

		return TransferMethod.Outputs.Pack(new(big.Int).SetUint64(sequence))

	default:
		return nil, fmt.Errorf("unknown method '%s'", methodID)
	}
}

func (ic *ICS20Precompile) Commit(ctx sdk.Context) error {
	// FIXME: use slices to avoid non-determinism
	for _, msgs := range ic.msgs {
		for sender, msg := range msgs {
			acc, err := sdk.AccAddressFromBech32(msg.msg.Sender)
			if err != nil {
				return err
			}

			c := msg.msg.Token
			ratio := sdk.NewIntFromBigInt(msg.ratio)
			amount8decRem := c.Amount.Mod(ratio)
			amountToBurn := c.Amount.Sub(amount8decRem)
			if amountToBurn.IsZero() {
				// Amount too small
				continue
			}

			changed := msgs[sender].Changed()
			coins := sdk.NewCoins(sdk.NewCoin(msg.transfer.Token.Denom, amountToBurn))
			amount8dec := c.Amount.Quo(ratio)

			hash := sha256.Sum256([]byte(fmt.Sprintf("%s/%s/%s", ibctransfertypes.ModuleName, msg.transfer.SourceChannel, msg.dstDenom)))
			ibcDenom := fmt.Sprintf("ibc/%s", strings.ToUpper(hex.EncodeToString(hash[:])))
			ibcCoin := sdk.NewCoin(ibcDenom, amount8dec)

			switch changed.Sign() {
			case -1:
				// Send evm tokens to escrow address
				if err = ic.bankKeeper.SendCoinsFromAccountToModule(
					ctx, acc, ic.module, coins); err != nil {
					return err
				}

				// Burns the evm tokens
				if err := ic.bankKeeper.BurnCoins(
					ctx, ic.module, coins); err != nil {
					return err
				}

				// Transfer ibc tokens back to the user
				if err := ic.bankKeeper.SendCoinsFromModuleToAccount(
					ctx, ic.module, acc, sdk.NewCoins(ibcCoin),
				); err != nil {
					return err
				}

				msg.msg.Token = ibcCoin
				res, err := ic.transferKeeper.Transfer(ctx, msg.transfer)
				if err != nil {
					if ibcchanneltypes.ErrPacketTimeout.Is(err) {
						if err := ic.bankKeeper.MintCoins(
							ctx, ic.module, coins); err != nil {
							return err
						}

						if err := ic.bankKeeper.SendCoinsFromModuleToAccount(
							ctx, ic.module, acc, coins); err != nil {
							return err
						}

						return nil
					}

					fmt.Printf("Transfer res: %+v, %+v\n", res, err)
					return err
				}
			case 1:
				// msg.transfer.Token = ibcCoin
				// res, err := ic.transferKeeper.Transfer(goCtx, msg.transfer)
				if err := ic.bankKeeper.SendCoinsFromAccountToModule(
					ctx, acc, ic.module, sdk.NewCoins(ibcCoin),
				); err != nil {
					return err
				}

				if err := ic.bankKeeper.MintCoins(
					ctx, ic.module, coins); err != nil {
					return err
				}

				if err := ic.bankKeeper.SendCoinsFromModuleToAccount(
					ctx, ic.module, acc, coins); err != nil {
					return err
				}
				return nil
			}
		}
	}
	return nil
}

type ics20Transfer struct {
	msg          *ibctransfertypes.MsgTransfer
	dstDenom     string
	ratio        *big.Int
	originAmount *big.Int
	dirtyAmount  *big.Int
}

func newICS20Transfer(msg *ibctransfertypes.MsgTransfer, ratio, originAmount, dirtyAmount *big.Int) *ics20Transfer {
	return &ics20Transfer{
		msg:          msg,
		ratio:        ratio,
		originAmount: originAmount,
		dirtyAmount:  dirtyAmount,
	}
}
