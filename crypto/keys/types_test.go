package keys

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/hd"
	"github.com/cosmos/ethermint/crypto"
	"github.com/stretchr/testify/assert"
)

func TestWriteReadInfo(t *testing.T) {
	tmpKey, err := crypto.GenerateKey()
	assert.NoError(t, err)
	pkey := tmpKey.PubKey()
	pubkey, ok := pkey.(crypto.PubKeySecp256k1)
	assert.True(t, ok)

	info := newOfflineInfo("offline", pubkey)
	bytes := writeInfo(info)
	assert.NotNil(t, bytes)

	regeneratedKey, err := readInfo(bytes)
	assert.NoError(t, err)
	assert.Equal(t, info.GetPubKey(), regeneratedKey.GetPubKey())
	assert.Equal(t, info.GetName(), regeneratedKey.GetName())

	info = newLocalInfo("local", pubkey, "testarmor")
	bytes = writeInfo(info)
	assert.NotNil(t, bytes)

	regeneratedKey, err = readInfo(bytes)
	assert.NoError(t, err)
	assert.Equal(t, info.GetPubKey(), regeneratedKey.GetPubKey())
	assert.Equal(t, info.GetName(), regeneratedKey.GetName())

	info = newLedgerInfo("ledger", pubkey,
		hd.BIP44Params{Purpose: 1, CoinType: 1, Account: 1, Change: false, AddressIndex: 1})
	bytes = writeInfo(info)
	assert.NotNil(t, bytes)

	regeneratedKey, err = readInfo(bytes)
	assert.NoError(t, err)
	assert.Equal(t, info.GetPubKey(), regeneratedKey.GetPubKey())
	assert.Equal(t, info.GetName(), regeneratedKey.GetName())

}
