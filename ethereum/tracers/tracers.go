// Copyright 2021 The go-ethereum Authors
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

package tracers

import (
	"encoding/json"
	"errors"

	ethtracers "github.com/ethereum/go-ethereum/eth/tracers"

	evm "github.com/evmos/ethermint/x/evm/vm"
)

var lookups []lookupFunc

type lookupFunc func(string, *ethtracers.Context, json.RawMessage) (Tracer, error)

type Tracer interface {
	evm.Logger
	GetResult() (json.RawMessage, error)
	// Stop terminates execution of the tracer at the first opportune moment.
	Stop(err error)
}

// New returns a new instance of a tracer, by iterating through the
// registered lookups.
func New(code string, ctx *ethtracers.Context, cfg json.RawMessage) (Tracer, error) {
	for _, lookup := range lookups {
		if tracer, err := lookup(code, ctx, cfg); err == nil {
			return tracer, nil
		}
	}
	return nil, errors.New("tracer not found")
}
