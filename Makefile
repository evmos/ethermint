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
BINDIR ?= $(GOPATH)/bin
BUILDDIR ?= $(CURDIR)/build
SIMAPP = github.com/cosmos/ethermint/app
RUNSIM = $(BINDIR)/runsim
LEDGER_ENABLED ?= true

ifeq ($(DETECTED_OS),)
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
endif
export GO111MODULE = on

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
	go build -mod=readonly $(BUILD_FLAGS) -o build/$(DETECTED_OS)/$(ETHERMINT_DAEMON_BINARY).exe ./cmd/$(ETHERMINT_DAEMON_BINARY)
	go build -mod=readonly $(BUILD_FLAGS) -o build/$(DETECTED_OS)/$(ETHERMINT_CLI_BINARY).exe ./cmd/$(ETHERMINT_CLI_BINARY)
else
	go build -mod=readonly $(BUILD_FLAGS) -o build/$(DETECTED_OS)/$(ETHERMINT_DAEMON_BINARY) ./cmd/$(ETHERMINT_DAEMON_BINARY)
	go build -mod=readonly $(BUILD_FLAGS) -o build/$(DETECTED_OS)/$(ETHERMINT_CLI_BINARY) ./cmd/$(ETHERMINT_CLI_BINARY)
endif
	go build -mod=readonly ./...

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

docker:
	docker build -t ${DOCKER_IMAGE}:${DOCKER_TAG} .
	docker tag ${DOCKER_IMAGE}:${DOCKER_TAG} ${DOCKER_IMAGE}:latest
	docker tag ${DOCKER_IMAGE}:${DOCKER_TAG} ${DOCKER_IMAGE}:${COMMIT_HASH}
	# update old container
	docker rm ethermint
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

# Install the runsim binary with a temporary workaround of entering an outside
# directory as the "go get" command ignores the -mod option and will polute the
# go.{mod, sum} files.
#
# ref: https://github.com/golang/go/issues/30515
$(RUNSIM):
	@echo "Installing runsim..."
	@(cd /tmp && go get github.com/cosmos/tools/cmd/runsim@v1.0.0)

tools: $(RUNSIM)

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
	--blockchain blockchain --timeout=10m
	rm -rf importer/tmp

test-rpc:
	./scripts/integration-test-all.sh -q 1 -z 1 -s 2

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
	@echo "Running Ethermint import/export simulation. This may take several minutes..."
	@$(BINDIR)/runsim -Jobs=4 -SimAppPkg=$(SIMAPP) 25 5 TestAppImportExport

test-sim-after-import: runsim
	@echo "Running Ethermint simulation-after-import. This may take several minutes..."
	@$(BINDIR)/runsim -Jobs=4 -SimAppPkg=$(SIMAPP) 25 5 TestAppSimulationAfterImport

test-sim-custom-genesis-multi-seed: runsim
	@echo "Running multi-seed custom genesis simulation..."
	@echo "By default, ${HOME}/.$(ETHERMINT_DAEMON_BINARY)/config/genesis.json will be used."
	@$(BINDIR)/runsim -Jobs=4 -Genesis=${HOME}/.$(ETHERMINT_DAEMON_BINARY)/config/genesis.json 400 5 TestFullAppSimulation

test-sim-multi-seed-long: runsim
	@echo "Running multi-seed application simulation. This may take awhile!"
	@$(BINDIR)/runsim -Jobs=4 -SimAppPkg=$(SIMAPP) 500 50 TestFullAppSimulation

test-sim-multi-seed-short: runsim
	@echo "Running multi-seed application simulation. This may take awhile!"
	@$(BINDIR)/runsim -Jobs=4 -SimAppPkg=$(SIMAPP) 50 10 TestFullAppSimulation

.PHONY: runsim test-sim-nondeterminism test-sim-custom-genesis-fast test-sim-fast sim-import-export \
	test-sim-simulation-after-import test-sim-custom-genesis-multi-seed test-sim-multi-seed

.PHONY: build install update-tools tools godocs clean format lint \
test-cli test-race test-unit test test-import

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

proto-all: proto-gen proto-lint proto-check-breaking

proto-gen:
	@./scripts/protocgen.sh

proto-lint:
	@buf check lint --error-format=json

# NOTE: should match the default repo branch 
proto-check-breaking:
	@buf check breaking --against-input '.git#branch=development'


TM_URL           = https://raw.githubusercontent.com/tendermint/tendermint/v0.33.3
GOGO_PROTO_URL   = https://raw.githubusercontent.com/regen-network/protobuf/cosmos
COSMOS_PROTO_URL = https://raw.githubusercontent.com/regen-network/cosmos-proto/master
SDK_PROTO_URL 	 = https://raw.githubusercontent.com/cosmos/cosmos-sdk/master

