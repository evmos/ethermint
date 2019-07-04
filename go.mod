module github.com/cosmos/ethermint

go 1.12

require (
	github.com/aristanetworks/goarista v0.0.0-20181101003910-5bb443fba8e0 // indirect
	github.com/cespare/cp v1.1.1 // indirect
	github.com/cosmos/cosmos-sdk v0.28.2-0.20190704145406-01d442565807
	github.com/deckarep/golang-set v1.7.1 // indirect
	github.com/edsrzf/mmap-go v0.0.0-20170320065105-0bce6a688712 // indirect
	github.com/elastic/gosigar v0.10.3 // indirect
	github.com/ethereum/go-ethereum v1.8.27
	github.com/fjl/memsize v0.0.0-20180929194037-2a09253e352a // indirect
	github.com/golangci/golangci-lint v1.17.1 // indirect
	github.com/google/uuid v1.0.0 // indirect
	github.com/gordonklaus/ineffassign v0.0.0-20190601041439-ed7b1b5ee0f8 // indirect
	github.com/hashicorp/golang-lru v0.5.0 // indirect
	github.com/huin/goupnp v1.0.0 // indirect
	github.com/influxdata/influxdb v1.7.7 // indirect
	github.com/jackpal/go-nat-pmp v1.0.1 // indirect
	github.com/karalabe/hid v0.0.0-20180420081245-2b4488a37358 // indirect
	github.com/kisielk/errcheck v1.2.0 // indirect
	github.com/mattn/go-colorable v0.1.2 // indirect
	github.com/mdempsky/unconvert v0.0.0-20190325185700-2f5dc3378ed3 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/onsi/ginkgo v1.8.0 // indirect
	github.com/onsi/gomega v1.5.0 // indirect
	github.com/pborman/uuid v0.0.0-20180906182336-adf5a7427709 // indirect
	github.com/pkg/errors v0.8.1
	github.com/rjeczalik/notify v0.9.2 // indirect
	github.com/stretchr/testify v1.3.0
	github.com/syndtr/goleveldb v0.0.0-20181105012736-f9080354173f // indirect
	github.com/tendermint/lint v0.0.1 // indirect
	github.com/tendermint/tendermint v0.32.0
	golang.org/x/crypto v0.0.0-20190701094942-4def268fd1a4 // indirect
	golang.org/x/net v0.0.0-20190628185345-da137c7871d7 // indirect
	golang.org/x/sys v0.0.0-20190626221950-04f50cda93cb // indirect
	golang.org/x/text v0.3.2 // indirect
	google.golang.org/genproto v0.0.0-20181107211654-5fc9ac540362 // indirect
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce // indirect
	gopkg.in/olebedev/go-duktape.v3 v3.0.0-20190213234257-ec84240a7772 // indirect
	gopkg.in/urfave/cli.v1 v1.20.0 // indirect
	mvdan.cc/unparam v0.0.0-20190310220240-1b9ccfa71afe // indirect
)

replace (
	github.com/ethereum/go-ethereum v1.8.27 => github.com/alexanderbez/go-ethereum v1.8.17-0.20181024144731-0a57b29f0c8e
	golang.org/x/crypto => github.com/tendermint/crypto v0.0.0-20180820045704-3764759f34a5
)
