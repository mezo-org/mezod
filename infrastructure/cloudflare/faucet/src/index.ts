import { AutoRouter, cors, error, html, IRequest, status, json } from "itty-router"
import { ethers } from "ethers"
import { indexHTML, errorHTML, successHTML } from "#/assets"
import { WorkerEntrypoint } from "cloudflare:workers";

const cfVerifyUrl = "https://challenges.cloudflare.com/turnstile/v0/siteverify"
const invalidTargetAddressError = "Invalid target address. Try to use a proper 20-byte hexadecimal address prefixed with 0x."
const mezoTokenAddress = "0x7b7c000000000000000000000000000000000001"

type Env = {
  MEZO_API_URL: string
  MEZO_FAUCET_PRIVATE_KEY: string
  TURNSTILE_SITE_KEY: string
  TURNSTILE_SECRET_KEY: string
  AMOUNT_BTC: string
  AMOUNT_MEZO: string
  RATE_LIMITER: any
  REQUEST_DELAY_SECONDS: number
  PUBLIC_ACCESS: string
  PUBLIC_ACCESS_REDIRECT: string
  API_KEY: string
}

type TokenType = "BTC" | "MEZO"

const ERC20_TRANSFER_ABI = [
  "function transfer(address to, uint256 amount) returns (bool)"
]

const publicSend = async (request: Request, env: Env) => {
  const requestForm = await request.formData()
  const targetAddress = requestForm.get("address") as string
  const token = (requestForm.get("token") as TokenType) || "BTC"
  const ip = request.headers.get('cf-connecting-ip') as string

  let cfVerifyForm = new FormData()
  cfVerifyForm.append('secret', env.TURNSTILE_SECRET_KEY)
  cfVerifyForm.append('response', requestForm.get("cf-turnstile-response") as string)
  cfVerifyForm.append('remoteip', ip)

  const cfVerifyResult = await fetch(cfVerifyUrl, {
    body: cfVerifyForm,
    method: 'POST',
  })

  // @ts-ignore
  if (!(await cfVerifyResult.json()).success) {
    return html(errorHTML("Captcha verification failed"))
  }

  // Early check the target address validity. Despite internalSend checks it
  // again, we don't want to waste rate limit on invalid addresses that may be
  // wrong by mistake.
  if (!ethers.isAddress(targetAddress)) {
    return html(errorHTML(invalidTargetAddressError))
  }

  const rl = await rateLimit(env, ip)
  if (!rl.success) {
    const leftMinutes = Math.ceil(rl.left! / 60)
    return html(errorHTML(`Rate limit exceeded. Try again after ${leftMinutes} min.`))
  }

  try {
    if (token === "MEZO") {
      const amountMEZO = env.AMOUNT_MEZO
      const transactionHash = await internalSendMEZO(env, targetAddress, amountMEZO)
      return html(successHTML(transactionHash, amountMEZO, "MEZO"))
    } else {
      const amountBTC = env.AMOUNT_BTC
      const transactionHash = await internalSendBTC(env, targetAddress, amountBTC)
      return html(successHTML(transactionHash, amountBTC, "BTC"))
    }
  } catch (error) {
    return html(errorHTML(`Unexpected error: ${error}`))
  }
}

async function internalSendBTC(
  env: Env,
  targetAddress: string,
  amountBTC: string
): Promise<string> {
  if (!ethers.isAddress(targetAddress)) {
    throw Error(invalidTargetAddressError)
  }

  let parsedAmountBTC: bigint
  try {
    parsedAmountBTC = ethers.parseEther(amountBTC)
  } catch (error) {
    throw Error(`Invalid BTC amount: ${error}`)
  }

  const wallet = new ethers.Wallet(
    env.MEZO_FAUCET_PRIVATE_KEY,
    ethers.getDefaultProvider(env.MEZO_API_URL)
  )

  const transaction = await wallet.sendTransaction({
    to: targetAddress,
    value: parsedAmountBTC,
  })

  return transaction.hash
}

