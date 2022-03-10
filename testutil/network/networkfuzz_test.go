//go:build go1.18
// +build go1.18

package network

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

func FuzzNetworkRPC(f *testing.F) {
	f.Fuzz(func(t *testing.T, msg []byte) {
		var ethJson *ethtypes.Transaction = new(ethtypes.Transaction)
		jsonErr := json.Unmarshal(msg, ethJson)
		if jsonErr == nil {
			testnetwork := New(t, DefaultConfig())
			testnetwork.Validators[0].JSONRPCClient.SendTransaction(context.Background(), ethJson)
			h, err := testnetwork.WaitForHeightWithTimeout(10, time.Minute)
			if err != nil {
				t.Fatalf("expected to reach 10 blocks; got %d", h)
			}
			latestHeight, err := testnetwork.LatestHeight()
			if err != nil {
				t.Fatalf("latest height failed")
			}
			if latestHeight < h {
				t.Errorf("latestHeight should be greater or equal to")
			}
			testnetwork.Cleanup()
		} else {
			t.Skip("invalid tx")
		}
	})
}
