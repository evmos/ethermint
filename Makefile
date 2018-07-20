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

PACKAGES=$(shell go list ./... | grep -v '/vendor/')
COMMIT_HASH := $(shell git rev-parse --short HEAD)
BUILD_FLAGS = -tags netgo -ldflags "-X github.com/cosmos/ethermint/version.GitCommit=${COMMIT_HASH}"
DOCKER_TAG = unstable
DOCKER_IMAGE = tendermint/ethermint
ETHERMINT_DAEMON_BINARY = emintd
ETHERMINT_CLI_BINARY = emintcli

DEP = github.com/golang/dep/cmd/dep
GOLINT = github.com/tendermint/lint/golint
GOMETALINTER = gopkg.in/alecthomas/gometalinter.v2
UNCONVERT = github.com/mdempsky/unconvert
INEFFASSIGN = github.com/gordonklaus/ineffassign
MISSPELL = github.com/client9/misspell/cmd/misspell
ERRCHECK = github.com/kisielk/errcheck
UNPARAM = mvdan.cc/unparam
GOCYCLO = github.com/alecthomas/gocyclo

DEP_CHECK := $(shell command -v dep 2> /dev/null)
GOLINT_CHECK := $(shell command -v golint 2> /dev/null)
GOMETALINTER_CHECK := $(shell command -v gometalinter.v2 2> /dev/null)
UNCONVERT_CHECK := $(shell command -v unconvert 2> /dev/null)
INEFFASSIGN_CHECK := $(shell command -v ineffassign 2> /dev/null)
MISSPELL_CHECK := $(shell command -v misspell 2> /dev/null)
ERRCHECK_CHECK := $(shell command -v errcheck 2> /dev/null)
UNPARAM_CHECK := $(shell command -v unparam 2> /dev/null)
GOCYCLO_CHECK := $(shell command -v gocyclo 2> /dev/null)

all: tools deps install

#######################
### Build / Install ###
#######################

build:
ifeq ($(OS),Windows_NT)
	go build $(BUILD_FLAGS) -o build/$(ETHERMINT_DAEMON_BINARY).exe ./cmd/ethermintd
	go build $(BUILD_FLAGS) -o build/$(ETHERMINT_CLI_BINARY).exe ./cmd/ethermintcli
else
	go build $(BUILD_FLAGS) -o build/$(ETHERMINT_DAEMON_BINARY) ./cmd/ethermintd/
	go build $(BUILD_FLAGS) -o build/$(ETHERMINT_CLI_BINARY) ./cmd/ethermintcli/
endif

install:
	go install $(BUILD_FLAGS) ./cmd/ethermintd
	go install $(BUILD_FLAGS) ./cmd/ethermintcli

clean:
	rm -rf ./build ./vendor

update-tools:
	@echo "Updating golang dependencies"
	go get -u -v $(DEP) $(GOLINT) $(GOMETALINTER) $(UNCONVERT) $(INEFFASSIGN) $(MISSPELL) $(ERRCHECK) $(UNPARAM) $(GOCYCLO)

############################
### Tools / Dependencies ###
############################

tools:
ifdef DEP_CHECK
	@echo "Dep is already installed. Run 'make update-tools' to update."
else
	@echo "Installing dep"
	go get -v $(DEP)
endif
ifdef GOLINT_CHECK
	@echo "Golint is already installed. Run 'make update-tools' to update."
else
	@echo "Installing golint"
	go get -v $(GOLINT)
endif
ifdef GOMETALINTER_CHECK
	@echo "Gometalinter.v2 is already installed. Run 'make update-tools' to update."
else
	@echo "Installing gometalinter.v2"
	go get -v $(GOMETALINTER)
endif
ifdef UNCONVERT_CHECK
	@echo "Unconvert is already installed. Run 'make update-tools' to update."
else
	@echo "Installing unconvert"
	go get -v $(UNCONVERT)
endif
ifdef INEFFASSIGN_CHECK
	@echo "Ineffassign is already installed. Run 'make update-tools' to update."
else
	@echo "Installing ineffassign"
	go get -v $(INEFFASSIGN)
endif
ifdef MISSPELL_CHECK
	@echo "misspell is already installed. Run 'make update-tools' to update."
else
	@echo "Installing misspell"
	go get -v $(MISSPELL)
endif
ifdef ERRCHECK_CHECK
	@echo "errcheck is already installed. Run 'make update-tools' to update."
else
	@echo "Installing errcheck"
	go get -v $(ERRCHECK)
endif
ifdef UNPARAM_CHECK
	@echo "unparam is already installed. Run 'make update-tools' to update."
else
	@echo "Installing unparam"
	go get -v $(UNPARAM)
endif
ifdef GOCYCLO_CHECK
	@echo "goyclo is already installed. Run 'make update-tools' to update."
else
	@echo "Installing goyclo"
	go get -v $(GOCYCLO)
endif

deps:
	@rm -rf vendor/
	@echo "--> Running dep ensure"
	@dep ensure -v

#######################
### Testing / Misc. ###
#######################

godocs:
	@echo "--> Wait a few seconds and visit http://localhost:6060/pkg/github.com/cosmos/ethermint"
	godoc -http=:6060

docker:
	docker build -t ${DOCKER_IMAGE}:${DOCKER_TAG} .
	docker tag ${DOCKER_IMAGE}:${DOCKER_TAG} ${DOCKER_IMAGE}:latest
	docker tag ${DOCKER_IMAGE}:${DOCKER_TAG} ${DOCKER_IMAGE}:${COMMIT_HASH}

.PHONY: build install update-tools tools deps godocs clean
