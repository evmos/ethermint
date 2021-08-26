package log

import (
	"io"
	"testing"

	"github.com/cosmos/cosmos-sdk/server"
	ethlog "github.com/ethereum/go-ethereum/log"
	"github.com/rs/zerolog"
	"github.com/tendermint/tendermint/types/time"
)

func BenchmarkHandler_Log(b *testing.B) {
	logger := &server.ZeroLogWrapper{
		Logger: zerolog.New(io.Discard).Level(zerolog.DebugLevel).With().Timestamp().Logger(),
	}
	h := NewHandler(logger)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		h.Log(&ethlog.Record{
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
	}
}
