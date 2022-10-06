// Copyright 2020 The go-ethereum Authors
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

package client

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/metrics"
)

// default value for metrics is nil holder to prevent
// storing the metrics if not enabled
var (
	// rpcRequestGauge stores count of rpc requests made
	rpcRequestGauge metrics.Gauge = metrics.NilGauge{}
	// successfulRequestGauge stores count of successful rpc requests
	successfulRequestGauge metrics.Gauge = metrics.NilGauge{}
	// failedRequestGauge stores count of failed rpc requests
	failedRequestGauge metrics.Gauge = metrics.NilGauge{}

	// serveTimeHistName is the prefix of the per-request serving time histograms.
	serveTimeHistName = "rpc/duration"

	// rpcServingTimer captures duration and rate of rpc calls
	rpcServingTimer metrics.Timer = metrics.NilTimer{}

	MetricsEnabled = false

	registry = metrics.NewRegistry()
)

// enableMetrics creates and registers metrics with registry
func EnableMetrics(r metrics.Registry) {
	registry = r

	MetricsEnabled = true

	// set go-ethereum metrics to true. Needed for the registered gauges
	metrics.Enabled = true

	rpcRequestGauge = metrics.NewRegisteredGauge("rpc/requests", registry)
	successfulRequestGauge = metrics.NewRegisteredGauge("rpc/success", registry)
	failedRequestGauge = metrics.NewRegisteredGauge("rpc/failure", registry)

	rpcServingTimer = metrics.NewRegisteredTimer("rpc/duration/all", registry)
}

// updateServeTimeHistogram tracks the serving time of a remote RPC call.
func updateServeTimeHistogram(method string, success bool, elapsed time.Duration) {
	if MetricsEnabled {
		note := "success"
		if !success {
			note = "failure"
		}
		h := fmt.Sprintf("%s/%s/%s", serveTimeHistName, method, note)
		sampler := func() metrics.Sample {
			return metrics.ResettingSample(
				metrics.NewExpDecaySample(1028, 0.015),
			)
		}
		metrics.GetOrRegisterHistogramLazy(h, registry, sampler).Update(elapsed.Microseconds())
	}
}
