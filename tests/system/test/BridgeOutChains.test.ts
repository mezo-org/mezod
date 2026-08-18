import { expect } from "chai"
import hre from "hardhat"
import { ethers } from "hardhat"
import assetsbridgeabi from "../../../precompile/assetsbridge/abi.json"
import btcabi from "../../../precompile/btctoken/abi.json"
import maintenanceabi from "../../../precompile/maintenance/abi.json"
import validatorpoolabi from "../../../precompile/validatorpool/abi.json"
import { BridgeOutDelegate, SimpleToken } from "../typechain-types"
import { getDeployedContract } from "./helpers/contract"
import { extractMessage } from "./helpers/rpc-error"

const validatorPoolPrecompileAddress =
  "0x7b7c000000000000000000000000000000000011"
const assetsBridgePrecompileAddress =
  "0x7b7c000000000000000000000000000000000012"
const maintenancePrecompileAddress =
  "0x7b7c000000000000000000000000000000000013"
const btcTokenPrecompileAddress = "0x7b7c000000000000000000000000000000000000"

// Target chains of the bridgeOut method.
const ethereumChain = 0
const bitcoinChain = 1

// A chain outside the target chain enum.
const unsupportedChain = 5

// A bridge-out to Ethereum takes a 20-byte recipient address. A bridge-out to
// Bitcoin takes a var-len recipient script.
const ethereumRecipient = "150bCF49Ee8E2Bd9f59e991821DE5B74C6D876aA"
const bitcoinRecipient = "1976a91462e907b15cbf27d5425399ebf6f0fb50ebb88f1888ac"

