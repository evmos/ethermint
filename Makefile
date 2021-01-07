# Copyright 2018 Tendermint. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#!/usr/bin/make -f

VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')
PACKAGES=$(shell go list ./... | grep -Ev 'vendor|importer|rpc/tester')
DOCKER_TAG = unstable
DOCKER_IMAGE = cosmos/ethermint
ETHERMINT_DAEMON_BINARY = ethermintd
ETHERMINT_CLI_BINARY = ethermintcli
GO_MOD=GO111MODULE=on
BUILDDIR ?= $(CURDIR)/build
SIMAPP = ./app
LEDGER_ENABLED ?= true
HTTPS_GIT := https://github.com/cosmos/ethermint.git
DOCKER := $(shell which docker)
DOCKER_BUF := $(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace bufbuild/buf


ifeq ($(OS),Windows_NT)
  DETECTED_OS := windows
else
  UNAME_S = $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
	DETECTED_OS := mac
  else
	DETECTED_OS := linux
  endif
endif
export DETECTED_OS
export GO111MODULE = on

##########################################
# Find OS and Go environment
# GO contains the Go binary
# FS contains the OS file separator
##########################################

ifeq ($(OS),Windows_NT)
  GO := $(shell where go.exe 2> NUL)
  FS := "\\"
else
  GO := $(shell command -v go 2> /dev/null)
  FS := "/"
endif

ifeq ($(GO),)
  $(error could not find go. Is it in PATH? $(GO))
endif

GOPATH ?= $(shell $(GO) env GOPATH)
BINDIR ?= $(GOPATH)/bin
RUNSIM = $(BINDIR)/runsim


# process build tags

build_tags = netgo
ifeq ($(LEDGER_ENABLED),true)
  ifeq ($(OS),Windows_NT)
    GCCEXE = $(shell where gcc.exe 2> NUL)
    ifeq ($(GCCEXE),)
      $(error gcc.exe not installed for ledger support, please install or set LEDGER_ENABLED=false)
    else
      build_tags += ledger
    endif
  else
    UNAME_S = $(shell uname -s)
    ifeq ($(UNAME_S),OpenBSD)
      $(warning OpenBSD detected, disabling ledger support (https://github.com/cosmos/cosmos-sdk/issues/1988))
    else
      GCC = $(shell command -v gcc 2> /dev/null)
      ifeq ($(GCC),)
        $(error gcc not installed for ledger support, please install or set LEDGER_ENABLED=false)
      else
        build_tags += ledger
      endif
    endif
  endif
endif

ifeq ($(WITH_CLEVELDB),yes)
  build_tags += gcc
endif
build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

whitespace :=
whitespace += $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))

# process linker flags

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=ethermint \
		  -X github.com/cosmos/cosmos-sdk/version.ServerName=$(ETHERMINT_DAEMON_BINARY) \
		  -X github.com/cosmos/cosmos-sdk/version.ClientName=$(ETHERMINT_CLI_BINARY) \
		  -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
		  -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
		  -X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)"

ifeq ($(WITH_CLEVELDB),yes)
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=cleveldb
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'

all: tools verify install

###############################################################################
###                                  Build                                  ###
###############################################################################

build: go.sum

ifeq ($(OS), Windows_NT)
	go build -mod=readonly $(BUILD_FLAGS) -o build/$(ETHERMINT_DAEMON_BINARY).exe ./cmd/$(ETHERMINT_DAEMON_BINARY)
	go build -mod=readonly $(BUILD_FLAGS) -o build/$(ETHERMINT_CLI_BINARY).exe ./cmd/$(ETHERMINT_CLI_BINARY)
else
	go build -mod=readonly $(BUILD_FLAGS) -o build/$(ETHERMINT_DAEMON_BINARY) ./cmd/$(ETHERMINT_DAEMON_BINARY)
	go build -mod=readonly $(BUILD_FLAGS) -o build/$(ETHERMINT_CLI_BINARY) ./cmd/$(ETHERMINT_CLI_BINARY)
