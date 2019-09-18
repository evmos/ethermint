package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// ** Modified version of github.com/cosmos/cosmos-sdk/x/auth/types/account_retriever.go
// ** to allow passing in a codec for decoding Account types
// AccountRetriever defines the properties of a type that can be used to
// retrieve accounts.
type AccountRetriever struct {
	querier auth.NodeQuerier
	codec   *codec.Codec
}

// * Modified to allow a codec to be passed in
// NewAccountRetriever initialises a new AccountRetriever instance.
func NewAccountRetriever(querier auth.NodeQuerier, codec *codec.Codec) AccountRetriever {
	if codec == nil {
		codec = auth.ModuleCdc
	}
	return AccountRetriever{querier: querier, codec: codec}
}

// GetAccount queries for an account given an address and a block height. An
// error is returned if the query or decoding fails.
func (ar AccountRetriever) GetAccount(addr sdk.AccAddress) (exported.Account, error) {
	account, _, err := ar.GetAccountWithHeight(addr)
	return account, err
}

// GetAccountWithHeight queries for an account given an address. Returns the
// height of the query with the account. An error is returned if the query
// or decoding fails.
func (ar AccountRetriever) GetAccountWithHeight(addr sdk.AccAddress) (exported.Account, int64, error) {
	// ** This line was changed to use non-static codec
	bs, err := ar.codec.MarshalJSON(auth.NewQueryAccountParams(addr))
	if err != nil {
		return nil, 0, err
	}

	res, height, err := ar.querier.QueryWithData(fmt.Sprintf("custom/%s/%s", auth.QuerierRoute, auth.QueryAccount), bs)
	if err != nil {
		return nil, 0, err
	}

	var account exported.Account
	// ** This line was changed to use non-static codec
	if err := ar.codec.UnmarshalJSON(res, &account); err != nil {
		return nil, 0, err
	}

	return account, height, nil
}

// EnsureExists returns an error if no account exists for the given address else nil.
func (ar AccountRetriever) EnsureExists(addr sdk.AccAddress) error {
	if _, err := ar.GetAccount(addr); err != nil {
		return err
	}
	return nil
}

// GetAccountNumberSequence returns sequence and account number for the given address.
// It returns an error if the account couldn't be retrieved from the state.
func (ar AccountRetriever) GetAccountNumberSequence(addr sdk.AccAddress) (uint64, uint64, error) {
	acc, err := ar.GetAccount(addr)
	if err != nil {
		return 0, 0, err
	}
	return acc.GetAccountNumber(), acc.GetSequence(), nil
}
