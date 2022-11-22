package app

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	simapputil "github.com/cosmos/cosmos-sdk/testutil/sims"

	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

func TestEthermintAppExport(t *testing.T) {
	db := dbm.NewMemDB()
	app := SetupWithDB(false, nil, db)
	app.Commit()

	// Making a new app object with the db, so that initchain hasn't been called
	app2 := NewEthermintApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil, true, map[int64]bool{}, DefaultNodeHome, 0, simapputil.EmptyAppOptions{})
	_, err := app2.ExportAppStateAndValidators(false, []string{}, []string{})
	require.NoError(t, err, "ExportAppStateAndValidators should not have an error")
}
