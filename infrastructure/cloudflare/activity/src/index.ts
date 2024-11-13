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

      const mockData: ActivityItem[] = [
        {
          address: "0x6d2C29C7E7C206dd7A0F2D47C79Ab828E6546d7C",
          txCount: 10,
          deployedContracts: 5,
          deployedContractsTxCount: 100,
        },
        {
          address: "0x7964f3985F1eA5E7f58b3aa44deCe2de7468d51B",
          txCount: 1,
          deployedContracts: 0,
          deployedContractsTxCount: 0,
        },
        {
          address: "0x9e0F855D302C633cF78F721A5c0A419D38EE949A",
          txCount: 500,
          deployedContracts: 20,
          deployedContractsTxCount: 3000,
        }
      ]

      return { success: true, activity: mockData }
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
