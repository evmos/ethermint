package log

import (
	"bytes"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/types/time"

	"github.com/cosmos/cosmos-sdk/server"

	ethlog "github.com/ethereum/go-ethereum/log"
)

const (
	timeKey = "t"
	lvlKey  = "lvl"
	msgKey  = "msg"
	ctxKey  = "ctx"
)

func TestLog(t *testing.T) {
	out := &bytes.Buffer{}

	logger := &server.ZeroLogWrapper{
		Logger: zerolog.New(out).Level(zerolog.DebugLevel).With().Timestamp().Logger(),
	}

	h := NewHandler(logger)

	err := h.Log(&ethlog.Record{
		Time: time.Now().UTC(),
		Lvl:  ethlog.LvlCrit,
		Msg:  "critical error",
		KeyNames: ethlog.RecordKeyNames{
			Time: timeKey,
			Msg:  msgKey,
			Lvl:  lvlKey,
			Ctx:  ctxKey,
		},
	})

	require.NoError(t, err)
	require.Contains(t, string(out.Bytes()), "\"message\":\"critical error\"")
	require.Contains(t, string(out.Bytes()), "\"level\":\"fatal\"")
}

func TestOverrideRootLogger(t *testing.T) {
	out := &bytes.Buffer{}

	logger := &server.ZeroLogWrapper{
		Logger: zerolog.New(out).Level(zerolog.DebugLevel).With().Timestamp().Logger(),
	}

	h := NewHandler(logger)
	ethlog.Root().SetHandler(h)

	ethlog.Root().Info("some info")
	require.Contains(t, string(out.Bytes()), "\"message\":\"some info\"")
	require.Contains(t, string(out.Bytes()), "\"level\":\"info\"")
}
