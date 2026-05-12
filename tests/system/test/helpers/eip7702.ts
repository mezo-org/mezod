import { ethers } from "hardhat"
import {
  Authorization,
  AuthorizationLike,
  BaseWallet,
  HDNodeWallet,
  Signature,
  TransactionRequest,
  Wallet,
} from "ethers"

export const SET_CODE_TX_TYPE = 0x04
export const DELEGATION_PREFIX = "0xef0100"
export const EMPTY_CODE = "0x"

export type Eip7702AuthInput = {
  chainId: bigint
  address: string
  nonce: bigint
}

export type SendSetCodeTxArgs = {
  to: string
  data?: string
  value?: bigint
  gasLimit?: bigint
  authorizationList: AuthorizationLike[]
  maxFeePerGas?: bigint
  maxPriorityFeePerGas?: bigint
}

export type SendSetCodeTxResult = {
  hash: string
  receipt: any
  rpcTx: any
}

export function expectedDelegationCode(target: string): string {
  return (DELEGATION_PREFIX + target.slice(2)).toLowerCase()
}

// Wraps wallet.authorize so callers can opt into a corrupted-signature
// variant via `corruptR` without reaching for private ethers internals.
export async function signAuthorization(
  wallet: BaseWallet,
  input: Eip7702AuthInput,
  options: { corruptR?: boolean } = {},
): Promise<AuthorizationLike> {
  const auth = await wallet.authorize({
    address: input.address,
    nonce: input.nonce,
    chainId: input.chainId,
  })
  if (options.corruptR) {
    return corruptAuthorizationSignature(auth)
  }
  return auth
}

// Flips one byte of `r` so signature recovery yields a non-authority
// address (or fails), which the keeper silently skips per EIP-7702.
function corruptAuthorizationSignature(
  auth: Authorization,
): AuthorizationLike {
  const rBytes = ethers.getBytes(auth.signature.r)
  rBytes[0] ^= 0xff
  const r = ethers.hexlify(rBytes)
  return {
    address: auth.address,
    nonce: auth.nonce,
    chainId: auth.chainId,
    signature: Signature.from({
      r,
      s: auth.signature.s,
      v: auth.signature.v,
    }),
  }
}

// Bypasses sponsor.sendTransaction's broadcastTransaction → wait() chain
// because HardhatEthersProvider does not implement waitForTransaction.
// The RPC-formatted tx envelope is fetched alongside the receipt so
// callers can assert on the JSON shape (type byte, authorizationList).
export async function sendSetCodeTx(
  sponsor: BaseWallet,
  args: SendSetCodeTxArgs,
): Promise<SendSetCodeTxResult> {
  const provider = ethers.provider
  const { chainId } = await provider.getNetwork()

  const sponsorAddr = await sponsor.getAddress()
  const nonce = await provider.getTransactionCount(sponsorAddr)

  let maxFeePerGas = args.maxFeePerGas
  let maxPriorityFeePerGas = args.maxPriorityFeePerGas
  if (maxFeePerGas === undefined || maxPriorityFeePerGas === undefined) {
    const fee = await provider.getFeeData()
    maxFeePerGas ??= fee.maxFeePerGas ?? ethers.parseUnits("1", "gwei")
    maxPriorityFeePerGas ??=
      fee.maxPriorityFeePerGas ?? ethers.parseUnits("1", "gwei")
  }

  const populated: TransactionRequest = {
    type: SET_CODE_TX_TYPE,
    chainId,
    nonce,
    to: args.to,
    data: args.data ?? "0x",
    value: args.value ?? 0n,
    gasLimit: args.gasLimit ?? 1_000_000n,
    maxFeePerGas,
    maxPriorityFeePerGas,
    authorizationList: args.authorizationList,
  }

  const raw = await sponsor.signTransaction(populated)
  const hash: string = await provider.send("eth_sendRawTransaction", [raw])

  const receipt = await pollForReceipt(hash)
  const rpcTx: any = await provider.send("eth_getTransactionByHash", [hash])
  return { hash, receipt, rpcTx }
}

// Loops eth_getTransactionReceipt up to 60s. Returns null only if the tx
// never lands; callers assert non-null so the failure message points to
// the specific suite.
export async function pollForReceipt(hash: string): Promise<any> {
  for (let i = 0; i < 60; i++) {
    const receipt: any = await ethers.provider.send(
      "eth_getTransactionReceipt",
      [hash],
    )
    if (receipt) {
      return receipt
    }
    await new Promise((r) => setTimeout(r, 1000))
  }
  return null
}

export async function fundWallet(
  recipient: string,
  value: bigint,
  fromSignerIndex: number = 0,
): Promise<void> {
  const signers = await ethers.getSigners()
  const funder = signers[fromSignerIndex]
  const tx = await funder.sendTransaction({ to: recipient, value })
  await pollForReceipt(tx.hash)
}

// Creates a fresh sponsor wallet funded with enough BTC to pay for a
// type-0x04 envelope plus the auth-tuple intrinsic charge.
export async function freshSponsor(
  value: bigint = ethers.parseEther("0.05"),
): Promise<HDNodeWallet> {
  const wallet = Wallet.createRandom().connect(ethers.provider)
  await fundWallet(wallet.address, value)
  return wallet
}

// Creates a fresh authority wallet. If `funded` is true, sends 1 wei so
// the state-trie node exists; refund accounting depends on
// state.Exist(authority).
export async function freshAuthority(
  options: { funded?: boolean } = {},
): Promise<HDNodeWallet> {
  const wallet = Wallet.createRandom().connect(ethers.provider)
  if (options.funded) {
    await fundWallet(wallet.address, 1n)
  }
  return wallet
}
