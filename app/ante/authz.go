// Copyright 2021 Evmos Foundation
// This file is part of Evmos' Ethermint library.
//
// The Ethermint library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Ethermint library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Ethermint library. If not, see https://github.com/evmos/ethermint/blob/main/LICENSE
package ante

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

// maxNestedMsgs defines a cap for the number of nested messages on a MsgExec message
const maxNestedMsgs = 7

// AuthzLimiterDecorator blocks certain msg types from being granted or executed
// within the authorization module.
type AuthzLimiterDecorator struct {
	// disabledMsgTypes is the type urls of the msgs to block.
	disabledMsgTypes []string
}

// NewAuthzLimiterDecorator creates a decorator to block certain msg types from being granted or executed within authz.
func NewAuthzLimiterDecorator(disabledMsgTypes ...string) AuthzLimiterDecorator {
	return AuthzLimiterDecorator{
		disabledMsgTypes: disabledMsgTypes,
	}
}

func (ald AuthzLimiterDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	if err := ald.checkDisabledMsgs(tx.GetMsgs(), false, 1); err != nil {
		return ctx, errorsmod.Wrapf(errortypes.ErrUnauthorized, err.Error())
	}
	return next(ctx, tx, simulate)
}

// checkDisabledMsgs iterates through the msgs and returns an error if it finds any unauthorized msgs.
//
// When searchOnlyInAuthzMsgs is enabled, only authz MsgGrant and MsgExec are blocked, if they contain unauthorized msg types.
// Otherwise any msg matching the disabled types are blocked, regardless of being in an authz msg or not.
//
// This method is recursive as MsgExec's can wrap other MsgExecs. The check for nested messages is performed up to the
// maxNestedMsgs threshold. If there are more than that limit, it returns an error
func (ald AuthzLimiterDecorator) checkDisabledMsgs(msgs []sdk.Msg, isAuthzInnerMsg bool, nestedLvl int) error {
	if nestedLvl >= maxNestedMsgs {
		return fmt.Errorf("found more nested msgs than permited. Limit is : %d", maxNestedMsgs)
	}
	for _, msg := range msgs {
		switch msg := msg.(type) {
		case *authz.MsgExec:
			innerMsgs, err := msg.GetMessages()
			if err != nil {
				return err
			}
			nestedLvl++
			if err := ald.checkDisabledMsgs(innerMsgs, true, nestedLvl); err != nil {
				return err
			}
		case *authz.MsgGrant:
			authorization, err := msg.GetAuthorization()
			if err != nil {
				return err
			}

			url := authorization.MsgTypeURL()
			if ald.isDisabledMsg(url) {
				return fmt.Errorf("found disabled msg type: %s", url)
			}
		default:
			url := sdk.MsgTypeURL(msg)
			if isAuthzInnerMsg && ald.isDisabledMsg(url) {
				return fmt.Errorf("found disabled msg type: %s", url)
			}
		}
	}
	return nil
}

// isDisabledMsg returns true if the given message is in the list of restricted
// messages from the AnteHandler.
func (ald AuthzLimiterDecorator) isDisabledMsg(msgTypeURL string) bool {
	for _, disabledType := range ald.disabledMsgTypes {
		if msgTypeURL == disabledType {
			return true
		}
	}

	return false
}
