root_dir := $(realpath ../../../../..)
artifacts_dir := ${root_dir}/tmp/contracts/${npm_package_name}/deployments/${bindings_environment}

$(info Package              ${npm_package_name})
$(info Artifacts directory  ${artifacts_dir})
$(info Required contracts   ${required_contracts})

# Go bindings generated for the solidity contracts.
contract_files = $(addprefix contract/,$(addsuffix .go,${required_contracts}))

# Files containing addresses extracted from the artifacts.
address_files = $(addprefix _address/,${required_contracts})

all: gen_address gen_contract_go

gen_address: clean check_artifacts ${address_files}

clean:
	$(info Cleaning up for ${npm_package_name})
	@rm -rf _address/*
	@rm -rf abi/*
	@rm -rf contract/*

# Check if artifacts directory exists. If the directory doesn't exists it is
# likely a problem inside a fetched contracts package.
check_artifacts:
	@[ -d "$(artifacts_dir)" ] || { echo "$(artifacts_dir) does not exist!"; exit 1; }

gen_contract_go: ${contract_files}

abi/%.abi: ${artifacts_dir}/%.json
	$(info $* - generating ABI)
	@jq .abi $< > abi/$*.abi

abi/%.go: abi/%.abi
	$(info $* - generating Ethereum bindings)
	@go run github.com/ethereum/go-ethereum/cmd/abigen --abi $< --pkg abi --type $* --out $@

_address/%: ${artifacts_dir}/%.json
	$(info $* - extracting address)
	@jq -jr .address ${artifacts_dir}/$*.json > _address/$*

contract/%.go: abi/%.abi abi/%.go _address/% ${artifacts_dir}/%.json
	$(info $* - generating Mezo bindings)
	@go run github.com/keep-network/keep-common/tools/generators/ethereum $< contract/$*.go

# Don't remove intermediate files that got generated.
.PRECIOUS: abi/%.abi abi/%.go _address/%
