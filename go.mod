module github.com/tharsis/ethermint

go 1.16

require (
	github.com/DataDog/zstd v1.4.8 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible // indirect
	github.com/aokoli/goutils v1.1.1 // indirect
	github.com/armon/go-metrics v0.3.9
	github.com/btcsuite/btcd v0.22.0-beta
	github.com/btcsuite/btcutil v1.0.3-0.20201208143702-a53e38424cce
	github.com/cespare/cp v1.1.1 // indirect
	github.com/cosmos/cosmos-sdk v0.43.0
	github.com/cosmos/go-bip39 v1.0.0
	github.com/cosmos/ibc-go v1.0.0
	github.com/deckarep/golang-set v1.7.1 // indirect
	github.com/dgraph-io/badger/v2 v2.2007.3 // indirect
	github.com/dgraph-io/ristretto v0.1.0 // indirect
	github.com/envoyproxy/protoc-gen-validate v0.6.1 // indirect
	github.com/ethereum/go-ethereum v1.10.7
	github.com/gogo/protobuf v1.3.3
	github.com/golang/glog v0.0.0-20210429001901-424d2337a529 // indirect
	github.com/golang/protobuf v1.5.2
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/btree v1.0.1 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.4.2
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/hashicorp/go-immutable-radix v1.3.0 // indirect
	github.com/hashicorp/go-uuid v1.0.2 // indirect
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/improbable-eng/grpc-web v0.14.0
	github.com/klauspost/compress v1.11.9 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/miguelmota/go-ethereum-hdwallet v0.1.1
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mwitkow/go-proto-validators v0.3.2 // indirect
	github.com/palantir/stacktrace v0.0.0-20161112013806-78658fd2d177
	github.com/pkg/errors v0.9.1
	github.com/prometheus/tsdb v0.10.0 // indirect
	github.com/pseudomuto/protoc-gen-doc v1.5.0 // indirect
	github.com/rakyll/statik v0.1.7
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/regen-network/cosmos-proto v0.3.1
	github.com/rjeczalik/notify v0.9.2 // indirect
	github.com/rs/cors v1.8.0
	github.com/rs/zerolog v1.23.0
	github.com/spf13/cast v1.4.1
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.8.1
	github.com/status-im/keycard-go v0.0.0-20200402102358-957c09536969
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/tendermint v0.34.12
	github.com/tendermint/tm-db v0.6.4
	github.com/tyler-smith/go-bip39 v1.1.0
	go.etcd.io/bbolt v1.3.6 // indirect
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5 // indirect
	golang.org/x/net v0.0.0-20210805182204-aaa1db679c0d // indirect
	golang.org/x/sys v0.0.0-20210820121016-41cdb8703e55 // indirect
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba // indirect
	google.golang.org/genproto v0.0.0-20210820002220-43fce44e7af1
	google.golang.org/grpc v1.40.0
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.4.0
	nhooyr.io/websocket v1.8.7 // indirect
)

replace google.golang.org/grpc => google.golang.org/grpc v1.33.2

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

replace github.com/99designs/keyring => github.com/cosmos/keyring v1.1.7-0.20210622111912-ef00f8ac3d76
