package app

import (

	// unnamed import of statik for swagger UI support
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	tmtypes "github.com/tendermint/tendermint/types"
	ethermint "github.com/tharsis/ethermint/types"

	// Force-load the tracer engines to trigger registration due to Go-Ethereum v1.10.15 changes
	_ "github.com/ethereum/go-ethereum/eth/tracers/js"
	_ "github.com/ethereum/go-ethereum/eth/tracers/native"
)

func (app *EthermintApp) GetHashFn() func(ctx sdk.Context) vm.GetHashFunc {
	return func(ctx sdk.Context) vm.GetHashFunc {
		return func(height uint64) common.Hash {
			h, err := ethermint.SafeInt64(height)
			if err != nil {
				return common.Hash{}
			}

			switch {
			case ctx.BlockHeight() == h:
				// Case 1: The requested height matches the one from the context so we can retrieve the header
				// hash directly from the context.
				// Note: The headerHash is only set at begin block, it will be nil in case of a query context
				headerHash := ctx.HeaderHash()
				if len(headerHash) != 0 {
					return common.BytesToHash(headerHash)
				}

				// only recompute the hash if not set (eg: checkTxState)
				contextBlockHeader := ctx.BlockHeader()
				header, err := tmtypes.HeaderFromProto(&contextBlockHeader)
				if err != nil {
					app.Logger().Error("failed to cast tendermint header from proto", "error", err)
					return common.Hash{}
				}

				headerHash = header.Hash()
				return common.BytesToHash(headerHash)

			case ctx.BlockHeight() > h:
				// Case 2: if the chain is not the current height we need to retrieve the hash from the store for the
				// current chain epoch. This only applies if the current height is greater than the requested height.
				histInfo, found := app.StakingKeeper.GetHistoricalInfo(ctx, h)
				if !found {
					app.Logger().Debug("historical info not found", "height", h)
					return common.Hash{}
				}

				header, err := tmtypes.HeaderFromProto(&histInfo.Header)
				if err != nil {
					app.Logger().Error("failed to cast tendermint header from proto", "error", err)
					return common.Hash{}
				}

				return common.BytesToHash(header.Hash())
			default:
				// Case 3: heights greater than the current one returns an empty hash.
				return common.Hash{}
			}
		}
	}
}

func (app *EthermintApp) GetValidatorOperatorByConsAddr() func(sdk.Context, sdk.ConsAddress) (sdk.AccAddress, bool) {
	return func(ctx sdk.Context, consAddr sdk.ConsAddress) (sdk.AccAddress, bool) {
		validator, found := app.StakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
		return sdk.AccAddress(validator.GetOperator()), found
	}
}
