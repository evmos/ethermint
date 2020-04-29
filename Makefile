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

PACKAGES=$(shell go list ./... | grep -Ev 'vendor|importer|rpc/tester')
COMMIT_HASH := $(shell git rev-parse --short HEAD)
BUILD_FLAGS = -tags netgo -ldflags "-X github.com/cosmos/ethermint/version.GitCommit=${COMMIT_HASH}"
DOCKER_TAG = unstable
DOCKER_IMAGE = cosmos/ethermint
ETHERMINT_DAEMON_BINARY = emintd
ETHERMINT_CLI_BINARY = emintcli
GO_MOD=GO111MODULE=on
BINDIR ?= $(GOPATH)/bin
SIMAPP = github.com/cosmos/ethermint/app
RUNSIM = $(BINDIR)/runsim

all: tools verify install

#######################
### Build / Install ###
#######################

build:
ifeq ($(OS),Windows_NT)
	${GO_MOD} go build $(BUILD_FLAGS) -o build/$(ETHERMINT_DAEMON_BINARY).exe ./cmd/emintd
	${GO_MOD} go build $(BUILD_FLAGS) -o build/$(ETHERMINT_CLI_BINARY).exe ./cmd/emintcli
else
	${GO_MOD} go build $(BUILD_FLAGS) -o build/$(ETHERMINT_DAEMON_BINARY) ./cmd/emintd/
	${GO_MOD} go build $(BUILD_FLAGS) -o build/$(ETHERMINT_CLI_BINARY) ./cmd/emintcli/
endif

install:
	${GO_MOD} go install $(BUILD_FLAGS) ./cmd/emintd
	${GO_MOD} go install $(BUILD_FLAGS) ./cmd/emintcli

clean:
	@rm -rf ./build ./vendor

update-tools:
	@echo "--> Updating vendor dependencies"
	${GO_MOD} go get -u -v $(GOLINT) $(UNCONVERT) $(INEFFASSIGN) $(MISSPELL) $(ERRCHECK) $(UNPARAM)
	${GO_MOD} go get -u -v $(GOCILINT)

verify:
	@echo "--> Verifying dependencies have not been modified"
	${GO_MOD} go mod verify


############################
### Tools / Dependencies ###
############################

##########################################################
### TODO: Move tool depedencies to a separate makefile ###
##########################################################

GOLINT = github.com/tendermint/lint/golint
GOCILINT = github.com/golangci/golangci-lint/cmd/golangci-lint
UNCONVERT = github.com/mdempsky/unconvert
INEFFASSIGN = github.com/gordonklaus/ineffassign
MISSPELL = github.com/client9/misspell/cmd/misspell
ERRCHECK = github.com/kisielk/errcheck
UNPARAM = mvdan.cc/unparam

GOLINT_CHECK := $(shell command -v golint 2> /dev/null)
GOCILINT_CHECK := $(shell command -v golangci-lint 2> /dev/null)
UNCONVERT_CHECK := $(shell command -v unconvert 2> /dev/null)
INEFFASSIGN_CHECK := $(shell command -v ineffassign 2> /dev/null)
MISSPELL_CHECK := $(shell command -v misspell 2> /dev/null)
ERRCHECK_CHECK := $(shell command -v errcheck 2> /dev/null)
UNPARAM_CHECK := $(shell command -v unparam 2> /dev/null)


# Install the runsim binary with a temporary workaround of entering an outside
# directory as the "go get" command ignores the -mod option and will polute the
# go.{mod, sum} files.
#
# ref: https://github.com/golang/go/issues/30515
$(RUNSIM):
	@echo "Installing runsim..."
	@(cd /tmp && go get github.com/cosmos/tools/cmd/runsim@v1.0.0)

tools: $(RUNSIM)
ifdef GOLINT_CHECK
	@echo "Golint is already installed. Run 'make update-tools' to update."
else
	@echo "--> Installing golint"
	${GO_MOD} go get -v $(GOLINT)
endif
ifdef GOCILINT_CHECK
	@echo "golangci-lint is already installed. Run 'make update-tools' to update."
else
	@echo "--> Installing golangci-lint"
	${GO_MOD} go get -v $(GOCILINT)
endif
ifdef UNCONVERT_CHECK
	@echo "Unconvert is already installed. Run 'make update-tools' to update."
else
	@echo "--> Installing unconvert"
	${GO_MOD} go get -v $(UNCONVERT)
endif
ifdef INEFFASSIGN_CHECK
	@echo "Ineffassign is already installed. Run 'make update-tools' to update."
else
	@echo "--> Installing ineffassign"
	${GO_MOD} go get -v $(INEFFASSIGN)