endif

build-ethermint: go.sum
	mkdir -p $(BUILDDIR)
	go build -mod=readonly $(BUILD_FLAGS) -o $(BUILDDIR) ./cmd/$(ETHERMINT_DAEMON_BINARY)
	go build -mod=readonly $(BUILD_FLAGS) -o $(BUILDDIR) ./cmd/$(ETHERMINT_CLI_BINARY)

build-ethermint-linux: go.sum
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 $(MAKE) build-ethermint

.PHONY: build build-ethermint build-ethermint-linux

install:
	${GO_MOD} go install $(BUILD_FLAGS) ./cmd/$(ETHERMINT_DAEMON_BINARY)
	${GO_MOD} go install $(BUILD_FLAGS) ./cmd/$(ETHERMINT_CLI_BINARY)

clean:
	@rm -rf ./build ./vendor

.PHONY: install clean

docker-build:
	docker build -t ${DOCKER_IMAGE}:${DOCKER_TAG} .
	docker tag ${DOCKER_IMAGE}:${DOCKER_TAG} ${DOCKER_IMAGE}:latest
	# docker tag ${DOCKER_IMAGE}:${DOCKER_TAG} ${DOCKER_IMAGE}:${COMMIT_HASH}
	# update old container
	docker rm ethermint || true
	# create a new container from the latest image
	docker create --name ethermint -t -i cosmos/ethermint:latest ethermint
	# move the binaries to the ./build directory
	mkdir -p ./build/
	docker cp ethermint:/usr/bin/ethermintd ./build/ ; \
	docker cp ethermint:/usr/bin/ethermintcli ./build/

docker-localnet:
	docker build -f ./networks/local/ethermintnode/Dockerfile . -t ethermintd/node

###############################################################################
###                          Tools & Dependencies                           ###
###############################################################################

TOOLS_DESTDIR  ?= $(GOPATH)/bin
RUNSIM         = $(TOOLS_DESTDIR)/runsim

# Install the runsim binary with a temporary workaround of entering an outside
# directory as the "go get" command ignores the -mod option and will polute the
# go.{mod, sum} files.
#
# ref: https://github.com/golang/go/issues/30515
runsim: $(RUNSIM)
$(RUNSIM):
	@echo "Installing runsim..."
	@(cd /tmp && ${GO_MOD} go get github.com/cosmos/tools/cmd/runsim@v1.0.0)

contract-tools:
ifeq (, $(shell which stringer))
	@echo "Installing stringer..."
	@go install golang.org/x/tools/cmd/stringer
else
	@echo "stringer already installed; skipping..."
endif

ifeq (, $(shell which go-bindata))
	@echo "Installing go-bindata..."
	@go install github.com/kevinburke/go-bindata/go-bindata
else
	@echo "go-bindata already installed; skipping..."
endif

ifeq (, $(shell which gencodec))
	@echo "Installing gencodec..."
	@go install github.com/fjl/gencodec
else
	@echo "gencodec already installed; skipping..."
endif

ifeq (, $(shell which protoc-gen-go))
	@echo "Installing protoc-gen-go..."
	@go install github.com/golang/protobuf/protoc-gen-go
else
	@echo "protoc-gen-go already installed; skipping..."
endif

ifeq (, $(shell which solcjs))
	@echo "Installing solcjs..."
	@apt-get install -f -y protobuf-compiler
	@npm install -g solc@0.5.11
else
	@echo "solcjs already installed; skipping..."
endif

docs-tools:
ifeq (, $(shell which yarn))
	@echo "Installing yarn..."
	@npm install -g yarn
else
	@echo "yarn already installed; skipping..."
endif

tools: tools-stamp
tools-stamp: contract-tools docs-tools runsim
	# Create dummy file to satisfy dependency and avoid
	# rebuilding when this Makefile target is hit twice
	# in a row.
	touch $@

