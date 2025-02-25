#!/usr/bin/make -f

PACKAGES_NOSIMULATION=$(shell go list ./... | grep -v '/simulation')
VERSION ?= $(shell echo $(shell git describe --tags --always) | sed 's/^v//')
CMTVERSION := $(shell go list -m github.com/cometbft/cometbft | sed 's:.* ::')
COMMIT := $(shell git log -1 --format='%H')
LEDGER_ENABLED ?= true
BINDIR ?= $(GOPATH)/bin
MEZO_BINARY = mezod
BUILDDIR ?= $(CURDIR)/build
DOCKER := $(shell which docker)
NAMESPACE := mezo-org
PROJECT := mezod
DOCKER_IMAGE := $(NAMESPACE)/$(PROJECT)
COMMIT_HASH := $(shell git rev-parse --short=7 HEAD)
DOCKER_TAG := $(COMMIT_HASH)
DEBUG_BUILD_ENABLED ?= false

export GO111MODULE = on

# Default target executed when no arguments are given to make.
default_target: all

.PHONY: default_target

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

ifeq (cleveldb,$(findstring cleveldb,$(COSMOS_BUILD_OPTIONS)))
  build_tags += gcc
endif
build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

# process linker flags

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=mezo \
          -X github.com/cosmos/cosmos-sdk/version.AppName=$(MEZO_BINARY) \
          -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
          -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
          -X github.com/cometbft/cometbft/version.CMTSemVer=$(CMTVERSION) \
          -X github.com/mezo-org/mezod/version.AppVersion=$(VERSION)

# DB backend selection
ifeq (cleveldb,$(findstring cleveldb,$(COSMOS_BUILD_OPTIONS)))
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=cleveldb
endif
ifeq (badgerdb,$(findstring badgerdb,$(COSMOS_BUILD_OPTIONS)))
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=badgerdb
endif
# handle rocksdb
ifeq (rocksdb,$(findstring rocksdb,$(COSMOS_BUILD_OPTIONS)))
  CGO_ENABLED=1
  build_tags += rocksdb
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=rocksdb
endif
# handle boltdb
ifeq (boltdb,$(findstring boltdb,$(COSMOS_BUILD_OPTIONS)))
  build_tags += boltdb
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=boltdb
endif

ifeq ($(DEBUG_BUILD_ENABLED),true)
  build_tags += debugprecompile
endif

# add build tags to linker flags
whitespace := $(subst ,, )
comma := ,
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))
ldflags += -X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)"

ifeq (,$(findstring nostrip,$(COSMOS_BUILD_OPTIONS)))
  ldflags += -w -s
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'
# check for nostrip option
ifeq (,$(findstring nostrip,$(COSMOS_BUILD_OPTIONS)))
  BUILD_FLAGS += -trimpath
endif

# check if no optimization option is passed
# used for remote debugging
ifneq (,$(findstring nooptimization,$(COSMOS_BUILD_OPTIONS)))
  BUILD_FLAGS += -gcflags "all=-N -l"
endif

###############################################################################
###                                  Build                                  ###
###############################################################################

BUILD_TARGETS := build install

build: clean
  BUILD_ARGS=-o $(BUILDDIR)/

# Set empty BUILD_ARGS for install. By default, BUILD_ARGS contain the -o
# flag that is not supported by go install.
install: BUILD_ARGS=

$(BUILD_TARGETS): go.sum $(BUILDDIR)/
	go $@ $(BUILD_FLAGS) $(BUILD_ARGS) ./...

$(BUILDDIR)/:
	mkdir -p $(BUILDDIR)/

build-docker:
	# TODO replace with kaniko
	$(DOCKER) build -t ${DOCKER_IMAGE}:${DOCKER_TAG} .
	$(DOCKER) tag ${DOCKER_IMAGE}:${DOCKER_TAG} ${DOCKER_IMAGE}:latest
	# docker tag ${DOCKER_IMAGE}:${DOCKER_TAG} ${DOCKER_IMAGE}:${COMMIT_HASH}
	# update old container
	$(DOCKER) rm mezo || true
	# create a new container from the latest image
	$(DOCKER) create --name mezo -t -i ${DOCKER_IMAGE}:latest mezo
	# move the binaries to the ./build directory
	mkdir -p ./build/
	$(DOCKER) cp mezo:/usr/bin/mezod ./build/

