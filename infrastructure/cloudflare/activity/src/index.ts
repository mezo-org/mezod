import { WorkerEntrypoint } from "cloudflare:workers"
import { BlockScoutAPI, ContractItem } from "#/blockscout"

type Env = {
  BLOCKSCOUT_API_URL: string
  FAUCET_ADDRESS: string
}

async function fetchActivity(env: Env): Promise<ActivityItem[]> {
  const bsAPI = new BlockScoutAPI(env.BLOCKSCOUT_API_URL)

  const addresses = await bsAPI.addresses()
  const contracts = await bsAPI.contracts()
  const txsFromFaucet = await bsAPI.txsFromAddress(env.FAUCET_ADDRESS)

  const addressToContracts = contracts.reduce(
    (acc, contract) => {
      if (!acc[contract.deployer]) {
        acc[contract.deployer] = []
      }
      acc[contract.deployer].push(contract)
      return acc
    },
    {} as Record<string, ContractItem[]>
  )

  const addressToClaimedBTC = txsFromFaucet.reduce(
    (acc, tx) => {
      if (!acc[tx.to]) {
        acc[tx.to] = BigInt(0)
      }
      acc[tx.to] += BigInt(tx.value)
      return acc
    },
    {} as Record<string, bigint>
  )

  return addresses.map((item) => {
    const contracts = addressToContracts[item.address] || []

    const deployedContractsTxCount = contracts.reduce(
      (acc, contract) => acc + contract.txCount,
      0
    )

    return {
      address: item.address,
      txCount: item.txCount,
      deployedContracts: contracts.length,
      deployedContractsTxCount: deployedContractsTxCount,
      claimedBTC: addressToClaimedBTC[item.address]?.toString() || "0"
    }
  })
}

export default {
  async fetch(request: Request, env: Env, _: ExecutionContext) {
    // TODO: Return for testing purposes. Return a 404 ultimately as we want to rely on internal RPC only.
    return Response.json({
      success: true,
      activity: await fetchActivity(env),
    })
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
          claimedBTC: "1000000000000000000"
        },
        {
          address: "0x7964f3985F1eA5E7f58b3aa44deCe2de7468d51B",
          txCount: 1,
          deployedContracts: 0,
          deployedContractsTxCount: 0,
          claimedBTC: "1000000000000000000"
        },
        {
          address: "0x9e0F855D302C633cF78F721A5c0A419D38EE949A",
          txCount: 500,
          deployedContracts: 20,
          deployedContractsTxCount: 3000,
          claimedBTC: "1000000000000000000"
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
  claimedBTC: string // 1e18 precision
}