async function internalSendMEZO(
  env: Env,
  targetAddress: string,
  amountMEZO: string
): Promise<string> {
  if (!ethers.isAddress(targetAddress)) {
    throw Error(invalidTargetAddressError)
  }

  let parsedAmountMEZO: bigint
  try {
    parsedAmountMEZO = ethers.parseEther(amountMEZO)
  } catch (error) {
    throw Error(`Invalid MEZO amount: ${error}`)
  }

  const wallet = new ethers.Wallet(
    env.MEZO_FAUCET_PRIVATE_KEY,
    ethers.getDefaultProvider(env.MEZO_API_URL)
  )

  const mezoToken = new ethers.Contract(
    mezoTokenAddress,
    ERC20_TRANSFER_ABI,
    wallet
  )

  const transaction = await mezoToken.transfer(targetAddress, parsedAmountMEZO)

  return transaction.hash
}

async function rateLimit(env: Env, ip: string): Promise<{
  success: boolean
  left?: number
}> {
  const now = Math.floor(Date.now() / 1000)
  const key = `rate-limiter:${ip}`

  // Get the timestamp when the next request is allowed.
  const nextRequestTimestamp: number | undefined = await env.RATE_LIMITER.get(key)

  if (nextRequestTimestamp && now < nextRequestTimestamp) {
    // The next request is not allowed yet.
    return { success: false, left: Number(nextRequestTimestamp) - Number(now) }
  }

  // Request is either allowed or the rate limiter is not initialized for this IP.
  // Set the next request timestamp to the current time plus the delay.
  const newTimestamp = Number(now) + Number(env.REQUEST_DELAY_SECONDS)
  await env.RATE_LIMITER.put(key, newTimestamp)
  return { success: true }
}

const isPublic = (env: Env) => env.PUBLIC_ACCESS === "true"

const { preflight, corsify } = cors()

const router = AutoRouter({
  before: [preflight],
  finally: [corsify],
})

router
  .post("/", (request: IRequest, env: Env) =>
    isPublic(env) ?
      publicSend(request, env) :
      error(403, "Forbidden")
  )
  .post("/internal", async (request: Request, env: Env) => {
    const authorization = request.headers.get("Authorization") as string
    if (!authorization || !authorization.startsWith("Basic ")) {
      return error(401, "Unauthorized")
    }

    const apiKey = authorization.replace("Basic ", "");
    if (apiKey !== env.API_KEY) {
      return error(403, "Forbidden")
    }

    const requestBody: { targetAddress: string; amount: string; token?: TokenType } = await request.json()
    const { targetAddress, amount, token = "BTC" } = requestBody

    try {
      let transactionHash: string
      if (token === "MEZO") {
        transactionHash = await internalSendMEZO(env, targetAddress, amount)
      } else {
        transactionHash = await internalSendBTC(env, targetAddress, amount)
      }
      return json({ success: true, transactionHash })
    } catch (error) {
      return json({ success: false, errorMsg: `${error}` })
    }
  })
  .all("*", (request: IRequest, env: Env) =>
    isPublic(env) ?
      html(indexHTML(env.TURNSTILE_SITE_KEY)) :
      status(302, {headers: {Location: env.PUBLIC_ACCESS_REDIRECT}})
  )

export default {
  async fetch(request: Request, env: Env, _: ExecutionContext) {
    return router.fetch(request, env)
  },
}

export class InternalEntrypoint extends WorkerEntrypoint {
  async send(targetAddress: string, amount: string, token: TokenType = "BTC"): Promise<InternalSendResponse> {
    try {
      let transactionHash: string
      if (token === "MEZO") {
        transactionHash = await internalSendMEZO(this.env as Env, targetAddress, amount)
      } else {
        transactionHash = await internalSendBTC(this.env as Env, targetAddress, amount)
      }
      return { success: true, transactionHash }
    } catch (error) {
      return { success: false, errorMsg: `${error}` }
    }
  }
}

export type InternalSendResponse = {
  success: boolean
  transactionHash?: string
  errorMsg?: string
}