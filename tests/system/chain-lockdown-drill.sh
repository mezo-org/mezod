#!/usr/bin/env bash
#
# ############################################################################
# # WARNING: THIS SCRIPT KILLS THE LOCALNODE.                                #
# #                                                                          #
# # The drill calls setChainLockdown on the maintenance precompile. That     #
# # schedules an upgrade plan with no registered handler at the next block   #
# # height. The x/upgrade PreBlocker then panics and the node stops. The     #
# # chain cannot recover on chain. The localnode stays down until you follow #
# # one of the recovery paths printed at the end of the drill.               #
# #                                                                          #
# # Never point this script at a shared network. Use a throwaway localnode.  #
# ############################################################################
#
# The halt name is the first major version above the last completed upgrade
# that has no registered upgrade handler in the running binary. A release
# binary registers no future handler, so the halt name is the next version. A
# dev binary registers every historical handler, so the halt name skips past
# them. Do not assume a fixed name. The drill reads it from the
# ChainLockdownSet event.
#
# The drill is destructive, so it is NOT part of system-tests.sh. Run it by
# hand:
#
#   ./tests/system/chain-lockdown-drill.sh
#
# Steps:
#   1. Check that a localnode serves JSON-RPC.
#   2. Grant the Emergency Team role to dev1. The PoA owner dev0 does it.
#   3. Call setChainLockdown as dev1.
#   4. Assert the ChainLockdownSet event and capture the derived halt name.
#   5. Assert that the node halts: the log shows UPGRADE "<name>" NEEDED and
#      the JSON-RPC height stops advancing.
#   6. Print both recovery paths.
#
# Environment variables:
#   RPC_URL       - JSON-RPC endpoint (default: http://127.0.0.1:8545)
#   LOCALNODE_DIR - localnode home with the dev key seeds
#                   (default: <repo>/.localnode)
#   MEZOD_LOG     - file with the localnode stdout. The log assertion runs
#                   only when this file is available. The localnode logs to
#                   stdout, so capture it, for example:
#                   make localnode-bin-start 2>&1 | tee /tmp/mezod.log
#   HALT_TIMEOUT  - seconds to wait for the halt (default: 90)
#   STALL_WAIT    - seconds to watch a still-reachable node for new blocks
#                   (default: 20)
#   EXPECT_HALT_NAME - the halt name the drill asserts. It is empty by
#                   default, so the drill accepts any version and uses the
#                   name from the ChainLockdownSet event. Set it only when you
#                   know which name the node must derive.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

RPC_URL="${RPC_URL:-http://127.0.0.1:8545}"
LOCALNODE_DIR="${LOCALNODE_DIR:-$REPO_DIR/.localnode}"
HALT_TIMEOUT="${HALT_TIMEOUT:-90}"
STALL_WAIT="${STALL_WAIT:-20}"
EXPECT_HALT_NAME="${EXPECT_HALT_NAME-}"

WORK_DIR="$(mktemp -d)"
trap 'rm -rf "$WORK_DIR"' EXIT

log() {
	printf '[drill] %s\n' "$1"
}

fail() {
	printf '[drill] FAILED: %s\n' "$1" >&2
	exit 1
}

# rpc_height prints the current block height as a decimal number. It fails if
# the endpoint does not answer.
rpc_height() {
	local response hex
	response="$(
		curl -s -m 5 -X POST -H 'Content-Type: application/json' \
			--data '{"jsonrpc":"2.0","id":1,"method":"eth_blockNumber","params":[]}' \
			"$RPC_URL"
	)" || return 1

	hex="$(printf '%s' "$response" | sed -n 's/.*"result":"0x\([0-9a-fA-F]*\)".*/\1/p')"
	[ -n "$hex" ] || return 1

	printf '%d\n' "$((16#$hex))"
}

for tool in curl node npm; do
	command -v "$tool" >/dev/null 2>&1 || fail "$tool is not installed"
done

log "checking the localnode on $RPC_URL"
START_HEIGHT="$(rpc_height)" || fail "no JSON-RPC on $RPC_URL, start the localnode first"
log "the localnode is at height $START_HEIGHT"

[ -f "$LOCALNODE_DIR/dev0_key_seed.json" ] ||
	fail "no dev0 key seed in $LOCALNODE_DIR"
