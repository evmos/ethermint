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

PACKAGES=$(shell go list ./... | grep -Ev 'vendor|importer')
COMMIT_HASH := $(shell git rev-parse --short HEAD)
BUILD_FLAGS = -tags netgo -ldflags "-X github.com/cosmos/ethermint/version.GitCommit=${COMMIT_HASH}"
DOCKER_TAG = unstable
DOCKER_IMAGE = cosmos/ethermint
ETHERMINT_DAEMON_BINARY = emintd
ETHERMINT_CLI_BINARY = emintcli
GO_MOD=GO111MODULE=on

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
	${GO_MOD} go get -v $(GOCILINT)

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
GOCILINT = github.com/golangci/golangci-lint/cmd/golangci-lint@v1.17.1
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

tools:
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

test-lint:
	@echo "--> Running golangci-lint..."
	@${GO_MOD} golangci-lint run --deadline=5m ./...

test-import:
	@${GO_MOD} go test ./importer -v --vet=off --run=TestImportBlocks --datadir tmp \
	--blockchain blockchain --timeout=5m
	# TODO: remove tmp directory after test run to avoid subsequent errors

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

.PHONY: build install update-tools tools godocs clean format test-lint \
test-cli test-race test-unit test test-import
