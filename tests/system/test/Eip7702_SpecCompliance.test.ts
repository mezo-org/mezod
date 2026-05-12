import { expect } from "chai"
import hre, { ethers } from "hardhat"
import { Wallet, getAddress, toQuantity, keccak256 } from "ethers"

import {
  EMPTY_CODE,
  SET_CODE_TX_TYPE,
  expectedDelegationCode,
  freshAuthority,
  freshSponsor,
  fundWallet,
  pollForReceipt,
  sendSetCodeTx,
  signAuthorization,
} from "./helpers/eip7702"

const ZERO = "0x0000000000000000000000000000000000000000"

// Authorities are charged CallNewAccountGas (25000) per tuple, then
// refunded CallNewAccountGas - TxAuthTupleGas (25000 - 12500 = 12500)
// when the authority pre-exists in state.
const REFUND_PER_EXISTING_AUTHORITY = 12500n

/**
 * Eip7702_SpecCompliance pins every behavior where mezod's EIP-7702
 * surface (type-0x04 "set code" transactions) MUST match the go-ethereum
 * Prague reference implementation. Mezo-specific divergences (no mempool
 * authority reservation, precompile-target rejection) live in
 * Eip7702_MezoDivergence — never duplicated here.
 *
 * Conventions: each it() builds its own fresh sponsor and authority via
 * freshSponsor() / freshAuthority() so cases are independent and order-
 * agnostic. Receipts are polled (HardhatEthersProvider doesn't implement
 * waitForTransaction). Coarse gas tolerances only — absolute gasUsed
 * couples to mezod's MinGasMultiplier floor and the EIP-3529 refund cap,
 * so assertions target deltas and code/storage shape.
 *
 * Test scenarios:
 *
 * | Scenario                                               | Given                                       | When                                                | Then                                                                              |
 * |--------------------------------------------------------|---------------------------------------------|-----------------------------------------------------|-----------------------------------------------------------------------------------|
 * | install delegation + type-4 envelope                   | fresh sponsor + fresh authority             | sendSetCodeTx with one auth + setSlot body          | receipt 0x1; rpcTx.type=0x04; code = designator(T1); slot written in EOA          |
 * | rotate delegation T1 -> T2, storage preserved          | authority delegated to T1, slot 42=7        | second set-code tx with auth to T2                  | code = designator(T2); slot 42 still 7; V2-only tickV2() succeeds                 |
 * | clear with 0x0 then re-delegate                        | authority delegated to T1, slot 42=7        | tuple to 0x0, then tuple back to T1                 | code empty after clear; code = designator(T1) after; slot 42 still 7              |
 * | self-sponsored authorization                           | sender == authority, auth.nonce = current+1 | sendSetCodeTx signed by the authority itself        | receipt 0x1; rpcTx.from = authority; code = designator(T1)                        |
 * | self-sponsored pre-bump nonce silently skipped         | sender == authority, auth.nonce = current   | sendSetCodeTx with that wrong-nonce tuple           | envelope OK; code empty; authority nonce += 1 (sender bump only)                  |
 * | self-delegation does not loop                          | tuple's address = authority's own address   | sendSetCodeTx with that self-pointing tuple         | code = designator(authority); auth processing does not recurse                    |
 * | A->B->T resolves exactly one hop                       | B already delegated to T                    | A signs tuple pointing to B (not T)                 | code(A)=designator(B); code(B)=designator(T); code(A)!=designator(T)              |
 * | refund CallNewAccountGas - TxAuthTupleGas              | 32-SSTORE body; fresh vs existing authority | one-tuple and two-tuple variants                    | gasUsed(fresh) - gasUsed(existing) = 12500 per existing authority                 |
 * | duplicated tuple silently skipped                      | one authority, same tuple twice             | sendSetCodeTx with [a, a]                           | receipt 0x1; code = designator(T1); authority nonce = 1                           |
 * | wrong chain id silently skipped                        | tuple signed for chainId+1                  | sendSetCodeTx with that tuple                       | code empty; authority nonce unchanged                                             |
 * | chainId = 0 (any-chain) accepted                       | tuple signed with chainId = 0               | sendSetCodeTx with that tuple                       | code = designator(T1)                                                             |
 * | corrupted-signature tuple silently skipped             | tuple's r field corrupted                   | sendSetCodeTx with that tuple                       | code empty; authority nonce unchanged                                             |
 * | delegated EOA sends a type-2 tx (EIP-3607 exempt)      | authority delegated to T1, funded           | authority sends type-2 setSlot(99, 123)             | receipt 0x1; from = authority; nonce = 2; slot 99 = 123                           |
 * | one-level resolve via CALL / STATICCALL / DELEGATECALL | authority delegated to T1, Eip7702Caller    | callInto / staticCallInto / delegateCallInto        | CALL writes to authority; STATIC reads it; DELEGATECALL writes to caller storage  |
 * | EXTCODE* observe 23-byte delegation designator         | authority delegated to T1                   | Eip7702ExtCodeReader.size / copy / hash             | size=23; copy=designator(T1); hash=keccak256(designator) != keccak256(targetCode) |
 * | intrinsic CallNewAccountGas (25000) per fresh tuple    | 32-SSTORE body; 1-tuple vs 2-tuple tx       | sendSetCodeTx with 1 vs 2 fresh authorities         | gasUsed(2 tuples) - gasUsed(1 tuple) = exactly 25000                              |
 * | EIP-2929 cold->warm CALL into delegated EOA            | authority delegated to T1                   | contract issues two CALLs to authority, tick() body | gasFirst - gasSecond ~= 5000 (+/- 200): cold A + cold T = 2 x 2500                |
 */
