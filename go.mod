module github.com/cosmos/ethermint

go 1.14

require (
	github.com/allegro/bigcache v1.2.1 // indirect
	github.com/aristanetworks/goarista v0.0.0-20200331225509-2cc472e8fbd6 // indirect
	github.com/cespare/cp v1.1.1 // indirect
	github.com/cosmos/cosmos-sdk v0.34.4-0.20200403200637-7f78e61b93a5
	github.com/deckarep/golang-set v1.7.1 // indirect
	github.com/ethereum/go-ethereum v1.9.17
	github.com/fjl/memsize v0.0.0-20190710130421-bcb5799ab5e5 // indirect
	github.com/gogo/protobuf v1.3.1
	github.com/gorilla/mux v1.7.4
	github.com/gorilla/websocket v1.4.2
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/onsi/ginkgo v1.11.0 // indirect
	github.com/onsi/gomega v1.8.1 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/tsdb v0.9.1 // indirect
	github.com/regen-network/cosmos-proto v0.1.1-0.20200213154359-02baa11ea7c2
	github.com/rjeczalik/notify v0.9.2 // indirect
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.7.0
	github.com/status-im/keycard-go v0.0.0-20190424133014-d95853db0f48
	github.com/stretchr/testify v1.6.1
	github.com/tendermint/tendermint v0.33.4
	github.com/tendermint/tm-db v0.5.1
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	gopkg.in/yaml.v2 v2.3.0
)

// forked SDK to avoid breaking changes
replace github.com/cosmos/cosmos-sdk => github.com/Chainsafe/cosmos-sdk v0.34.4-0.20200622114457-35ea97f29c5f
