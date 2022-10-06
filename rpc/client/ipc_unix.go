// Copyright 2015 The go-ethereum Authors
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

//go:build darwin || dragonfly || freebsd || linux || nacl || netbsd || openbsd || solaris
// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package client

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/log"
)

// ipcListen will create a Unix socket on the given endpoint.
func IpcListen(endpoint string) (net.Listener, error) {
	if len(endpoint) > int(maxPathSize) {
		log.Warn(fmt.Sprintf("The ipc endpoint is longer than %d characters. ", maxPathSize),
			"endpoint", endpoint)
	}

	// Ensure the IPC path exists and remove any previous leftover
	if err := os.MkdirAll(filepath.Dir(endpoint), 0o750); err != nil {
		return nil, err
	}
	err := os.Remove(endpoint)
	if err != nil {
		log.Trace("endpoint remove failed", "err", err)
	}
	l, err := net.Listen("unix", endpoint)
	if err != nil {
		return nil, err
	}
	err = os.Chmod(endpoint, 0o600)
	if err != nil {
		log.Trace("update ipc file permissions failed", "err", err)
		return nil, err
	}
	return l, nil
}

// newIPCConnection will connect to a Unix socket on the given endpoint.
func newIPCConnection(ctx context.Context, endpoint string) (net.Conn, error) {
	return new(net.Dialer).DialContext(ctx, "unix", endpoint)
}