build-docker-linux:
	$(DOCKER) buildx build --platform linux/amd64 --tag ${DOCKER_IMAGE}:${DOCKER_TAG} .

$(MOCKS_DIR):
	mkdir -p $(MOCKS_DIR)

clean:
	rm -rf \
    $(BUILDDIR)/ \
    artifacts/

all: build

.PHONY: clean

###############################################################################
###                              Dependencies                               ###
###############################################################################

go.sum: go.mod
	echo "Ensure dependencies have not been modified ..." >&2
	go mod verify
	go mod tidy

vulncheck: $(BUILDDIR)/
	GOBIN=$(BUILDDIR) go install golang.org/x/vuln/cmd/govulncheck@latest
	$(BUILDDIR)/govulncheck ./...

###############################################################################
###                           Tests & Simulation                            ###
###############################################################################

test: test-unit
test-all: test-unit test-race
# For unit tests we don't want to execute the upgrade tests in tests/e2e but
# we want to include all unit tests in the subfolders (tests/e2e/*)
PACKAGES_UNIT=$(shell go list ./... | grep -v '/tests/e2e$$')
TEST_PACKAGES=./...
TEST_TARGETS := test-unit test-unit-cover test-race

# Test runs-specific rules. To add a new test target, just add
# a new rule, customise ARGS or TEST_PACKAGES ad libitum, and
# append the new rule to the TEST_TARGETS list.
test-unit: ARGS=-timeout=15m -race
test-unit: TEST_PACKAGES=$(PACKAGES_UNIT)

test-race: ARGS=-race
test-race: TEST_PACKAGES=$(PACKAGES_NOSIMULATION)
$(TEST_TARGETS): run-tests

test-unit-cover: ARGS=-timeout=15m -race -coverprofile=coverage.txt -covermode=atomic
test-unit-cover: TEST_PACKAGES=$(PACKAGES_UNIT)

run-tests:
ifneq (,$(shell which tparse 2>/dev/null))
	go test -mod=readonly -json $(ARGS) $(EXTRA_ARGS) $(TEST_PACKAGES) | tparse
else
	go test -mod=readonly $(ARGS)  $(EXTRA_ARGS) $(TEST_PACKAGES)
endif

.PHONY: run-tests test test-all $(TEST_TARGETS)

benchmark:
	@go test -mod=readonly -bench=. $(PACKAGES_NOSIMULATION)
.PHONY: benchmark

###############################################################################
###                                Linting                                  ###
###############################################################################

lint:
	golangci-lint run --out-format=tab

lint-fix:
	golangci-lint run --fix --out-format=tab --issues-exit-code=0

.PHONY: lint lint-fix

format:
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/docs/statik/statik.go" -not -name '*.pb.go' | xargs gofumpt -w -l

.PHONY: format

###############################################################################
###                                Protobuf                                 ###
###############################################################################

protoVer=0.11.5
protoImageName=ghcr.io/cosmos/proto-builder:$(protoVer)
protoImage=$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace --user 0 $(protoImageName)

# ------
# NOTE: If you are experiencing problems running these commands, try deleting
#       the docker images and execute the desired command again.
#
proto-all: proto-format proto-lint proto-gen

proto-gen:
	@echo "Generating Protobuf files"
	$(protoImage) sh ./scripts/protocgen.sh

proto-format:
	@echo "Formatting Protobuf files"
	$(protoImage) find ./ -name *.proto -exec clang-format -i {} \;

proto-lint:
	@echo "Linting Protobuf files"
	$(protoImage) buf lint --error-format=json

