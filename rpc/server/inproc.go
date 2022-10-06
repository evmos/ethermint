// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package server

import (
	"context"
	"net"

	"github.com/ethereum/go-ethereum/log"
	rpcClient "github.com/evmos/ethermint/rpc/client"
	"github.com/evmos/ethermint/rpc/codec"
)

// DialInProc attaches an in-process connection to the given RPC server.
func DialInProc(handler *Server) *rpcClient.Client {
	initctx := context.Background()
	c, err := rpcClient.NewClient(initctx, func(context.Context) (codec.ServerCodec, error) {
		p1, p2 := net.Pipe()
		go handler.ServeCodec(codec.NewCodec(p1), 0)
		return codec.NewCodec(p2), nil
	})
	if err != nil {
		log.Trace("creating new in process rpc client failed", "err", err)
	}
	return c
}