// The bridge-out chain set holds the target chains enabled for bridge-outs. The
// assets bridge precompile (0x7b7c...0012) exposes the owner-only methods that
// remove a chain from the set and add it back. A bridge-out to a chain outside
// the set reverts in the bridge keeper, so the gate covers direct and indirect
// bridge-outs alike.
describe("BridgeOutChains", function () {
  const { deployments } = hre

  let assetsBridge: any
  let btcToken: any
  let maintenance: any
  let validatorPool: any
  let bridgeOutDelegate: BridgeOutDelegate
  let simpleToken: SimpleToken
  let poolOwner: any
  let emergencyTeam: any
  let thirdParty: any
  let sender: any
  let simpleTokenAddress: string
  let bridgeOutDelegateAddress: string

  // The Ethereum-side token of the ERC20 mapping the bridge-out to Ethereum
  // needs. A fresh address per run keeps the mapping free of leftovers.
  const sourceTokenAddress = ethers.Wallet.createRandom().address

  const btcAmount = ethers.parseEther("0.5")

  // The ERC20 amount stays above the minimum bridge-out amount that the
  // AssetsBridgeIndirectBridgeOut suite sets for the same token and never
  // clears, so the suites can run in any order.
  const erc20Amount = ethers.parseEther("10")

  const assetsBridgeAs = (signer: any) =>
    new ethers.Contract(assetsBridgePrecompileAddress, assetsbridgeabi, signer)

  const maintenanceAs = (signer: any) =>
    new ethers.Contract(maintenancePrecompileAddress, maintenanceabi, signer)

  // eventArguments returns the arguments of the named event emitted by the
  // transaction. It throws when the transaction does not emit that event.
  const eventArguments = (receipt: any, name: string) => {
    const iface = new ethers.Interface(assetsbridgeabi)

    for (const log of receipt.logs) {
      try {
        const parsed = iface.parseLog({
          topics: log.topics as string[],
          data: log.data,
        })
        if (parsed && parsed.name === name) {
          return parsed.args
        }
      } catch {
        // Not a decodable event from our ABI, skip.
      }
    }

    throw new Error(`the transaction did not emit the ${name} event`)
  }

  // bridgeOutChains returns the chain set as plain numbers.
  const bridgeOutChains = async () =>
    Array.from(await assetsBridge.getBridgeOutChains(), (chain: any) =>
      Number(chain),
    )

  // callFailure returns the error message of a failed call. It throws when the
  // call succeeds.
  const callFailure = async (call: () => Promise<any>) => {
    try {
      await call()
    } catch (error: any) {
      return extractMessage(error)
    }

    throw new Error("the call did not fail")
  }

  const btcBridgeOutArgs = (chain: number) => [
    btcTokenPrecompileAddress,
    btcAmount,
    chain,
    Buffer.from(
      chain === bitcoinChain ? bitcoinRecipient : ethereumRecipient,
      "hex",
    ),
  ]

  // sendBTCBridgeOut approves the assets bridge and sends a native BTC
  // bridge-out to the given target chain.
  const sendBTCBridgeOut = async (chain: number) => {
    await (
      await btcToken
        .connect(sender)
        .approve(assetsBridgePrecompileAddress, btcAmount)
    ).wait()

    const tx = await assetsBridgeAs(sender).bridgeOut(
      ...btcBridgeOutArgs(chain),
    )
    return tx.wait()
  }

  // simulateBTCBridgeOut runs the same bridge-out as a call. The call surfaces
  // the revert reason of the precompile, which a sent transaction does not.
  const simulateBTCBridgeOut = (chain: number) =>
    assetsBridgeAs(sender).bridgeOut.staticCall(...btcBridgeOutArgs(chain))

  before(async function () {
    const signers = await ethers.getSigners()

    await deployments.fixture(["BridgeOutDelegate"])

    validatorPool = new ethers.Contract(
      validatorPoolPrecompileAddress,
      validatorpoolabi,
      ethers.provider,
    )
    assetsBridge = new ethers.Contract(
      assetsBridgePrecompileAddress,
      assetsbridgeabi,
      ethers.provider,
    )
    maintenance = new ethers.Contract(
      maintenancePrecompileAddress,
      maintenanceabi,
      ethers.provider,
    )
    btcToken = new ethers.Contract(
      btcTokenPrecompileAddress,
      btcabi,
      ethers.provider,
    )

    bridgeOutDelegate = await getDeployedContract("BridgeOutDelegate")
    simpleToken = await getDeployedContract("SimpleToken")
    simpleTokenAddress = await simpleToken.getAddress()
    bridgeOutDelegateAddress = await bridgeOutDelegate.getAddress()

    poolOwner = await ethers.getSigner(await validatorPool.owner())
    emergencyTeam = signers[1]
    thirdParty = signers[2]

    // The suite bridges out from a fresh account so the balances of the shared
    // localnode accounts stay untouched.
    sender = ethers.Wallet.createRandom().connect(ethers.provider)
    await (
      await btcToken
        .connect(signers[0])
        .transfer(sender.address, ethers.parseEther("10"))
    ).wait()

    // The outflow capacity of a token is zero until the owner sets a limit, so
    // every bridge-out here needs one.
    await (
      await assetsBridgeAs(poolOwner).setOutflowLimit(
        btcTokenPrecompileAddress,
        ethers.MaxUint256,
      )
    ).wait()
    await (
      await assetsBridgeAs(poolOwner).setOutflowLimit(
        simpleTokenAddress,
        ethers.MaxUint256,
      )
    ).wait()

    // A bridge-out to Ethereum needs the ERC20 mapping of the bridged token.
    await (
      await assetsBridgeAs(poolOwner).createERC20TokenMapping(
        sourceTokenAddress,
        simpleTokenAddress,
      )
    ).wait()
    await (
      await simpleToken.connect(sender).mint(sender.address, erc20Amount)
    ).wait()
  })

  // Leave both chains enabled, the mapping deleted and the chain without an
  // Emergency Team for the other suites.
  after(async function () {
    if (!(await bridgeOutChains()).includes(bitcoinChain)) {
      await (
        await assetsBridgeAs(poolOwner).addBridgeOutChain(bitcoinChain)
      ).wait()
    }

    await (
      await assetsBridgeAs(poolOwner).deleteERC20TokenMapping(
        sourceTokenAddress,
      )
    ).wait()
    await (
      await maintenanceAs(poolOwner).setEmergencyTeam(ethers.ZeroAddress)
    ).wait()
  })

  it("reports both chains by default and bridges out to both", async function () {
    expect(await bridgeOutChains()).to.deep.equal([ethereumChain, bitcoinChain])

    expect((await sendBTCBridgeOut(bitcoinChain)).status).to.equal(1)
    expect((await sendBTCBridgeOut(ethereumChain)).status).to.equal(1)
  })

  it("rejects the chain set methods from a third party", async function () {
    expect(
      await callFailure(() =>
        assetsBridgeAs(thirdParty).removeBridgeOutChain.staticCall(
          bitcoinChain,
        ),
      ),
    ).to.include("not the owner")

    expect(
      await callFailure(() =>
        assetsBridgeAs(thirdParty).addBridgeOutChain.staticCall(bitcoinChain),
      ),
    ).to.include("not the owner")

    expect(await bridgeOutChains()).to.deep.equal([ethereumChain, bitcoinChain])
  })

  it("rejects the chain set methods from the emergency team", async function () {
    // The chain set retires a bridging venue, so it stays owner-only. The
    // Emergency Team drives the bridge lockdown instead.
    await (
      await maintenanceAs(poolOwner).setEmergencyTeam(emergencyTeam.address)
    ).wait()

    expect(
      await callFailure(() =>
        assetsBridgeAs(emergencyTeam).removeBridgeOutChain.staticCall(
          bitcoinChain,
        ),
      ),
    ).to.include("not the owner")

    expect(
      await callFailure(() =>
        assetsBridgeAs(emergencyTeam).addBridgeOutChain.staticCall(
          bitcoinChain,
        ),
      ),
    ).to.include("not the owner")

    expect(await bridgeOutChains()).to.deep.equal([ethereumChain, bitcoinChain])
  })

  it("rejects a chain outside the target chain enum", async function () {
    expect(
      await callFailure(() =>
        assetsBridgeAs(poolOwner).removeBridgeOutChain.staticCall(
          unsupportedChain,
        ),
      ),
    ).to.include("unsupported chain")

    expect(
      await callFailure(() =>
        assetsBridgeAs(poolOwner).addBridgeOutChain.staticCall(
          unsupportedChain,
        ),
      ),
    ).to.include("unsupported chain")
  })

  it("rejects adding a chain that is already in the set", async function () {
    expect(
      await callFailure(() =>
        assetsBridgeAs(poolOwner).addBridgeOutChain.staticCall(bitcoinChain),
      ),
    ).to.include("chain is already enabled for bridge-outs")
  })

  it("lets the owner remove the Bitcoin chain", async function () {
    const receipt = await (
      await assetsBridgeAs(poolOwner).removeBridgeOutChain(bitcoinChain)
    ).wait()

    expect(receipt.status).to.equal(1)

    const args = eventArguments(receipt, "BridgeOutChainRemoved")
    expect(Number(args.chain)).to.equal(bitcoinChain)

    expect(await bridgeOutChains()).to.deep.equal([ethereumChain])
  })

  it("blocks only the bridge-out to Bitcoin", async function () {
    expect(
      await callFailure(() => simulateBTCBridgeOut(bitcoinChain)),
    ).to.include("target chain is not enabled for bridge-outs")

    expect((await sendBTCBridgeOut(ethereumChain)).status).to.equal(1)

    await (
      await simpleToken
        .connect(sender)
        .approve(assetsBridgePrecompileAddress, erc20Amount)
    ).wait()

    const receipt = await (
      await assetsBridgeAs(sender).bridgeOut(
        simpleTokenAddress,
        erc20Amount,
        ethereumChain,
        Buffer.from(ethereumRecipient, "hex"),
      )
    ).wait()

    expect(receipt.status).to.equal(1)
  })

  it("rejects removing the Bitcoin chain again", async function () {
    expect(
      await callFailure(() =>
        assetsBridgeAs(poolOwner).removeBridgeOutChain.staticCall(bitcoinChain),
      ),
    ).to.include("chain is not enabled for bridge-outs")

    expect(await bridgeOutChains()).to.deep.equal([ethereumChain])
  })

  it("blocks the indirect bridge-out to Bitcoin", async function () {
    // A contract-mediated bridge-out fails without a revert reason, so the
    // receipt status and the untouched balance carry the assertion.
    await (
      await btcToken
        .connect(sender)
        .transfer(bridgeOutDelegateAddress, btcAmount)
    ).wait()

    const balanceBefore = await btcToken.balanceOf(bridgeOutDelegateAddress)

    // The explicit gas limit skips the gas estimation, so the transaction
    // reaches the chain and leaves a receipt to assert on.
    let tx: any
    try {
      tx = await bridgeOutDelegate
        .connect(sender)
        .bridgeOutBTCSuccess(Buffer.from(bitcoinRecipient, "hex"), btcAmount, {
          gasLimit: 1000000,
        })
      await tx.wait()
    } catch (error: any) {
      expect(extractMessage(error)).to.include("reverted")
    }

    const receipt = await ethers.provider.getTransactionReceipt(tx.hash)
    expect(receipt!.status).to.equal(0)
    expect(await btcToken.balanceOf(bridgeOutDelegateAddress)).to.equal(
      balanceBefore,
    )
  })

  it("lets the owner add the Bitcoin chain back", async function () {
    const receipt = await (
      await assetsBridgeAs(poolOwner).addBridgeOutChain(bitcoinChain)
    ).wait()

    expect(receipt.status).to.equal(1)

    const args = eventArguments(receipt, "BridgeOutChainAdded")
    expect(Number(args.chain)).to.equal(bitcoinChain)

    expect(await bridgeOutChains()).to.deep.equal([ethereumChain, bitcoinChain])

    expect((await sendBTCBridgeOut(bitcoinChain)).status).to.equal(1)
  })
})