.PHONY: proto-all proto-gen proto-format proto-lint

###############################################################################
###                         Localnet binary-based                           ###
###############################################################################

LOCALNET_DIR = .localnet
LOCALNET_CHAIN_ID = mezo_31611-10
# LOCALNET_ASSETS_LOCKED_SEQUENCE_TIP is set to the sequence tip the
# MezoBridge contract on Sepolia was initialized with. This ensures the
# localnet can start bridging from the first AssetLocked event emitted
# by the MezoBridge contract.
LOCALNET_ASSETS_LOCKED_SEQUENCE_TIP = 21061
# LOCALNET_SOURCE_BTC_TOKEN is the TBTC on Ethereum Sepolia.
LOCALNET_SOURCE_BTC_TOKEN = 0x517f2982701695D4E52f1ECFBEf3ba31Df470161

localnet-bin-init:
	@if ! [ -d build ]; then \
		echo "Build directory not found. Running build..."; \
		make build; \
	fi
	@if ! [ -d $(LOCALNET_DIR) ]; then \
		echo "Initializing localnet configuration..."; \
		./build/mezod testnet init-files \
		--v 4 \
		--output-dir $(LOCALNET_DIR) \
		--home $(LOCALNET_DIR) \
		--keyring-backend=test \
		--starting-ip-address localhost \
		--chain-id $(LOCALNET_CHAIN_ID) \
		--assets-locked-sequence-tip=$(LOCALNET_ASSETS_LOCKED_SEQUENCE_TIP) \
		--source-btc-token=$(LOCALNET_SOURCE_BTC_TOKEN); \
	else \
		echo "Skipped initializing localnet configuration."; \
	fi

localnet-bin-start:
	LOCALNET_CHAIN_ID=$(LOCALNET_CHAIN_ID) ./scripts/localnet-start.sh

localnet-bin-sidecars-start:
	./scripts/localnet-sidecars-start.sh

localnet-bin-clean:
	rm -rf $(LOCALNET_DIR) build

.PHONY: localnet-bin-init localnet-bin-start localnet-bin-sidecars-start localnet-bin-clean

###############################################################################
###                         Local node binary-based                         ###
###############################################################################

localnode-bin-start:
	./scripts/localnode-start.sh

.PHONY: localnode-bin-start

###############################################################################
###                       Contract bindings generation                      ###
###############################################################################

# bindings_environment determines the network type that should be used for contract
# binding generation. The default value is mainnet.
ifndef bindings_environment
# TODO: Once we are production ready, this has to be changed to mainnet.
override bindings_environment = sepolia
endif

export bindings_environment

# List of NPM packages for which to generate bindings - expand if needed.
npm_packages := @mezo-org/contracts

# Working directory where contracts artifacts should be stored.
contracts_dir := tmp/contracts

# It requires npm of at least 7.x version to support `pack-destination` flag.
define get_npm_package
$(info Fetching package $(1))
$(eval destination_dir := ${contracts_dir}/$(1))
@rm -rf ${destination_dir} && mkdir -p ${destination_dir}
@tarball=$$(npm pack --silent --pack-destination=${destination_dir} $(1)); \
tar -zxf ${destination_dir}/$${tarball} -C ${destination_dir} --strip-components 1 package/deployments/
$(info Downloaded NPM package $(1) to ${contracts_dir})
endef

get_npm_packages:
	$(foreach pkg,$(npm_packages),$(call get_npm_package,$(pkg)))

generate:
	# go generate needs some go.sum dependencies that are not actually used
    # in the codebase and are pruned by go mod tidy. As a workaround, we
    # temporarily set GOFLAGS=-mod=mod to let go generate fetch necessary
    # dependencies.
	go env -w GOFLAGS=-mod=mod
	go generate ./...
	# Reset GOFLAGS to its original value and run go mod tidy to remove
	# unnecessary dependencies fetched by go generate.
	go env -u GOFLAGS
	go mod tidy

bindings: get_npm_packages generate
	$(info Bindings generated)
