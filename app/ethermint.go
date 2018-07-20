package app

import (
	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/wire"
)

const (
	appName = "Ethermint"
)

// EthermintApp implements an extended ABCI application.
type EthermintApp struct {
	*bam.BaseApp

	codec  *wire.Codec
	sealed bool

	// TODO: stores and keys

	// TODO: keepers

	// TODO: mappers
}

// NewEthermintApp returns a reference to a new initialized Ethermint
// application.
func NewEthermintApp(opts ...func(*EthermintApp)) *EthermintApp {
	app := &EthermintApp{}

	// TODO: implement constructor

	for _, opt := range opts {
		opt(app)
	}

	app.seal()
	return app
}

// seal seals the Ethermint application and prohibits any future modifications
// that change critical components.
func (app *EthermintApp) seal() {
	app.sealed = true
}
