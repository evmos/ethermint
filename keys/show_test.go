package keys

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/tests"
)

func TestShowKeysCmd(t *testing.T) {
	cmd := showKeysCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "false", cmd.Flag(FlagAddress).DefValue)
	assert.Equal(t, "false", cmd.Flag(FlagPublicKey).DefValue)
}

func TestRunShowCmd(t *testing.T) {
	cmd := showKeysCmd()

	err := runShowCmd(cmd, []string{"invalid"})
	assert.EqualError(t, err, "Key invalid not found")

	// Prepare a key base
	// Now add a temporary keybase
	kbHome, cleanUp := tests.NewTestCaseDir(t)
	defer cleanUp()
	viper.Set(flags.FlagHome, kbHome)

	fakeKeyName1 := "runShowCmd_Key1"
	fakeKeyName2 := "runShowCmd_Key2"
	kb, err := NewKeyBaseFromHomeFlag()
	assert.NoError(t, err)
	_, err = kb.CreateAccount(fakeKeyName1, tests.TestMnemonic, "", "", 0, 0)
	assert.NoError(t, err)
	_, err = kb.CreateAccount(fakeKeyName2, tests.TestMnemonic, "", "", 0, 1)
	assert.NoError(t, err)

	// // Now try single key
	// err = runShowCmd(cmd, []string{fakeKeyName1})
	// assert.EqualError(t, err, "invalid Bech32 prefix encoding provided: ")

	// // Now try single key - set bech to acc
	// viper.Set(FlagBechPrefix, sdk.PrefixAccount)
	err = runShowCmd(cmd, []string{fakeKeyName1})
	assert.NoError(t, err)
	err = runShowCmd(cmd, []string{fakeKeyName2})
	assert.NoError(t, err)
}