describe("Eip7702_SpecCompliance", function () {
  const { deployments } = hre

  let targetAddr: string
  let target2Addr: string
  let readerAddr: string
  let callerAddr: string
  let chainId: bigint

  before(async function () {
    await deployments.fixture(["eip7702-fixtures"])
    targetAddr = (await deployments.get("Eip7702TargetV1")).address
    target2Addr = (await deployments.get("Eip7702TargetV2")).address
    readerAddr = (await deployments.get("Eip7702ExtCodeReader")).address
    callerAddr = (await deployments.get("Eip7702Caller")).address
    chainId = (await ethers.provider.getNetwork()).chainId
  })

  it("installs delegation, exposes the type-4 RPC envelope, and routes calls into target code", async function () {
    const sponsor = await freshSponsor()
    const authority = await freshAuthority()

    const target = await ethers.getContractAt("Eip7702TargetV1", targetAddr)
    const setSlotData = target.interface.encodeFunctionData("setSlot", [42n, 7n])

    const auth = await signAuthorization(authority, {
      chainId,
      address: targetAddr,
      nonce: 0n,
    })

    const { receipt, rpcTx } = await sendSetCodeTx(sponsor, {
      to: authority.address,
      data: setSlotData,
      authorizationList: [auth],
    })

    expect(receipt, "receipt for set-code tx").to.not.be.null
    expect(receipt.status).to.equal("0x1")

    expect(rpcTx.type).to.equal(toQuantity(SET_CODE_TX_TYPE))
    expect(rpcTx.authorizationList).to.have.length(1)
    expect(getAddress(rpcTx.authorizationList[0].address)).to.equal(targetAddr)
    expect(BigInt(rpcTx.authorizationList[0].chainId)).to.equal(chainId)

    const code: string = await ethers.provider.send("eth_getCode", [
      authority.address,
      "latest",
    ])
    expect(code.toLowerCase()).to.equal(expectedDelegationCode(targetAddr))

    // Storage write landed in the EOA's slot, not the target's.
    const eoaView = target.attach(authority.address) as typeof target
    expect(await eoaView.readSlot(42n)).to.equal(7n)
    expect(await target.readSlot(42n)).to.equal(0n)

    // The Touched event's `self` is the EOA, not the target.
    const touchedTopic = target.interface.getEvent("Touched")!.topicHash
    const touched = receipt.logs.find(
      (l: any) => l.topics?.[0]?.toLowerCase() === touchedTopic.toLowerCase(),
    )
    expect(touched, "Touched log").to.exist
    // topic[2] is `self` (second indexed parameter). 32-byte left-padded
    // address; right-most 20 bytes equal the EOA address.
    const selfFromLog = "0x" + touched.topics[2].slice(-40)
    expect(getAddress(selfFromLog)).to.equal(getAddress(authority.address))
  })

  it("rotates delegation to a second target while preserving storage", async function () {
    const sponsor = await freshSponsor()
    const authority = await freshAuthority()

    const target = await ethers.getContractAt("Eip7702TargetV1", targetAddr)
    const target2 = await ethers.getContractAt("Eip7702TargetV2", target2Addr)

    // Install T1 and write storage.
    const installAuth = await signAuthorization(authority, {
      chainId,
      address: targetAddr,
      nonce: 0n,
    })
    await sendSetCodeTx(sponsor, {
      to: authority.address,
      data: target.interface.encodeFunctionData("setSlot", [42n, 7n]),
      authorizationList: [installAuth],
    })

    // Rotate to T2.
    const rotateAuth = await signAuthorization(authority, {
      chainId,
      address: target2Addr,
      nonce: 1n,
    })
    await sendSetCodeTx(sponsor, {
      to: authority.address,
      authorizationList: [rotateAuth],
    })

    const code: string = await ethers.provider.send("eth_getCode", [
      authority.address,
      "latest",
    ])
    expect(code.toLowerCase()).to.equal(expectedDelegationCode(target2Addr))

    // Storage survives rotation: reading via T2's readSlot still returns 7.
    const eoaAsT2 = target2.attach(authority.address) as typeof target2
    expect(await eoaAsT2.readSlot(42n)).to.equal(7n)

    // V2-only function `tickV2` is callable from the EOA. V1 doesn't
    // expose this selector, so a successful send proves the delegated
    // code actually swapped (not just the designator address).
    const tickV2Tx = await eoaAsT2.tickV2()
    const tickV2Receipt = await pollForReceipt(tickV2Tx.hash)
    expect(tickV2Receipt.status).to.equal("0x1")
  })

  it("clears delegation with address=0x0 and re-delegates later", async function () {
    const sponsor = await freshSponsor()
    const authority = await freshAuthority()

    const target = await ethers.getContractAt("Eip7702TargetV1", targetAddr)

    // Install T1 and write storage.
    const install = await signAuthorization(authority, {
      chainId,
      address: targetAddr,
      nonce: 0n,
    })
    await sendSetCodeTx(sponsor, {
      to: authority.address,
      data: target.interface.encodeFunctionData("setSlot", [42n, 7n]),
      authorizationList: [install],
    })

    // Clear.
    const clearAuth = await signAuthorization(authority, {
      chainId,
      address: ZERO,
      nonce: 1n,
    })
    await sendSetCodeTx(sponsor, {
      to: authority.address,
      authorizationList: [clearAuth],
    })

    expect(
      await ethers.provider.send("eth_getCode", [authority.address, "latest"]),
    ).to.equal(EMPTY_CODE)

    // Re-delegate.
    const reAuth = await signAuthorization(authority, {
      chainId,
      address: targetAddr,
      nonce: 2n,
    })
    await sendSetCodeTx(sponsor, {
      to: authority.address,
      authorizationList: [reAuth],
    })

    const code: string = await ethers.provider.send("eth_getCode", [
      authority.address,
      "latest",
    ])
    expect(code.toLowerCase()).to.equal(expectedDelegationCode(targetAddr))

    // Storage survives clear-then-re-delegate: slot 42 still reads 7.
    const eoaView = target.attach(authority.address) as typeof target
    expect(await eoaView.readSlot(42n)).to.equal(7n)
  })

  it("supports self-sponsored authorizations (authority == sender)", async function () {
    // The authority IS the sender; fund it enough to pay for the
    // type-0x04 tx and the auth-tuple intrinsic charge.
    const authority = await freshSponsor()

    const target = await ethers.getContractAt("Eip7702TargetV1", targetAddr)

    // Self-sponsored: auth.nonce is the authority's nonce *after* the
    // sender-sequence bump, so current+1.
    const current = BigInt(
      await ethers.provider.getTransactionCount(authority.address),
    )
    const auth = await signAuthorization(authority, {
      chainId,
      address: targetAddr,
      nonce: current + 1n,
    })

    const { receipt, rpcTx } = await sendSetCodeTx(authority, {
      to: targetAddr,
      // Use `tick()` so the executed body returns successfully; an
      // empty calldata would hit Solidity's automatic revert on
      // unmatched selector.
      data: target.interface.encodeFunctionData("tick"),
      authorizationList: [auth],
    })

    expect(receipt.status).to.equal("0x1")
    expect(getAddress(rpcTx.from)).to.equal(getAddress(authority.address))

    const code: string = await ethers.provider.send("eth_getCode", [
      authority.address,
      "latest",
    ])
    expect(code.toLowerCase()).to.equal(expectedDelegationCode(targetAddr))
  })

  it("silently skips a self-sponsored auth signed against the pre-bump nonce", async function () {
    // Spec pins that self-sponsored authorizations sign their tuple
    // against the POST-bump sender nonce (current + 1). A tuple signed
    // against the pre-bump nonce (current) must be silently skipped at
    // validate-time: envelope still applies (sender nonce advances),
    // but the authority's code stays empty and its own nonce does NOT
    // bump.
    const authority = await freshSponsor()
    const target = await ethers.getContractAt("Eip7702TargetV1", targetAddr)

    const current = BigInt(
      await ethers.provider.getTransactionCount(authority.address),
    )
    const wrongAuth = await signAuthorization(authority, {
      chainId,
      address: targetAddr,
      nonce: current,
    })

    const { receipt } = await sendSetCodeTx(authority, {
      to: targetAddr,
      data: target.interface.encodeFunctionData("tick"),
      authorizationList: [wrongAuth],
    })

    expect(receipt.status).to.equal("0x1")
    expect(
      await ethers.provider.send("eth_getCode", [authority.address, "latest"]),
    ).to.equal(EMPTY_CODE)
    expect(
      await ethers.provider.getTransactionCount(authority.address),
    ).to.equal(current + 1n)
  })

  it("handles self-delegation without looping", async function () {
    // Authority delegates to itself. geth's Call resolves exactly one
    // level, so even though the delegation forms a cycle, the keeper's
    // authorization loop must NOT iterate. We do not call the self-
    // delegated EOA in this test — executing the raw 23-byte designator
    // as bytecode would hit INVALID (0xef) and revert. The point pinned
    // here is that auth processing does not recurse and the resulting
    // code is the self-pointing designator.
    const sponsor = await freshSponsor()
    const authority = await freshAuthority()

    const target = await ethers.getContractAt("Eip7702TargetV1", targetAddr)

    const auth = await signAuthorization(authority, {
      chainId,
      address: authority.address,
      nonce: 0n,
    })

    const { receipt } = await sendSetCodeTx(sponsor, {
      to: targetAddr,
      data: target.interface.encodeFunctionData("tick"),
      authorizationList: [auth],
    })

    expect(receipt.status).to.equal("0x1")

    const code: string = await ethers.provider.send("eth_getCode", [
      authority.address,
      "latest",
    ])
    expect(code.toLowerCase()).to.equal(
      expectedDelegationCode(authority.address),
    )
  })

  it("delegation resolves exactly one level (A→B→T does not collapse to A→T)", async function () {
    const sponsorB = await freshSponsor()
    const sponsorA = await freshSponsor()
    const authA = await freshAuthority()
    const authB = await freshAuthority()

    // Install B → T.
    const bToT = await signAuthorization(authB, {
      chainId,
      address: targetAddr,
      nonce: 0n,
    })
    await sendSetCodeTx(sponsorB, {
      to: authB.address,
      authorizationList: [bToT],
    })

    // Install A → B (B's address, NOT the underlying target).
    const aToB = await signAuthorization(authA, {
      chainId,
      address: authB.address,
      nonce: 0n,
    })
    await sendSetCodeTx(sponsorA, {
      to: authA.address,
      authorizationList: [aToB],
    })

    // Pins the no-collapse rule: A points at B, B points at T, eth_getCode
    // on A returns the designator for B (NOT for T). Delegation resolves
    // exactly one hop.
    const codeA: string = await ethers.provider.send("eth_getCode", [
      authA.address,
      "latest",
    ])
    expect(codeA.toLowerCase()).to.equal(expectedDelegationCode(authB.address))

    const codeB: string = await ethers.provider.send("eth_getCode", [
      authB.address,
      "latest",
    ])
    expect(codeB.toLowerCase()).to.equal(expectedDelegationCode(targetAddr))

    // Sanity: A's code is NOT the designator for T.
    expect(codeA.toLowerCase()).to.not.equal(expectedDelegationCode(targetAddr))
  })

  it("refunds CallNewAccountGas - TxAuthTupleGas for an existing authority", async function () {
    // Two confounders to keep out of the way:
    //
    //  1. mezod floors gasUsed at gasLimit * MinGasMultiplier, so both
    //     fresh and existing runs must consume more than that floor for
    //     the refund to be observable in gasUsed. We size the workload
    //     to clear the floor at MinGasMultiplier values well above the
    //     current 0.5 default (resilient to a non-spec bump up to ~0.9).
    //  2. EIP-3529 caps refunds at gasUsed / 5. To see the full
    //     12500-per-existing-authority refund, gasUsed must be >= 62500
    //     per existing auth (so cap >= refund).
    //
    // We satisfy both by writing 32 fresh storage slots in the call
    // body (32 * 22100 = 707_200 gas of SSTOREs) with gasLimit at
    // 800_000. Pre-refund gasUsed lands ~750k-780k, above the
    // 720_000 floor at MinGasMultiplier=0.9 and well above the
    // 62_500 cap-pivot.

    const target = await ethers.getContractAt("Eip7702TargetV1", targetAddr)
    const callData = target.interface.encodeFunctionData("setSlotN", [
      0n,
      1n,
      32n,
    ])
    const gasLimit = 800_000n

    async function runOne(funded: boolean): Promise<bigint> {
      const sponsor = await freshSponsor()
      const authority = await freshAuthority({ funded })
      const auth = await signAuthorization(authority, {
        chainId,
        address: targetAddr,
        nonce: 0n,
      })
      const { receipt } = await sendSetCodeTx(sponsor, {
        to: authority.address,
        data: callData,
        authorizationList: [auth],
        gasLimit,
      })
      expect(receipt.status).to.equal("0x1")
      return BigInt(receipt.gasUsed)
    }

    async function runTwo(funded: boolean): Promise<bigint> {
      // Single sponsor signs the type-0x04 envelope, two distinct
      // authorities supply the tuples. msg.To is one of the authorities
      // so the call body still routes through a delegated EOA.
      const sponsor = await freshSponsor()
      const authA = await freshAuthority({ funded })
      const authB = await freshAuthority({ funded })
      const sA = await signAuthorization(authA, {
        chainId,
        address: targetAddr,
        nonce: 0n,
      })
      const sB = await signAuthorization(authB, {
        chainId,
        address: targetAddr,
        nonce: 0n,
      })
      const { receipt } = await sendSetCodeTx(sponsor, {
        to: authA.address,
        data: callData,
        authorizationList: [sA, sB],
        gasLimit,
      })
      expect(receipt.status).to.equal("0x1")
      return BigInt(receipt.gasUsed)
    }

    const gasUsedFresh = await runOne(false)
    const gasUsedExisting = await runOne(true)
    expect(gasUsedFresh - gasUsedExisting).to.equal(
      REFUND_PER_EXISTING_AUTHORITY,
    )

    const gasUsedFresh2 = await runTwo(false)
    const gasUsedExisting2 = await runTwo(true)
    expect(gasUsedFresh2 - gasUsedExisting2).to.equal(
      REFUND_PER_EXISTING_AUTHORITY * 2n,
    )
  })

  it("silently skips a duplicated authorization tuple", async function () {
    const sponsor = await freshSponsor()
    const authority = await freshAuthority()

    const a = await signAuthorization(authority, {
      chainId,
      address: targetAddr,
      nonce: 0n,
    })
    const { receipt } = await sendSetCodeTx(sponsor, {
      to: authority.address,
      authorizationList: [a, a],
    })

    expect(receipt.status).to.equal("0x1")
    const code: string = await ethers.provider.send("eth_getCode", [
      authority.address,
      "latest",
    ])
    // First tuple wins; second hits the nonce-mismatch branch and is
    // silently dropped.
    expect(code.toLowerCase()).to.equal(expectedDelegationCode(targetAddr))
    expect(
      await ethers.provider.getTransactionCount(authority.address),
    ).to.equal(1)
  })

  it("silently skips a tuple with the wrong chain id", async function () {
    const sponsor = await freshSponsor()
    const authority = await freshAuthority()

    // chainId mismatched but not 0 -> validateSetCodeAuthorization
    // returns errSetCodeAuthorizationWrongChainID, silent skip.
    const wrongChain = await signAuthorization(authority, {
      chainId: chainId + 1n,
      address: targetAddr,
      nonce: 0n,
    })
    const { receipt } = await sendSetCodeTx(sponsor, {
      to: authority.address,
      authorizationList: [wrongChain],
    })

    expect(receipt.status).to.equal("0x1")
    expect(
      await ethers.provider.send("eth_getCode", [authority.address, "latest"]),
    ).to.equal(EMPTY_CODE)
    // Silent skip must NOT bump authority nonce — validate-time reject.
    expect(
      await ethers.provider.getTransactionCount(authority.address),
    ).to.equal(0)
  })

  it("accepts a tuple with chainId = 0 (any-chain)", async function () {
    const sponsor = await freshSponsor()
    const authority = await freshAuthority()

    const anyChain = await signAuthorization(authority, {
      chainId: 0n,
      address: targetAddr,
      nonce: 0n,
    })
    const { receipt } = await sendSetCodeTx(sponsor, {
      to: authority.address,
      authorizationList: [anyChain],
    })

    expect(receipt.status).to.equal("0x1")
    const code: string = await ethers.provider.send("eth_getCode", [
      authority.address,
      "latest",
    ])
    expect(code.toLowerCase()).to.equal(expectedDelegationCode(targetAddr))
  })

  it("silently skips a tuple with a corrupted signature", async function () {
    const sponsor = await freshSponsor()
    const authority = await freshAuthority()

    const corrupted = await signAuthorization(
      authority,
      { chainId, address: targetAddr, nonce: 0n },
      { corruptR: true },
    )
    const { receipt } = await sendSetCodeTx(sponsor, {
      to: authority.address,
      authorizationList: [corrupted],
    })

    expect(receipt.status).to.equal("0x1")
    expect(
      await ethers.provider.send("eth_getCode", [authority.address, "latest"]),
    ).to.equal(EMPTY_CODE)
    // Silent skip must NOT bump authority nonce — validate-time reject.
    expect(
      await ethers.provider.getTransactionCount(authority.address),
    ).to.equal(0)
  })

  it("accepts a normal DynamicFee tx FROM a delegated EOA (EIP-3607 exemption, sender perspective)", async function () {
    // app/ante/evm/eth.go:95-101: a sender whose account code parses as
    // a valid EIP-7702 delegation designator is exempted from the 3607
    // contract-sender check.
    const sponsor = await freshSponsor()
    const authority = Wallet.createRandom().connect(ethers.provider)
    await fundWallet(authority.address, ethers.parseEther("0.1"))

    const installAuth = await signAuthorization(authority, {
      chainId,
      address: targetAddr,
      nonce: 0n,
    })
    await sendSetCodeTx(sponsor, {
      to: authority.address,
      authorizationList: [installAuth],
    })

    // The authority's account now carries the 0xef0100||target code.
    // A plain type-2 tx FROM authority must still be accepted by the
    // ante despite the non-empty code field. Use setSlot so the body
    // has an observable side-effect — proves the delegated code actually
    // ran, not just that the ante didn't reject the sender.
    const target = await ethers.getContractAt("Eip7702TargetV1", targetAddr)
    const tx = await authority.sendTransaction({
      type: 2,
      to: authority.address,
      data: target.interface.encodeFunctionData("setSlot", [99n, 123n]),
      gasLimit: 200_000n,
    })
    const receipt = await pollForReceipt(tx.hash)
    expect(receipt, "receipt for delegated-EOA type-2 tx").to.not.be.null
    expect(receipt.status).to.equal("0x1")
    expect(getAddress(receipt.from)).to.equal(getAddress(authority.address))
    // Auth tuple bumped authority nonce to 1; the type-2 tx from
    // authority bumps to 2. Pins the envelope was processed by the
    // authority and was not a sponsor-driven retry.
    expect(
      await ethers.provider.getTransactionCount(authority.address),
    ).to.equal(2)
    // Storage write landed in the authority's slot — proves the body
    // executed under delegation (DELEGATECALL-style routing into target
    // code).
    const eoaView = target.attach(authority.address) as typeof target
    expect(await eoaView.readSlot(99n)).to.equal(123n)
  })

  it("resolves one level of delegation when invoked from a contract via CALL / STATICCALL / DELEGATECALL", async function () {
    // Pins geth's EVM call resolver: a contract calling a delegated EOA
    // must follow exactly one hop of delegation. Top-level
    // EOA-originated calls are already exercised by the install test at
    // L52; this test pins the same rule from inside a call frame, which
    // is a distinct code path through the EVM's resolver.
    const sponsor = await freshSponsor()
    const authority = await freshAuthority()

    const target = await ethers.getContractAt("Eip7702TargetV1", targetAddr)
    const caller = await ethers.getContractAt("Eip7702Caller", callerAddr)

    const installAuth = await signAuthorization(authority, {
      chainId,
      address: targetAddr,
      nonce: 0n,
    })
    await sendSetCodeTx(sponsor, {
      to: authority.address,
      authorizationList: [installAuth],
    })

    // (a) CALL via contract — storage write lands in the AUTHORITY's
    // slots (CALL preserves the callee's storage context, and under
    // delegation the callee's storage IS the authority's storage).
    const setSlotData = target.interface.encodeFunctionData("setSlot", [
      55n,
      77n,
    ])
    const callTx = await caller.callInto(authority.address, setSlotData)
    await pollForReceipt(callTx.hash)
    const targetView = target.attach(authority.address) as typeof target
    expect(await targetView.readSlot(55n)).to.equal(77n)
    // Sanity: the underlying target's storage is untouched.
    expect(await target.readSlot(55n)).to.equal(0n)

    // (b) STATICCALL via contract — read back the value just written;
    // confirms the resolver routes static reads through delegation.
    const readSlotData = target.interface.encodeFunctionData("readSlot", [55n])
    const [okStatic, retStatic] = await caller.staticCallInto.staticCall(
      authority.address,
      readSlotData,
    )
    expect(okStatic).to.equal(true)
    expect(BigInt(retStatic)).to.equal(77n)

    // (c) DELEGATECALL via contract — storage write under
    // DELEGATECALL preserves the CALLER's storage context. The caller
    // here is Eip7702Caller (the deployed contract), so the slot lands
    // in Eip7702Caller's storage, NOT the authority's.
    const setSlotDelegateData = target.interface.encodeFunctionData("setSlot", [
      88n,
      99n,
    ])
    const delegateTx = await caller.delegateCallInto(
      authority.address,
      setSlotDelegateData,
    )
    await pollForReceipt(delegateTx.hash)
    expect(await caller.readSlot(88n)).to.equal(99n)
    // Sanity: the storage did NOT land in the authority's slot 88.
    expect(await targetView.readSlot(88n)).to.equal(0n)
  })

  it("makes EXTCODE* observe the 23-byte delegation designator", async function () {
    const sponsor = await freshSponsor()
    const authority = await freshAuthority()

    const installAuth = await signAuthorization(authority, {
      chainId,
      address: targetAddr,
      nonce: 0n,
    })
    await sendSetCodeTx(sponsor, {
      to: authority.address,
      authorizationList: [installAuth],
    })

    const reader = await ethers.getContractAt(
      "Eip7702ExtCodeReader",
      readerAddr,
    )

    expect(await reader.sizeOf(authority.address)).to.equal(23n)

    const expected = expectedDelegationCode(targetAddr)
    expect((await reader.copyOf(authority.address)).toLowerCase()).to.equal(
      expected,
    )

    const expectedHash = keccak256(expected)
    const observedHash = await reader.hashOf(authority.address)
    expect(observedHash.toLowerCase()).to.equal(expectedHash.toLowerCase())

    // Sanity: not the same as keccak256 of the underlying target code.
    const targetCode: string = await ethers.provider.send("eth_getCode", [
      targetAddr,
      "latest",
    ])
    expect(observedHash.toLowerCase()).to.not.equal(
      keccak256(targetCode).toLowerCase(),
    )
  })

  it("charges CallNewAccountGas (25000) intrinsic per fresh authorization tuple", async function () {
    // Companion to the refund test above: the refund case pins
    // (CallNewAccountGas - TxAuthTupleGas) as the per-existing-authority
    // refund. This case isolates the COMPLEMENTARY signal: the absolute
    // intrinsic charge per fresh authority, with no refund path engaged.
    // Two type-0x04 txs with identical call bodies but a different
    // number of fresh-authority auth tuples (1 vs 2); the gasUsed delta
    // MUST equal exactly CallNewAccountGas = 25 000, i.e. the per-auth
    // contribution to intrinsic gas that geth's core.IntrinsicGas folds
    // in via SetCodeAuthorizations.
    //
    // Workload sizing follows the refund test: 32 fresh SSTOREs to clear
    // mezod's MinGasMultiplier floor so per-tuple intrinsic gas is
    // observable in the receipt's gasUsed.

    const target = await ethers.getContractAt("Eip7702TargetV1", targetAddr)
    const callData = target.interface.encodeFunctionData("setSlotN", [
      0n,
      1n,
      32n,
    ])
    const gasLimit = 800_000n
    const CALL_NEW_ACCOUNT_GAS = 25_000n

    async function runWithAuths(n: number): Promise<bigint> {
      const sponsor = await freshSponsor()
      const authorities: Awaited<ReturnType<typeof freshAuthority>>[] = []
      const auths = []
      for (let i = 0; i < n; i++) {
        const authority = await freshAuthority({ funded: false })
        authorities.push(authority)
        auths.push(
          await signAuthorization(authority, {
            chainId,
            address: targetAddr,
            nonce: 0n,
          }),
        )
      }
      const { receipt } = await sendSetCodeTx(sponsor, {
        to: authorities[0].address,
        data: callData,
        authorizationList: auths,
        gasLimit,
      })
      expect(receipt.status).to.equal("0x1")
      return BigInt(receipt.gasUsed)
    }

    const gasOneAuth = await runWithAuths(1)
    const gasTwoAuths = await runWithAuths(2)
    expect(gasTwoAuths - gasOneAuth).to.equal(CALL_NEW_ACCOUNT_GAS)
  })

  it("CALL into a delegated EOA pays cold access on first call, warm on second (EIP-2929)", async function () {
    // First install A -> targetV1 in a stand-alone tx. Then, in a second
    // tx, have a contract issue two sequential CALLs to A and emit the
    // gasleft() delta around each. Access lists are tx-scoped, so the
    // first call sees A and the delegate T as cold (each adding 2 500
    // gas via COLD_ACCOUNT_ACCESS_COST minus the warm baseline of 100);
    // the second call sees both as warm. Predicted delta:
    //   first  ≈ CALL_BASE + COLD_EXTRA(A) + COLD_EXTRA(T) + body
    //   second ≈ CALL_BASE                                  + body
    // → delta = 2 × 2 500 = 5 000.
    //
    // tick() is used as the call body precisely because it touches no
    // storage and returns no data: the body's gas cancels exactly across
    // the two calls so the only surviving signal is the cold-vs-warm
    // access-list cost.
    const sponsor = await freshSponsor()
    const authority = await freshAuthority({ funded: true })

    const installAuth = await signAuthorization(authority, {
      chainId,
      address: targetAddr,
      nonce: 0n,
    })
    const installRes = await sendSetCodeTx(sponsor, {
      to: authority.address,
      authorizationList: [installAuth],
    })
    expect(installRes.receipt.status).to.equal("0x1")

    const caller = await ethers.getContractAt("Eip7702Caller", callerAddr)
    const target = await ethers.getContractAt("Eip7702TargetV1", targetAddr)
    const tickData = target.interface.encodeFunctionData("tick", [])

    const txResp = await caller.callTwiceMeasured(authority.address, tickData)
    const receipt = await pollForReceipt(txResp.hash)
    expect(receipt, "receipt for callTwiceMeasured tx").to.not.be.null
    expect(receipt.status).to.equal("0x1")

    const callGasEvent = caller.interface.getEvent("CallGas")
    const log = receipt.logs.find(
      (l: any) =>
        getAddress(l.address) === callerAddr &&
        l.topics[0] === callGasEvent.topicHash,
    )
    expect(log, "CallGas event in receipt").to.not.be.undefined
    const decoded = caller.interface.decodeEventLog(
      callGasEvent,
      log.data,
      log.topics,
    )
    const gasFirst: bigint = decoded.gasFirst
    const gasSecond: bigint = decoded.gasSecond

    expect(gasSecond).to.be.lessThan(gasFirst)
    const delta = gasFirst - gasSecond
    // EIP-2929 cold-vs-warm spread is 2 500 per address (COLD 2 600 minus
    // WARM 100). Two cold addresses on the first call (A + delegate T)
    // turn warm by the second, giving a 5 000 spread. Tolerance ±200
    // absorbs any nondeterminism in the surrounding call frame
    // bookkeeping (memory-expansion rounding, returndata copy) — none of
    // which legitimately fluctuate for tick() in practice.
    expect(delta).to.be.greaterThanOrEqual(4800n)
    expect(delta).to.be.lessThanOrEqual(5200n)
  })
})