[ -f "$LOCALNODE_DIR/dev1_key_seed.json" ] ||
	fail "no dev1 key seed in $LOCALNODE_DIR"

if [ ! -d "$SCRIPT_DIR/node_modules/ethers" ]; then
	log "installing the JavaScript dependencies"
	(cd "$SCRIPT_DIR" && npm i --silent)
fi

cat >"$WORK_DIR/drill.js" <<'NODE_SCRIPT'
const fs = require("fs")
const path = require("path")

const systemTestsDir = process.env.SYSTEM_TESTS_DIR
const localnodeDir = process.env.LOCALNODE_DIR
const rpcUrl = process.env.RPC_URL

const { ethers } = require(path.join(systemTestsDir, "node_modules", "ethers"))

const maintenanceAddress = "0x7b7c000000000000000000000000000000000013"

const maintenanceAbi = [
  "function setEmergencyTeam(address team) external returns (bool)",
  "function getEmergencyTeam() external view returns (address team)",
  "function setChainLockdown() external returns (bool)",
  "event ChainLockdownSet(string name)",
]

// The gas limit is explicit. Gas estimation for setChainLockdown would be one
// more call against a chain that is about to stop.
const gasLimit = 1000000

// devWallet derives the same key as hardhat.config.ts: the seed phrase of the
// localnode key, on the default Ethereum HD path.
function devWallet(index, provider) {
  const file = path.join(localnodeDir, `dev${index}_key_seed.json`)
  const seed = JSON.parse(fs.readFileSync(file, "utf8"))
  const privateKey = ethers.Wallet.fromPhrase(seed.secret).privateKey

  return new ethers.Wallet(privateKey, provider)
}

// waitForReceipt polls for the receipt. The node stops one block after the
// lockdown transaction, so the receipt must be read in that short window.
async function waitForReceipt(provider, hash, timeoutMs) {
  const deadline = Date.now() + timeoutMs

  while (Date.now() < deadline) {
    try {
      const receipt = await provider.getTransactionReceipt(hash)
      if (receipt) {
        return receipt
      }
    } catch (err) {
      throw new Error(
        `cannot read the receipt of ${hash}, the node may already be down: ${err.message}`
      )
    }

    await new Promise((resolve) => setTimeout(resolve, 250))
  }

  throw new Error(`timed out waiting for the receipt of ${hash}`)
}

async function main() {
  const provider = new ethers.JsonRpcProvider(rpcUrl)
  provider.pollingInterval = 250

  const owner = devWallet(0, provider)
  const team = devWallet(1, provider)

  const asOwner = new ethers.Contract(maintenanceAddress, maintenanceAbi, owner)
  const asTeam = new ethers.Contract(maintenanceAddress, maintenanceAbi, team)

  console.error(`[drill] granting the Emergency Team role to ${team.address}`)
  const grantTx = await asOwner.setEmergencyTeam(team.address, { gasLimit })
  const grantReceipt = await grantTx.wait()
  if (grantReceipt.status !== 1) {
    throw new Error(`setEmergencyTeam reverted in ${grantTx.hash}`)
  }

  const currentTeam = await asOwner.getEmergencyTeam()
  if (currentTeam.toLowerCase() !== team.address.toLowerCase()) {
    throw new Error(
      `the Emergency Team is ${currentTeam}, expected ${team.address}`
    )
  }

  console.error("[drill] calling setChainLockdown as the Emergency Team")
  const lockdownTx = await asTeam.setChainLockdown({ gasLimit })
  console.error(`[drill] lockdown transaction ${lockdownTx.hash}`)

  const receipt = await waitForReceipt(provider, lockdownTx.hash, 30000)
  if (receipt.status !== 1) {
    throw new Error(`setChainLockdown reverted in ${lockdownTx.hash}`)
  }

  const iface = new ethers.Interface(maintenanceAbi)
  const events = receipt.logs
    .filter(
      (entry) =>
        entry.address.toLowerCase() === maintenanceAddress.toLowerCase()
    )
    .map((entry) => {
      try {
        return iface.parseLog(entry)
      } catch {
        return null
      }
    })
    .filter((parsed) => parsed !== null && parsed.name === "ChainLockdownSet")

  if (events.length !== 1) {
    throw new Error(
      `expected a single ChainLockdownSet event, got ${events.length}`
    )
  }

  const haltName = events[0].args.name
  if (!/^v\d+\.\d+\.\d+$/.test(haltName)) {
    throw new Error(`the halt name ${haltName} is not a version`)
  }

  // The plan sits one block after the lockdown transaction.
  console.log(`HALT_NAME=${haltName}`)
  console.log(`HALT_HEIGHT=${receipt.blockNumber + 1}`)
  console.log(`LOCKDOWN_TX=${lockdownTx.hash}`)
}

