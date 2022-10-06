package client_test

import (
	"context"
	"net/http"
	"time"

	"github.com/evmos/ethermint/rpc/client"
)

// This example configures a HTTP-based RPC client with two options - one setting the
// overall request timeout, the other adding a custom HTTP header to all requests.
func ExampleDialOptions() {
	tokenHeader := client.WithHeader("x-token", "foo")
	httpClient := client.WithHTTPClient(&http.Client{
		Timeout: 10 * time.Second,
	})

	ctx := context.Background()
	c, err := client.DialOptions(ctx, "http://rpc.example.com", httpClient, tokenHeader)
	if err != nil {
		panic(err)
	}
	c.Close()
}