tools-clean:
	rm -f $(RUNSIM)
	rm -f tools-stamp

docs-tools-stamp: docs-tools
	# Create dummy file to satisfy dependency and avoid
	# rebuilding when this Makefile target is hit twice
	# in a row.
	touch $@

.PHONY: runsim tools tools-stamp tools-clean docs-tools-stamp

###############################################################################
###                           Tests & Simulation                            ###
###############################################################################

test: test-unit

test-unit:
	@go test -v ./... $(PACKAGES)

test-race:
	@go test -v --vet=off -race ./... $(PACKAGES)

test-import:
	@go test ./importer -v --vet=off --run=TestImportBlocks --datadir tmp \
	--blockchain blockchain
	rm -rf importer/tmp

test-rpc:
	./scripts/integration-test-all.sh -t "rpc" -q 1 -z 1 -s 2 -m "rpc"

test-rpc-pending:
	./scripts/integration-test-all.sh -t "pending" -q 1 -z 1 -s 2 -m "pending"

test-contract:
	@type "npm" 2> /dev/null || (echo 'Npm does not exist. Please install node.js and npm."' && exit 1)
	@type "solcjs" 2> /dev/null || (echo 'Solcjs does not exist. Please install solcjs using make contract-tools."' && exit 1)
	@type "protoc" 2> /dev/null || (echo 'Failed to install protoc. Please reinstall protoc using make contract-tools.' && exit 1)
	bash scripts/contract-test.sh

test-sim-nondeterminism:
	@echo "Running non-determinism test..."
	@go test -mod=readonly $(SIMAPP) -run TestAppStateDeterminism -Enabled=true \
		-NumBlocks=100 -BlockSize=200 -Commit=true -Period=0 -v -timeout 24h

test-sim-custom-genesis-fast:
	@echo "Running custom genesis simulation..."
	@echo "By default, ${HOME}/.$(ETHERMINT_DAEMON_BINARY)/config/genesis.json will be used."
	@go test -mod=readonly $(SIMAPP) -run TestFullAppSimulation -Genesis=${HOME}/.$(ETHERMINT_DAEMON_BINARY)/config/genesis.json \
		-Enabled=true -NumBlocks=100 -BlockSize=200 -Commit=true -Seed=99 -Period=5 -v -timeout 24h

test-sim-import-export: runsim
	@echo "Running application import/export simulation. This may take several minutes..."
	@$(BINDIR)/runsim -Jobs=4 -SimAppPkg=$(SIMAPP) -ExitOnFail 50 5 TestAppImportExport

test-sim-after-import: runsim
	@echo "Running application simulation-after-import. This may take several minutes..."
	@$(BINDIR)/runsim -Jobs=4 -SimAppPkg=$(SIMAPP) -ExitOnFail 50 5 TestAppSimulationAfterImport

test-sim-custom-genesis-multi-seed: runsim
	@echo "Running multi-seed custom genesis simulation..."
	@echo "By default, ${HOME}/.$(ETHERMINT_DAEMON_BINARY)/config/genesis.json will be used."
	@$(BINDIR)/runsim -Jobs=4 -SimAppPkg=$(SIMAPP) -ExitOnFail -Genesis=${HOME}/.$(ETHERMINT_DAEMON_BINARY)/config/genesis.json 400 5 TestFullAppSimulation

test-sim-multi-seed-long: runsim
	@echo "Running multi-seed application simulation. This may take awhile!"
	@$(BINDIR)/runsim -Jobs=4 -SimAppPkg=$(SIMAPP) -ExitOnFail 500 50 TestFullAppSimulation

test-sim-multi-seed-short: runsim
	@echo "Running multi-seed application simulation. This may take awhile!"
	@$(BINDIR)/runsim -Jobs=4 -SimAppPkg=$(SIMAPP) -ExitOnFail 50 10 TestFullAppSimulation