TM_KV_TYPES         = third_party/proto/tendermint/libs/kv
TM_MERKLE_TYPES     = third_party/proto/tendermint/crypto/merkle
TM_ABCI_TYPES       = third_party/proto/tendermint/abci/types
GOGO_PROTO_TYPES    = third_party/proto/gogoproto
COSMOS_PROTO_TYPES  = third_party/proto/cosmos-proto
SDK_PROTO_TYPES     = third_party/proto/cosmos-sdk/types
AUTH_PROTO_TYPES    = third_party/proto/cosmos-sdk/x/auth/types
VESTING_PROTO_TYPES = third_party/proto/cosmos-sdk/x/auth/vesting/types
SUPPLY_PROTO_TYPES  = third_party/proto/cosmos-sdk/x/supply/types

proto-update-deps:
	@mkdir -p $(GOGO_PROTO_TYPES)
	@curl -sSL $(GOGO_PROTO_URL)/gogoproto/gogo.proto > $(GOGO_PROTO_TYPES)/gogo.proto

	@mkdir -p $(COSMOS_PROTO_TYPES)
	@curl -sSL $(COSMOS_PROTO_URL)/cosmos.proto > $(COSMOS_PROTO_TYPES)/cosmos.proto

	@mkdir -p $(TM_ABCI_TYPES)
	@curl -sSL $(TM_URL)/abci/types/types.proto > $(TM_ABCI_TYPES)/types.proto
	@sed -i '' '8 s|crypto/merkle/merkle.proto|third_party/proto/tendermint/crypto/merkle/merkle.proto|g' $(TM_ABCI_TYPES)/types.proto
	@sed -i '' '9 s|libs/kv/types.proto|third_party/proto/tendermint/libs/kv/types.proto|g' $(TM_ABCI_TYPES)/types.proto

	@mkdir -p $(TM_KV_TYPES)
	@curl -sSL $(TM_URL)/libs/kv/types.proto > $(TM_KV_TYPES)/types.proto

	@mkdir -p $(TM_MERKLE_TYPES)
	@curl -sSL $(TM_URL)/crypto/merkle/merkle.proto > $(TM_MERKLE_TYPES)/merkle.proto

	@mkdir -p $(SDK_PROTO_TYPES)
	@curl -sSL $(SDK_PROTO_URL)/types/types.proto > $(SDK_PROTO_TYPES)/types.proto

	@mkdir -p $(AUTH_PROTO_TYPES)
	@curl -sSL $(SDK_PROTO_URL)/x/auth/types/types.proto > $(AUTH_PROTO_TYPES)/types.proto
	@sed -i '' '5 s|types/types.proto|third_party/proto/cosmos-sdk/types/types.proto|g' $(AUTH_PROTO_TYPES)/types.proto

	@mkdir -p $(VESTING_PROTO_TYPES)
	curl -sSL $(SDK_PROTO_URL)/x/auth/vesting/types/types.proto > $(VESTING_PROTO_TYPES)/types.proto
	@sed -i '' '5 s|types/types.proto|third_party/proto/cosmos-sdk/types/types.proto|g' $(VESTING_PROTO_TYPES)/types.proto
	@sed -i '' '6 s|x/auth/types/types.proto|third_party/proto/cosmos-sdk/x/auth/types/types.proto|g' $(VESTING_PROTO_TYPES)/types.proto

	@mkdir -p $(SUPPLY_PROTO_TYPES)
	curl -sSL $(SDK_PROTO_URL)/x/supply/types/types.proto > $(SUPPLY_PROTO_TYPES)/types.proto
	@sed -i '' '5 s|types/types.proto|third_party/proto/cosmos-sdk/types/types.proto|g' $(SUPPLY_PROTO_TYPES)/types.proto
	@sed -i '' '6 s|x/auth/types/types.proto|third_party/proto/cosmos-sdk/x/auth/types/types.proto|g' $(SUPPLY_PROTO_TYPES)/types.proto


.PHONY: proto-all proto-gen proto-lint proto-check-breaking proto-update-deps


###############################################################################
###                              Documentation                              ###
###############################################################################

# Start docs site at localhost:8080
docs-serve:
	@cd docs && \
	npm install && \
	npm run serve

# Build the site into docs/.vuepress/dist
docs-build:
	@cd docs && \
	npm install && \
	npm run build

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

.PHONY: build-docker-local-ethermint localnet-start localnet-stop
