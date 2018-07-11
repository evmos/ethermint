PACKAGES=$(shell go list ./... | grep -v '/vendor/')
COMMIT_HASH := $(shell git rev-parse --short HEAD)
BUILD_FLAGS = -tags netgo -ldflags "-X github.com/cosmos/ethermint/version.GitCommit=${COMMIT_HASH}"

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

all: get-tools get-vendor-deps install

ci: get-tools get-vendor-deps install

build:
ifeq ($(OS),Windows_NT)
	go build $(BUILD_FLAGS) -o build/ethermint.exe ./*.go
else
	go build $(BUILD_FLAGS) -o build/ethermint ./*.go
endif

install:
	go install $(BUILD_FLAGS) ./*.go

clean:
	rm -rf ./build ./vendor

update-tools:
	@echo "Updating golang dependencies"
	go get -u -v $(DEP) $(GOLINT) $(GOMETALINTER) $(UNCONVERT) $(INEFFASSIGN) $(MISSPELL) $(ERRCHECK) $(UNPARAM) $(GOCYCLO)

get-tools:
	ifdef DEP_CHECK
		@echo "Dep is already installed.  Run 'make update-tools' to update."
	else
		@echo "Installing dep"
		go get -v $(DEP)
	endif
	ifdef GOLINT_CHECK
		@echo "Golint is already installed.  Run 'make update-tools' to update."
	else
		@echo "Installing golint"
		go get -v $(GOLINT)
	endif
	ifdef GOMETALINTER_CHECK
		@echo "Gometalinter.v2 is already installed.  Run 'make update-tools' to update."
	else
		@echo "Installing gometalinter.v2"
		go get -v $(GOMETALINTER)
	endif
	ifdef UNCONVERT_CHECK
		@echo "Unconvert is already installed.  Run 'make update-tools' to update."
	else
		@echo "Installing unconvert"
		go get -v $(UNCONVERT)
	endif
	ifdef INEFFASSIGN_CHECK
		@echo "Ineffassign is already installed.  Run 'make update-tools' to update."
	else
		@echo "Installing ineffassign"
		go get -v $(INEFFASSIGN)
	endif
	ifdef MISSPELL_CHECK
		@echo "misspell is already installed.  Run 'make update-tools' to update."
	else
		@echo "Installing misspell"
		go get -v $(MISSPELL)
	endif
	ifdef ERRCHECK_CHECK
		@echo "errcheck is already installed.  Run 'make update-tools' to update."
	else
		@echo "Installing errcheck"
		go get -v $(ERRCHECK)
	endif
	ifdef UNPARAM_CHECK
		@echo "unparam is already installed.  Run 'make update-tools' to update."
	else
		@echo "Installing unparam"
		go get -v $(UNPARAM)
	endif
	ifdef GOYCLO_CHECK
		@echo "goyclo is already installed.  Run 'make update-tools' to update."
	else
		@echo "Installing goyclo"
		go get -v $(GOCYCLO)
	endif

get-vendor-deps:
	@rm -rf vendor/
	@echo "--> Running dep ensure"
	@dep ensure -v

godocs:
	@echo "--> Wait a few seconds and visit http://localhost:6060/pkg/github.com/cosmos/ethermint"
	godoc -http=:6060

.PHONY: build install update-tools get-tools get-vendor-deps godocs
