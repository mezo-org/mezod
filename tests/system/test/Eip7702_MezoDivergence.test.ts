import { expect } from "chai"
import hre, { ethers } from "hardhat"
import { getAddress } from "ethers"

import {
  EMPTY_CODE,
  expectedDelegationCode,
  freshAuthority,
  freshSponsor,
  sendSetCodeTx,
  signAuthorization,
} from "./helpers/eip7702"

/**
 * Eip7702_MezoDivergence pins mezod's intentional deviations from the
 * EIP-7702 reference implementation (go-ethereum Prague). If one of
 * these starts failing, mezod has drifted toward the spec — investigate
 * before flipping the assertion to make it green again. The companion
 * suite Eip7702_SpecCompliance pins the EIP-conformant surface; the two
 * files are mutually exclusive — never duplicated.
 *
 * Conventions: each it() builds its own fresh sponsor and authorities
 * so cases are independent and order-agnostic. Receipts are polled
 * (HardhatEthersProvider doesn't implement waitForTransaction).
 *
 * Test scenarios:
 *
 * | Scenario                                             | Given                                     | When                                                | Then                                                                             |
 * |------------------------------------------------------|-------------------------------------------|-----------------------------------------------------|----------------------------------------------------------------------------------|
 * | two sequential txs from same authority accepted      | no mempool authority reservation in mezod | two type-0x04 txs from same authority, back-to-back | both included; final delegation is the second target's; authority nonce = 2      |
 * | two auths from same authority in one tx accepted     | both tuples target different addresses    | single sendSetCodeTx with [auth(T1), auth(T2)]      | receipt 0x1; final delegation is the second target's; authority nonce = 2        |
 * | precompile-targeted tuples rejected                  | stock 0x01 + every Mezo custom 0x7b7c...  | mixed tx with precompile tuples plus one real tuple | precompile tuples silently skipped (no code, no nonce bump); real tuple installs |
 * | sender + auth recovery via PragueSigner (no adapter) | no Mezo-specific signer wrapper           | sendSetCodeTx and read rpcTx fields                 | rpcTx.from = sponsor; auth.address round-trips; code = designator(T1)            |
 */

// Stock geth precompile 0x01 (ecRecover). Holds no stored bytecode on
// upstream, so a delegation pointing here is a no-op runtime; mezod
// rejects it at validate time.
const STOCK_PRECOMPILE = "0x0000000000000000000000000000000000000001"

// Mezo custom precompile addresses. These have facade bytecode in
// genesis, so an unguarded delegation pointing here would execute the
// precompile's surface in the authority's storage. Mezo rejects all of
// them at validate time.
//
// KEEP THIS LIST IN SYNC WITH x/evm/types/precompile.go. TestBed
// (0x7b7c1…) is intentionally excluded: its registration is
// environment-dependent (gated on chainID == "mezo_31611-10" AND the
// --enable-testbed-precompile flag in app/app.go), so it's loaded in
// localnode but not in localnet / testnet / mainnet builds — asserting
// rejection of a tuple targeting it would couple this test to the
// localnode-only test scaffolding.
const MEZO_CUSTOM_PRECOMPILES = [
  "0x7b7c000000000000000000000000000000000000", // BTCToken
  "0x7b7c000000000000000000000000000000000001", // MEZOToken
  "0x7b7c000000000000000000000000000000000011", // ValidatorPool
  "0x7b7c000000000000000000000000000000000012", // AssetsBridge
  "0x7b7c000000000000000000000000000000000013", // Maintenance
  "0x7b7c000000000000000000000000000000000014", // Upgrade
  "0x7b7c000000000000000000000000000000000015", // PriceOracle
]

