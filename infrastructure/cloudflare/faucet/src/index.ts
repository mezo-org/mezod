import { AutoRouter, cors, html, IRequest } from "itty-router"
import { ethers } from "ethers"
import { indexHTML, errorHTML, successHTML } from "#/assets"

const cfVerifyUrl = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

type Env = {
  MEZO_API_URL: string
  MEZO_FAUCET_PRIVATE_KEY: string
  TURNSTILE_SITE_KEY: string
  TURNSTILE_SECRET_KEY: string
  AMOUNT_BTC: string
  RATE_LIMITER: any
  REQUEST_DELAY_SECONDS: number
}

const sendBTC = async (request: Request, env: Env) => {
  const requestForm = await request.formData()
  const targetAddress = requestForm.get("address") as string
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

  if (!ethers.isAddress(targetAddress)) {
    return html(errorHTML("Invalid target address. Try to use a proper 20-byte hexadecimal address prefixed with 0x."))
  }

  const rl = await rateLimit(env, ip)
  if (!rl.success) {
    const leftMinutes = Math.ceil(rl.left! / 60)
    return html(errorHTML(`Rate limit exceeded. Try again after ${leftMinutes} min.`))
  }

  try {
    const wallet = new ethers.Wallet(
      env.MEZO_FAUCET_PRIVATE_KEY,
      ethers.getDefaultProvider(env.MEZO_API_URL)
    )

    const amountBTC = env.AMOUNT_BTC

    const transaction = await wallet.sendTransaction({
      to: targetAddress,
      value: ethers.parseEther(amountBTC),
    })

    return html(successHTML(transaction.hash, amountBTC))
  } catch (error) {
    return html(errorHTML(`Unexpected error: ${error}`))
  }
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

const { preflight, corsify } = cors()

const router = AutoRouter({
  before: [preflight],
  finally: [corsify],
})

router
  .post("/", sendBTC)
  .all("*", (_: IRequest, env: Env) => html(indexHTML(env.TURNSTILE_SITE_KEY)))

export default {
  async fetch(request: Request, env: Env, _: ExecutionContext) {
    return router.fetch(request, env)
  },
}
