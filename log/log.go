package log

import (
	tmlog "github.com/tendermint/tendermint/libs/log"

	"github.com/sirupsen/logrus"
)

type EthermintLogger struct {
	TendermintLogger *tmlog.Logger
}

func (f EthermintLogger) Levels() []logrus.Level {
	//return make([]logrus.Level, 0)
	return []logrus.Level{logrus.DebugLevel, logrus.InfoLevel, logrus.TraceLevel, logrus.WarnLevel}
}

func (f EthermintLogger) Fire(entry *logrus.Entry) error {
	// a, _ := (*entry).String()
	a := (*entry).Message
	(*f.TendermintLogger).Info(a)
	return nil
}

var EthermintLoggerInstance EthermintLogger = EthermintLogger{nil}
