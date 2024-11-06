import { AutoRouter, error, json, cors } from "itty-router"

const baseApi = "/api"

type Env = {
  ASSETS: any
}

const sendBTC = async (request: Request, env: Env) => {
  const formData = await request.formData()
  const address = formData.get("address")
  const turnstile = formData.get("cf-turnstile-response")
  return json({ address, turnstile })
}

const { preflight, corsify } = cors()

const router = AutoRouter({
  base: baseApi,
  before: [preflight],
  finally: [corsify],
})

router
  .post("/", sendBTC)
  .all("*", () => error(404, "Not Found"))

export default {
  async fetch(request: Request, env: Env, context: ExecutionContext) {
    const url = new URL(request.url);
    if (url.pathname.startsWith(baseApi)) {
      return router.fetch(request, env);
    }

    // Passes the incoming request through to the assets binding.
    // No asset matched this request before the worker code was invoked,
    // so this will evaluate `not_found_handling` behavior.
    return env.ASSETS.fetch(request);
  },
}