endif
ifdef MISSPELL_CHECK
	@echo "misspell is already installed. Run 'make update-tools' to update."
else
	@echo "--> Installing misspell"
	${GO_MOD} go get -v $(MISSPELL)
endif
ifdef ERRCHECK_CHECK
	@echo "errcheck is already installed. Run 'make update-tools' to update."
else
	@echo "--> Installing errcheck"
	${GO_MOD} go get -v $(ERRCHECK)
endif
ifdef UNPARAM_CHECK
	@echo "unparam is already installed. Run 'make update-tools' to update."
else
	@echo "--> Installing unparam"
	${GO_MOD} go get -v $(UNPARAM)
endif


#######################
### Testing / Misc. ###
#######################

test: test-unit

test-unit:
	@${GO_MOD} go test -v --vet=off $(PACKAGES)

test-race:
	@${GO_MOD} go test -v --vet=off -race $(PACKAGES)

test-cli:
	@echo "NO CLI TESTS"

lint:
	@echo "--> Running golangci-lint..."
	@${GO_MOD} golangci-lint run ./... -c .golangci.yml --deadline=5m

test-import:
	@${GO_MOD} go test ./importer -v --vet=off --run=TestImportBlocks --datadir tmp \
	--blockchain blockchain --timeout=5m
	# TODO: remove tmp directory after test run to avoid subsequent errors

test-rpc:
	@${GO_MOD} go test -v --vet=off ./rpc/tester

it-tests:
	./scripts/integration-test-all.sh -q 1 -z 1 -s 10

godocs:
	@echo "--> Wait a few seconds and visit http://localhost:6060/pkg/github.com/cosmos/ethermint"
	godoc -http=:6060

docker:
	docker build -t ${DOCKER_IMAGE}:${DOCKER_TAG} .
	docker tag ${DOCKER_IMAGE}:${DOCKER_TAG} ${DOCKER_IMAGE}:latest
	docker tag ${DOCKER_IMAGE}:${DOCKER_TAG} ${DOCKER_IMAGE}:${COMMIT_HASH}

format:
	@echo "--> Formatting go files"
	@find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" | xargs gofmt -w -s
	@find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" | xargs misspell -w


.PHONY: build install update-tools tools godocs clean format lint \
test-cli test-race test-unit test test-import

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

#######################
### Simulations     ###
#######################

test-sim-nondeterminism:
	@echo "Running non-determinism test..."
	@go test -mod=readonly $(SIMAPP) -run TestAppStateDeterminism -Enabled=true \
		-NumBlocks=100 -BlockSize=200 -Commit=true -Period=0 -v -timeout 24h

test-sim-custom-genesis-fast:
	@echo "Running custom genesis simulation..."
	@echo "By default, ${HOME}/.emintd/config/genesis.json will be used."
	@go test -mod=readonly $(SIMAPP) -run TestFullAppSimulation -Genesis=${HOME}/.emintd/config/genesis.json \
		-Enabled=true -NumBlocks=100 -BlockSize=200 -Commit=true -Seed=99 -Period=5 -v -timeout 24h

test-sim-import-export: runsim
	@echo "Running Ethermint import/export simulation. This may take several minutes..."
	@$(BINDIR)/runsim -Jobs=4 -SimAppPkg=$(SIMAPP) 25 5 TestAppImportExport

test-sim-after-import: runsim
	@echo "Running Ethermint simulation-after-import. This may take several minutes..."
	@$(BINDIR)/runsim -Jobs=4 -SimAppPkg=$(SIMAPP) 25 5 TestAppSimulationAfterImport

test-sim-custom-genesis-multi-seed: runsim
	@echo "Running multi-seed custom genesis simulation..."
	@echo "By default, ${HOME}/.emintd/config/genesis.json will be used."
	@$(BINDIR)/runsim -Jobs=4 -Genesis=${HOME}/.emintd/config/genesis.json 400 5 TestFullAppSimulation

test-sim-multi-seed-long: runsim
	@echo "Running multi-seed application simulation. This may take awhile!"
	@$(BINDIR)/runsim -Jobs=4 -SimAppPkg=$(SIMAPP) 500 50 TestFullAppSimulation

test-sim-multi-seed-short: runsim
	@echo "Running multi-seed application simulation. This may take awhile!"
	@$(BINDIR)/runsim -Jobs=4 -SimAppPkg=$(SIMAPP) 50 10 TestFullAppSimulation

.PHONY: runsim test-sim-nondeterminism test-sim-custom-genesis-fast test-sim-fast sim-import-export \
	test-sim-simulation-after-import test-sim-custom-genesis-multi-seed test-sim-multi-seed \