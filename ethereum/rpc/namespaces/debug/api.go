package debug

import (
	"errors"
)

type DebugAPI struct {
}

func NewDebugAPI() *DebugAPI {

	return &DebugAPI{}
}

func (a *DebugAPI) DumpBlock() error {
	return errors.New("Currently not supported.")
}

func (a *DebugAPI) GetBlockRlp() error {
	return errors.New("Currently not supported.")
}

func (a *DebugAPI) SeedHash() error {
	return errors.New("Currently not supported.")
}

func (a *DebugAPI) SetHead() error {
	return errors.New("Currently not supported.")
}

func (a *DebugAPI) TraceBlock() error {
	return errors.New("Currently not supported.")
}

func (a *DebugAPI) TraceBlockByNumber() error {
	return errors.New("Currently not supported.")
}

func (a *DebugAPI) TraceBlockByHash() error {
	return errors.New("Currently not supported.")
}

func (a *DebugAPI) TraceBlockFromFile() error {
	return errors.New("Currently not supported.")
}

func (a *DebugAPI) StandardTraceBlockToFile() error {
	return errors.New("Currently not supported.")
}

func (a *DebugAPI) StandardTraceBadBlockToFile() error {
	return errors.New("Currently not supported.")
}

func (a *DebugAPI) TraceTransaction(hashHex string) error {
	return errors.New("Currently not supported.")
}

func (a *DebugAPI) TraceCall() error {
	return errors.New("Currently not supported.")
}
