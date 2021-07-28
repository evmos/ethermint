module github.com/tharsis/ethermint

go 1.16

require (
	github.com/armon/go-metrics v0.3.9
	github.com/btcsuite/btcd v0.22.0-beta
	github.com/btcsuite/btcutil v1.0.3-0.20201208143702-a53e38424cce
	github.com/cespare/cp v1.1.1 // indirect
	github.com/cosmos/cosmos-sdk v0.43.0-rc2
	github.com/cosmos/go-bip39 v1.0.0
	github.com/cosmos/ibc-go v1.0.0-rc3
	github.com/deckarep/golang-set v1.7.1 // indirect
	github.com/ethereum/go-ethereum v1.10.3
	github.com/gogo/protobuf v1.3.3
	github.com/golang/protobuf v1.5.2
	github.com/google/uuid v1.2.0 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.4.2
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/hashicorp/go-immutable-radix v1.3.0 // indirect
	github.com/improbable-eng/grpc-web v0.14.0
	github.com/miguelmota/go-ethereum-hdwallet v0.0.1
	github.com/palantir/stacktrace v0.0.0-20161112013806-78658fd2d177
	github.com/pkg/errors v0.9.1
	github.com/prometheus/tsdb v0.10.0 // indirect
	github.com/rakyll/statik v0.1.7
	github.com/regen-network/cosmos-proto v0.3.1
	github.com/rjeczalik/notify v0.9.2 // indirect
	github.com/rs/cors v1.8.0
	github.com/rs/zerolog v1.23.0
	github.com/spf13/cast v1.4.0
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.8.1
	github.com/status-im/keycard-go v0.0.0-20200402102358-957c09536969 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/tendermint v0.34.11
	github.com/tendermint/tm-db v0.6.4
	github.com/tyler-smith/go-bip39 v1.1.0
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97 // indirect
	google.golang.org/genproto v0.0.0-20210712171009-5404628d0f46
	google.golang.org/grpc v1.39.0
	gopkg.in/yaml.v2 v2.4.0
	nhooyr.io/websocket v1.8.7 // indirect
)

replace google.golang.org/grpc => google.golang.org/grpc v1.33.2

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

// FIXME: update after PR merged: https://github.com/bugsnag/panicwrap/pull/23
replace github.com/bugsnag/panicwrap => github.com/yihuang/panicwrap v1.3.4-0.20210716032932-61c0a7c0cd05
