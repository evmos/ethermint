package log

import (
	"github.com/rs/zerolog"

	"github.com/cosmos/cosmos-sdk/server"

	ethlog "github.com/ethereum/go-ethereum/log"
	tmlog "github.com/tendermint/tendermint/libs/log"
)

var _ ethlog.Handler = &Handler{}

// Logger wraps the zero log Wrapper and extends it to support the ethereum logger interface.
type Handler struct {
	*server.ZeroLogWrapper
}

func NewHandler(logger tmlog.Logger) ethlog.Handler {
	zerologger, ok := logger.(*server.ZeroLogWrapper)
	if !ok {
		// default to Stdout if not an SDK logger wrapper
		return ethlog.StdoutHandler
	}

	return &Handler{
		ZeroLogWrapper: zerologger,
	}
}

// Log implements the go-ethereum Logger Handler interface
func (h *Handler) Log(r *ethlog.Record) error {
	lvl := EthLogLvlToZerolog(r.Lvl)

	h.WithLevel(lvl).
		Fields(getLogFields(r.Ctx...)).
		Time(r.KeyNames.Time, r.Time).
		Msg(r.Msg)
	return nil
}

func EthLogLvlToZerolog(lvl ethlog.Lvl) zerolog.Level {
	var level zerolog.Level

	switch lvl {
	case ethlog.LvlCrit:
		level = zerolog.FatalLevel
	case ethlog.LvlDebug:
		level = zerolog.DebugLevel
	case ethlog.LvlError:
		level = zerolog.ErrorLevel
	case ethlog.LvlInfo:
		level = zerolog.InfoLevel
	case ethlog.LvlTrace:
		level = zerolog.TraceLevel
	case ethlog.LvlWarn:
		level = zerolog.WarnLevel
	default:
		level = zerolog.NoLevel
	}

	return level
}

func getLogFields(keyVals ...interface{}) map[string]interface{} {
	if len(keyVals)%2 != 0 {
		return nil
	}

	fields := make(map[string]interface{})
	for i := 0; i < len(keyVals); i += 2 {
		fields[keyVals[i].(string)] = keyVals[i+1]
	}

	return fields
}

// var ethermintLogger *tmlog.Logger = nil

// func NewHandler(logger tmlog.Logger) ethlog.Handler {

// 	ethermintLogger = &logger

// 	return ethlog.FuncHandler(func(r *ethlog.Record) error {
// 		(*ethermintLogger).Debug(fmt.Sprintf("[EVM] %v", r))
// 		return nil
// 	})
// }
