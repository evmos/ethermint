package log

import (
	"fmt"

	ethlog "github.com/ethereum/go-ethereum/log"
	tmlog "github.com/tendermint/tendermint/libs/log"
)

var ethermintLogger *tmlog.Logger = nil

func FuncHandler(fn func(r *ethlog.Record) error) ethlog.Handler {
	return funcHandler(fn)
}

type funcHandler func(r *ethlog.Record) error

func (h funcHandler) Log(r *ethlog.Record) error {
	return h(r)
}

func NewHandler(logger tmlog.Logger) ethlog.Handler {

	ethermintLogger = &logger

	return FuncHandler(func(r *ethlog.Record) error {
		(*ethermintLogger).Debug(fmt.Sprintf("[EVM] %v", r))
		return nil
	})
}
