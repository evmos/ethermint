//go:build go1.18
// +build go1.18

package network

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/tharsis/ethermint/testutil/network"
)

func FuzzNetworkRPC(f *testing.F) {

	f.Fuzz(func(t *testing.T, msg []byte) {
		var ethjson *ethtypes.Transaction = new(ethtypes.Transaction)
		jsonerr := json.Unmarshal(msg, ethjson)
		if jsonerr == nil {
			testnetwork := network.New(t, network.DefaultConfig())
			testnetwork.Validators[0].JSONRPCClient.SendTransaction(context.Background(), ethjson)
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
