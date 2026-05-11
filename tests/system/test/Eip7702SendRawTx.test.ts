import { expect } from "chai"
import { ethers } from "hardhat"
import {
  concat,
  encodeRlp,
  getAddress,
  getBytes,
  hexlify,
  keccak256,
  SigningKey,
  toBeArray,
  toQuantity,
  Wallet,
} from "ethers"

// Ethers v6.13.5 has no native type-4 support, so the wire format is built
// from ethers' RLP/minimal-int primitives.
const SET_CODE_TX_TYPE = 0x04
const AUTH_MAGIC = 0x05

type RlpItem = string | Uint8Array | RlpItem[]

function signAuthorization(
  key: SigningKey,
  chainId: bigint,
  address: string,
  nonce: bigint,
): RlpItem[] {
  const payload = encodeRlp([toBeArray(chainId), getAddress(address), toBeArray(nonce)])
  const digest = keccak256(concat([new Uint8Array([AUTH_MAGIC]), getBytes(payload)]))
  const sig = key.sign(digest)
  return [
    toBeArray(chainId),
    getAddress(address),
    toBeArray(nonce),
    toBeArray(BigInt(sig.yParity)),
    toBeArray(BigInt(sig.r)),
    toBeArray(BigInt(sig.s)),
  ]
}

describe("Eip7702SendRawTx", function () {
  it("includes a type-0x04 transaction submitted via eth_sendRawTransaction", async function () {
    const provider = ethers.provider
    const { chainId } = await provider.getNetwork()

    const [funder] = await ethers.getSigners()
    const authority = Wallet.createRandom().connect(provider)
    await (
      await funder.sendTransaction({
        to: authority.address,
        value: ethers.parseEther("0.001"),
      })
    ).wait()

    const target = getAddress("0x1111111111111111111111111111111111111111")
    const key = new SigningKey(authority.privateKey)
    const nonce = BigInt(await provider.getTransactionCount(authority.address))

    // EIP-7702: auth_nonce is the authority's nonce *after* the tx is applied.
    const auth = signAuthorization(key, chainId, target, nonce + 1n)

    const fee = await provider.getFeeData()
    const maxFeePerGas = fee.maxFeePerGas ?? ethers.parseUnits("1", "gwei")
    const maxPriorityFeePerGas =
      fee.maxPriorityFeePerGas ?? ethers.parseUnits("1", "gwei")

    const fields: RlpItem[] = [
      toBeArray(chainId),
      toBeArray(nonce),
      toBeArray(maxPriorityFeePerGas),
      toBeArray(maxFeePerGas),
      toBeArray(200_000n),
      authority.address,
      toBeArray(0n),
      "0x",
      [], // access list
      [auth],
    ]
    const unsigned = concat([
      new Uint8Array([SET_CODE_TX_TYPE]),
      getBytes(encodeRlp(fields)),
    ])
    const sig = key.sign(keccak256(unsigned))
    const raw = hexlify(
      concat([
        new Uint8Array([SET_CODE_TX_TYPE]),
        getBytes(
          encodeRlp([
            ...fields,
            toBeArray(BigInt(sig.yParity)),
            toBeArray(BigInt(sig.r)),
            toBeArray(BigInt(sig.s)),
          ]),
        ),
      ]),
    )

    const hash: string = await provider.send("eth_sendRawTransaction", [raw])

    // HardhatEthersProvider doesn't implement waitForTransaction.
    let receipt: any = null
    for (let i = 0; i < 60; i++) {
      receipt = await provider.send("eth_getTransactionReceipt", [hash])
      if (receipt) break
      await new Promise((r) => setTimeout(r, 1000))
    }
    expect(receipt, "receipt for set-code tx").to.not.be.null
    expect(receipt.status).to.equal("0x1")

    const rpcTx: any = await provider.send("eth_getTransactionByHash", [hash])
    expect(rpcTx).to.not.be.null
    expect(rpcTx.type).to.equal(toQuantity(SET_CODE_TX_TYPE))
    expect(rpcTx.authorizationList).to.have.length(1)
    expect(getAddress(rpcTx.authorizationList[0].address)).to.equal(target)
    expect(BigInt(rpcTx.authorizationList[0].chainId)).to.equal(chainId)

    // EIP-7702 delegation designator: 0xef0100 || target (23 bytes total).
    const code: string = await provider.send("eth_getCode", [authority.address, "latest"])
    expect(code.toLowerCase()).to.equal("0xef0100" + target.slice(2).toLowerCase())
  })
})