test-solidity:
	 @echo "Beginning solidity tests..."
	 ./scripts/run-solidity-tests.sh

.PHONY: test test-unit test-race test-import test-rpc test-contract test-solidity

.PHONY: test-sim-nondeterminism test-sim-custom-genesis-fast test-sim-import-export test-sim-after-import \
	test-sim-custom-genesis-multi-seed test-sim-multi-seed-long test-sim-multi-seed-short

###############################################################################
###                                Linting                                  ###
###############################################################################

lint:
	golangci-lint run --out-format=tab --issues-exit-code=0
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" | xargs gofmt -d -s

format:
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -name '*.pb.go' | xargs gofmt -w -s
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -name '*.pb.go' | xargs misspell -w
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -name '*.pb.go' | xargs goimports -w -local github.com/tendermint
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -name '*.pb.go' | xargs goimports -w -local github.com/ethereum/go-ethereum
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -name '*.pb.go' | xargs goimports -w -local github.com/cosmos/cosmos-sdk
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -name '*.pb.go' | xargs goimports -w -local github.com/cosmos/ethermint

.PHONY: lint format

###############################################################################
###                                Protobuf                                 ###
###############################################################################

proto-all: proto-format proto-lint proto-gen

proto-gen:
	@echo "Generating Protobuf files"
	$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace tendermintdev/sdk-proto-gen sh ./scripts/protocgen.sh

proto-format:
	@echo "Formatting Protobuf files"
	$(DOCKER) run --rm -v $(CURDIR):/workspace \
	--workdir /workspace tendermintdev/docker-build-proto \
	find ./ -not -path "./third_party/*" -name *.proto -exec clang-format -i {} \;

proto-swagger-gen:
	@./scripts/protoc-swagger-gen.sh

proto-lint:
	@$(DOCKER_BUF) check lint --error-format=json

proto-check-breaking:
	@$(DOCKER_BUF) check breaking --against-input $(HTTPS_GIT)#branch=development

TM_URL              = https://raw.githubusercontent.com/tendermint/tendermint/v0.34.0-rc6/proto/tendermint
GOGO_PROTO_URL      = https://raw.githubusercontent.com/regen-network/protobuf/cosmos
COSMOS_SDK_URL      = https://raw.githubusercontent.com/cosmos/cosmos-sdk/master
COSMOS_PROTO_URL    = https://raw.githubusercontent.com/regen-network/cosmos-proto/master

TM_CRYPTO_TYPES     = third_party/proto/tendermint/crypto
TM_ABCI_TYPES       = third_party/proto/tendermint/abci
TM_TYPES            = third_party/proto/tendermint/types

GOGO_PROTO_TYPES    = third_party/proto/gogoproto
COSMOS_SDK_PROTO    = third_party/proto/cosmos-sdk
COSMOS_PROTO_TYPES  = third_party/proto/cosmos_proto

proto-update-deps:
	@mkdir -p $(GOGO_PROTO_TYPES)
	@curl -sSL $(GOGO_PROTO_URL)/gogoproto/gogo.proto > $(GOGO_PROTO_TYPES)/gogo.proto

	@mkdir -p $(COSMOS_PROTO_TYPES)
	@curl -sSL $(COSMOS_PROTO_URL)/cosmos.proto > $(COSMOS_PROTO_TYPES)/cosmos.proto

## Importing of tendermint protobuf definitions currently requires the
## use of `sed` in order to build properly with cosmos-sdk's proto file layout
## (which is the standard Buf.build FILE_LAYOUT)
## Issue link: https://github.com/tendermint/tendermint/issues/5021
	@mkdir -p $(TM_ABCI_TYPES)
	@curl -sSL $(TM_URL)/abci/types.proto > $(TM_ABCI_TYPES)/types.proto

	@mkdir -p $(TM_TYPES)
	@curl -sSL $(TM_URL)/types/types.proto > $(TM_TYPES)/types.proto

	@mkdir -p $(TM_CRYPTO_TYPES)
	@curl -sSL $(TM_URL)/crypto/proof.proto > $(TM_CRYPTO_TYPES)/proof.proto
	@curl -sSL $(TM_URL)/crypto/keys.proto > $(TM_CRYPTO_TYPES)/keys.proto