describe("Eip7702_MezoDivergence", function () {
  const { deployments } = hre

  let targetAddr: string
  let target2Addr: string
  let chainId: bigint

  before(async function () {
    await deployments.fixture(["eip7702-fixtures"])
    targetAddr = (await deployments.get("Eip7702TargetV1")).address
    target2Addr = (await deployments.get("Eip7702TargetV2")).address
    chainId = (await ethers.provider.getNetwork()).chainId
  })

  it("accepts two sequential type-0x04 txs from same authority (no mempool reservation)", async function () {
    // Divergence #1, surface (a): cross-tx authority reservation.
    // Geth's txpool rejects a second pending type-0x04 tx whose author
    // matches one already in flight. Mezo runs on CometBFT without a
    // peer-to-peer txpool of that kind; the validator-led submission
    // path accepts both transactions, and both get included.
    const sponsor = await freshSponsor()
    const authority = await freshAuthority()

    const authToT1 = await signAuthorization(authority, {
      chainId,
      address: targetAddr,
      nonce: 0n,
    })
    const r1 = await sendSetCodeTx(sponsor, {
      to: authority.address,
      authorizationList: [authToT1],
    })
    expect(r1.receipt.status).to.equal("0x1")

    const authToT2 = await signAuthorization(authority, {
      chainId,
      address: target2Addr,
      nonce: 1n,
    })
    const r2 = await sendSetCodeTx(sponsor, {
      to: authority.address,
      authorizationList: [authToT2],
    })
    expect(r2.receipt.status).to.equal("0x1")

    // Final delegation is the second target — second auth wins.
    const code: string = await ethers.provider.send("eth_getCode", [
      authority.address,
      "latest",
    ])
    expect(code.toLowerCase()).to.equal(expectedDelegationCode(target2Addr))
    expect(
      await ethers.provider.getTransactionCount(authority.address),
    ).to.equal(2)
  })

  it("processes two auths from the same authority in a single tx", async function () {
    // Divergence #1, surface (b): single-tx multi-auth single-slot rejection.
    // Geth's txpool would single-slot-reject this construction; mezod
    // accepts it because authorizations are processed inline by the
    // keeper, not gated by a mempool-level authority reservation. Both
    // tuples apply; the second installs the final delegation.
    const sponsor = await freshSponsor()
    const authority = await freshAuthority()

    const first = await signAuthorization(authority, {
      chainId,
      address: targetAddr,
      nonce: 0n,
    })
    const second = await signAuthorization(authority, {
      chainId,
      address: target2Addr,
      nonce: 1n,
    })

    const { receipt } = await sendSetCodeTx(sponsor, {
      to: authority.address,
      authorizationList: [first, second],
    })
    expect(receipt.status).to.equal("0x1")

    const code: string = await ethers.provider.send("eth_getCode", [
      authority.address,
      "latest",
    ])
    expect(code.toLowerCase()).to.equal(expectedDelegationCode(target2Addr))
    expect(
      await ethers.provider.getTransactionCount(authority.address),
    ).to.equal(2)
  })

  it("rejects authorization tuples targeting precompiles (stock and every Mezo custom)", async function () {
    // Stock geth accepts authorization tuples whose target is any
    // 20-byte address, including the precompile range (0x01..0x12);
    // those are no-ops at runtime because precompiles have no stored
    // bytecode upstream. Mezo additionally stores facade bytecode at
    // the custom precompile addresses (0x7b7c…), so an unguarded
    // delegation there would actually execute the precompile's surface
    // in the authority's storage. To keep the surface uniform across
    // precompile families, mezod rejects any authorization whose
    // target is a precompile; the offending tuple is silently skipped
    // per the EIP's per-tuple rule, the rest of the tx proceeds.
    //
    // The loop covers every currently-registered custom precompile so
    // adding a new precompile without updating the keeper's reject
    // list (or renumbering the range) trips the test.
    const sponsor = await freshSponsor()
    const authStock = await freshAuthority()
    const authReal = await freshAuthority()
    const customAuthorities = await Promise.all(
      MEZO_CUSTOM_PRECOMPILES.map(() => freshAuthority()),
    )

    const tupleStock = await signAuthorization(authStock, {
      chainId,
      address: STOCK_PRECOMPILE,
      nonce: 0n,
    })
    const tuplesCustom = await Promise.all(
      MEZO_CUSTOM_PRECOMPILES.map((addr, i) =>
        signAuthorization(customAuthorities[i], {
          chainId,
          address: addr,
          nonce: 0n,
        }),
      ),
    )
    const tupleReal = await signAuthorization(authReal, {
      chainId,
      address: targetAddr,
      nonce: 0n,
    })

    const { receipt } = await sendSetCodeTx(sponsor, {
      to: targetAddr,
      authorizationList: [tupleStock, ...tuplesCustom, tupleReal],
    })
    expect(receipt.status).to.equal("0x1")

    // Stock precompile target — silently skipped.
    expect(
      await ethers.provider.send("eth_getCode", [authStock.address, "latest"]),
    ).to.equal(EMPTY_CODE)
    expect(
      await ethers.provider.getTransactionCount(authStock.address),
    ).to.equal(0)

    // Every custom precompile target — silently skipped, nonce not bumped.
    for (let i = 0; i < customAuthorities.length; i++) {
      const auth = customAuthorities[i]
      const addr = MEZO_CUSTOM_PRECOMPILES[i]
      expect(
        await ethers.provider.send("eth_getCode", [auth.address, "latest"]),
        `code for authority targeting ${addr}`,
      ).to.equal(EMPTY_CODE)
      expect(
        await ethers.provider.getTransactionCount(auth.address),
        `nonce for authority targeting ${addr}`,
      ).to.equal(0)
    }

    // The real-target authority is installed normally.
    const realCode: string = await ethers.provider.send("eth_getCode", [
      authReal.address,
      "latest",
    ])
    expect(realCode.toLowerCase()).to.equal(expectedDelegationCode(targetAddr))
    expect(
      await ethers.provider.getTransactionCount(authReal.address),
    ).to.equal(1)
  })

  it("pins type-0x04 sender recovery via geth's PragueSigner (no Mezo-specific signer adapter)", async function () {
    // Divergence #3: Mezo does NOT install a custom signer adapter
    // over geth's PragueSigner for type-0x04 envelopes. Sender
    // recovery and auth-tuple recovery both flow through unmodified
    // upstream code. A regression introducing a Mezo wrapper that
    // mishandles the type-0x04 envelope (or its embedded auth tuples)
    // would break these RPC-surfaced addresses.
    const sponsor = await freshSponsor()
    const authority = await freshAuthority()

    const auth = await signAuthorization(authority, {
      chainId,
      address: targetAddr,
      nonce: 0n,
    })

    const { receipt, rpcTx } = await sendSetCodeTx(sponsor, {
      to: authority.address,
      authorizationList: [auth],
    })

    expect(receipt.status).to.equal("0x1")
    // Envelope signer recovered via PragueSigner.
    expect(getAddress(rpcTx.from)).to.equal(getAddress(sponsor.address))
    // Auth tuple's target field round-trips through the RPC layer.
    expect(rpcTx.authorizationList).to.have.length(1)
    expect(getAddress(rpcTx.authorizationList[0].address)).to.equal(
      getAddress(targetAddr),
    )
    // The delegation was installed on `authority` — this can only
    // happen if the auth tuple's signer was correctly recovered
    // (PragueSigner over the auth-tuple hash, unmodified). A
    // Mezo-specific signer adapter that mis-recovers the tuple signer
    // would silently skip the tuple and leave the authority code empty.
    const code: string = await ethers.provider.send("eth_getCode", [
      authority.address,
      "latest",
    ])
    expect(code.toLowerCase()).to.equal(expectedDelegationCode(targetAddr))
  })
})