main().catch((err) => {
  console.error(`[drill] ${err.message}`)
  process.exit(1)
})
NODE_SCRIPT

log "running the lockdown transactions"
SYSTEM_TESTS_DIR="$SCRIPT_DIR" \
	LOCALNODE_DIR="$LOCALNODE_DIR" \
	RPC_URL="$RPC_URL" \
	node "$WORK_DIR/drill.js" >"$WORK_DIR/result.env" ||
	fail "the lockdown transactions did not go through"

# shellcheck disable=SC1091
. "$WORK_DIR/result.env"

[ -n "${HALT_NAME:-}" ] || fail "the drill did not report a halt name"
[ -n "${HALT_HEIGHT:-}" ] || fail "the drill did not report a halt height"

if [ -n "$EXPECT_HALT_NAME" ] && [ "$HALT_NAME" != "$EXPECT_HALT_NAME" ]; then
	fail "the halt name is $HALT_NAME, expected $EXPECT_HALT_NAME"
fi

log "the ChainLockdownSet event carries the halt name $HALT_NAME"
log "the halt is scheduled for height $HALT_HEIGHT"

LOG_MARKER="UPGRADE \"$HALT_NAME\" NEEDED"

log "waiting up to ${HALT_TIMEOUT}s for the halt"
log_seen=0
rpc_down=0
deadline=$(($(date +%s) + HALT_TIMEOUT))
while [ "$(date +%s)" -lt "$deadline" ]; do
	if [ -n "${MEZOD_LOG:-}" ] && [ -f "$MEZOD_LOG" ] &&
		grep -q "$LOG_MARKER" "$MEZOD_LOG"; then
		log_seen=1
	fi

	if ! rpc_height >/dev/null 2>&1; then
		rpc_down=1
	fi

	if [ "$log_seen" -eq 1 ] || [ "$rpc_down" -eq 1 ]; then
		break
	fi

	sleep 2
done

if [ -n "${MEZOD_LOG:-}" ]; then
	# The node panics and then exits, so the loop above can break on the dead
	# JSON-RPC endpoint before the marker reaches the log file. Read the log
	# once more.
	if [ "$log_seen" -eq 0 ] && [ -f "$MEZOD_LOG" ]; then
		sleep 5
		if grep -q "$LOG_MARKER" "$MEZOD_LOG"; then
			log_seen=1
		fi
	fi

	[ "$log_seen" -eq 1 ] ||
		fail "the log $MEZOD_LOG does not contain: $LOG_MARKER"
	log "the node log contains: $LOG_MARKER"
else
	log "WARNING: MEZOD_LOG is not set, skipping the node log assertion"
fi

if [ "$rpc_down" -eq 1 ]; then
	log "the JSON-RPC endpoint is gone, the node process stopped"
else
	first_height="$(rpc_height)" || fail "cannot read the height"
	log "watching the height $first_height for ${STALL_WAIT}s"
	sleep "$STALL_WAIT"
	second_height="$(rpc_height)" || {
		log "the JSON-RPC endpoint is gone, the node process stopped"
		second_height="$first_height"
	}

	[ "$second_height" -eq "$first_height" ] ||
		fail "the height advanced from $first_height to $second_height, the chain did not halt"
	log "the height stayed at $first_height, the chain halted"
fi

cat <<EOF

[drill] PASSED. The chain is down at the halt name $HALT_NAME.

Recovery path 1, the primary one. Ship a release whose upgrade handler
registers exactly $HALT_NAME. Note the exact name: a release named
$HALT_NAME resolves the halt, a patch release does not. Validators install
the new binary and restart. The halt then resolves as a normal upgrade.

Recovery path 2, the fallback. Validators restart the same binary and skip
the plan:

  mezod start --unsafe-skip-upgrades $HALT_HEIGHT [other flags] --home $LOCALNODE_DIR

The chain resumes with the state it had before the halt.
EOF