.PHONY: proto-all proto-gen proto-gen-any proto-swagger-gen proto-format proto-lint proto-check-breaking proto-update-deps



###############################################################################
###                              Documentation                              ###
###############################################################################

# Start docs site at localhost:8080
docs-serve:
	@cd docs && \
	yarn install && \
	yarn run serve

# Build the site into docs/.vuepress/dist
docs-build:
	@$(MAKE) docs-tools-stamp && \
	cd docs && \
	yarn install && \
	yarn run build

godocs:
	@echo "--> Wait a few seconds and visit http://localhost:6060/pkg/github.com/cosmos/ethermint"
	godoc -http=:6060

###############################################################################
###                                Localnet                                 ###
###############################################################################

build-docker-local-ethermint:
	@$(MAKE) -C networks/local

# Run a 4-node testnet locally
localnet-start: localnet-stop
ifeq ($(OS),Windows_NT)
	mkdir build &
	@$(MAKE) docker-localnet

	IF not exist "build/node0/$(ETHERMINT_DAEMON_BINARY)/config/genesis.json" docker run --rm -v $(CURDIR)/build\ethermint\Z ethermintd/node "ethermintd testnet --v 4 -o /ethermint --starting-ip-address 192.168.10.2 --keyring-backend=test"
	docker-compose up -d
else
	mkdir -p ./build/
	@$(MAKE) docker-localnet

	if ! [ -f build/node0/$(ETHERMINT_DAEMON_BINARY)/config/genesis.json ]; then docker run --rm -v $(CURDIR)/build:/ethermint:Z ethermintd/node "ethermintd testnet --v 4 -o /ethermint --starting-ip-address 192.168.10.2 --keyring-backend=test"; fi
	docker-compose up -d
endif

localnet-stop:
	docker-compose down

# clean testnet
localnet-clean:
	docker-compose down
	sudo rm -rf build/*

 # reset testnet
localnet-unsafe-reset:
	docker-compose down
ifeq ($(OS),Windows_NT)
	@docker run --rm -v $(CURDIR)/build\ethermint\Z ethermintd/node "ethermintd unsafe-reset-all --home=/ethermint/node0/ethermintd"
	@docker run --rm -v $(CURDIR)/build\ethermint\Z ethermintd/node "ethermintd unsafe-reset-all --home=/ethermint/node1/ethermintd"
	@docker run --rm -v $(CURDIR)/build\ethermint\Z ethermintd/node "ethermintd unsafe-reset-all --home=/ethermint/node2/ethermintd"
	@docker run --rm -v $(CURDIR)/build\ethermint\Z ethermintd/node "ethermintd unsafe-reset-all --home=/ethermint/node3/ethermintd"
else
	@docker run --rm -v $(CURDIR)/build:/ethermint:Z ethermintd/node "ethermintd unsafe-reset-all --home=/ethermint/node0/ethermintd"
	@docker run --rm -v $(CURDIR)/build:/ethermint:Z ethermintd/node "ethermintd unsafe-reset-all --home=/ethermint/node1/ethermintd"
	@docker run --rm -v $(CURDIR)/build:/ethermint:Z ethermintd/node "ethermintd unsafe-reset-all --home=/ethermint/node2/ethermintd"
	@docker run --rm -v $(CURDIR)/build:/ethermint:Z ethermintd/node "ethermintd unsafe-reset-all --home=/ethermint/node3/ethermintd"
endif

.PHONY: build-docker-local-ethermint localnet-start localnet-stop
