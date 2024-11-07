import { AutoRouter, cors, html, IRequest } from "itty-router"
import { ethers } from "ethers"
import { indexHTML, errorHTML, successHTML } from "#/assets.ts";

const cfVerifyUrl = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

type Env = {
  MEZO_API_URL: string
  MEZO_FAUCET_PRIVATE_KEY: string
  TURNSTILE_SITE_KEY: string
  TURNSTILE_SECRET_KEY: string
}

const sendBTC = async (request: Request, env: Env) => {
  const requestForm = await request.formData()
  const targetAddress = requestForm.get("address") as string

  let cfVerifyForm = new FormData();
  cfVerifyForm.append('secret', env.TURNSTILE_SECRET_KEY);
  cfVerifyForm.append('response', requestForm.get("cf-turnstile-response") as string);
  cfVerifyForm.append('remoteip', request.headers.get('CF-Connecting-IP') as string);

  const cfVerifyResult = await fetch(cfVerifyUrl, {
    body: cfVerifyForm,
    method: 'POST',
  });

  // @ts-ignore
  if (!(await cfVerifyResult.json()).success) {
    return html(errorHTML("Captcha verification failed"))
  }

  if (!ethers.isAddress(targetAddress)) {
    return html(errorHTML("Invalid target address. Try to use a proper 20-byte hexadecimal address prefixed with 0x."))
  }

  try {
    const wallet = new ethers.Wallet(
      env.MEZO_FAUCET_PRIVATE_KEY,
      ethers.getDefaultProvider(env.MEZO_API_URL)
    )

    const transaction = await wallet.sendTransaction({
      to: targetAddress,
      value: ethers.parseEther("0.001")
    });

    return html(successHTML(transaction.hash))
  } catch (error) {
    return html(errorHTML(`Unexpected error: ${error}`))
  }
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
