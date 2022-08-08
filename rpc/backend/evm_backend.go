package backend

import (
	"github.com/cosmos/cosmos-sdk/client"
)

// ClientCtx returns client context
func (b *Backend) ClientCtx() client.Context {
	return b.clientCtx
}
