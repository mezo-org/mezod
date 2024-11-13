import { WorkerEntrypoint } from "cloudflare:workers";

type Env = {
  BLOCKSCOUT_API_URL: string
}

export default {
  async fetch(request: Request, env: Env, _: ExecutionContext) {
    return Response.json({ message: "Hello, world!" })
  },
}

export class InternalEntrypoint extends WorkerEntrypoint {
  async get(): Promise<InternalGetResponse> {
    try {
      // TODO: Implement the logic for fetching activity data.
      return { success: true, activity: [] }
    } catch (error) {
      return { success: false, errorMsg: `${error}` }
    }
  }
}

export type InternalGetResponse = {
  success: boolean
  activity?: ActivityItem[]
  errorMsg?: string
}

export type ActivityItem = {
  address: string
  txCount: number,
  deployedContracts: number,
  deployedContractsTxCount: number
}
