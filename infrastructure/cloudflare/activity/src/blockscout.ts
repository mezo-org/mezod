export type AddressItem = {
  address: string
  txCount: number
}

export type ContractItem = {
  address: string
  deployer: string
  txCount: number
}

export type TxItem = {
  hash: string
  from: string
  to: string
  value: string
}

export class BlockScoutAPI {
  readonly #apiUrl: string

  constructor(readonly apiUrl: string) {
    this.#apiUrl = apiUrl
  }

  async #call(endpoint: string) {
    const response = await fetch(`${this.#apiUrl}/${endpoint}`)
    return response.json()
  }

  async addresses(): Promise<AddressItem[]> {
    type Item = {
      tx_count: string
      hash: string
      is_contract: boolean
    }

    type NextPageParams = {
      fetched_coin_balance: string,
      hash: string,
      items_count: number
    }

    const addresses: AddressItem[] = []

    let queryString: string = ""
    for (let i = 0 ; ; i++) {
      console.log(`fetch addresses page number: ${i+1}`)

      const responseJSON = await this.#call(`addresses${queryString}`) as {
        items: Item[]
        next_page_params: NextPageParams
      }

      const batch = responseJSON.items
        .filter((item: Item) => {
          const hasTxs = item.tx_count.length > 0 && item.tx_count !== "0"
          return hasTxs && !item.is_contract
        })
        .map((item: Item): AddressItem => {
          return {
            address: item.hash,
            txCount: parseInt(item.tx_count),
          }
        })

      addresses.push(...batch)

      if (!responseJSON.next_page_params) {
        break
      }

      const {fetched_coin_balance, hash, items_count} = responseJSON.next_page_params

      queryString = `?fetched_coin_balance=${fetched_coin_balance}&` +
        `hash=${hash}&` +
        `items_count=${items_count}`
    }

    return addresses
  }

  async txCount(address: string): Promise<number> {
    type Item = {
      transactions_count: string
    }

    const item = await this.#call(`addresses/${address}/counters`) as Item

    if (item.transactions_count.length === 0) {
      return 0
    }

    return parseInt(item.transactions_count)
  }

  async txsFromAddress(address: string): Promise<TxItem[]> {
    type Item = {
      hash: string
      status: string
      from: {
        hash: string
      }
      to: {
        hash: string
      },
      value: string
    }

    const responseJSON = await this.#call(`addresses/${address}/transactions?filter=from`) as { items: Item[] }
    const items = responseJSON.items

    return items
      .filter((item: Item) => {
        return item.status === "ok"
      })
      .map((item: Item): TxItem => {
        return {
          hash: item.hash,
          from: item.from.hash,
          to: item.to.hash,
          value: item.value,
        }
      })
  }

  async contracts(): Promise<ContractItem[]> {
    type Item = {
      status: string
      from: {
        hash: string
      }
      created_contract: {
        hash: string
      }
    }

    const responseJSON = await this.#call("transactions?type=contract_creation") as { items: Item[] }
    const items = responseJSON.items

    const contracts = items
      .filter((item: Item) => {
        return item.status === "ok"
      })
      .map(async (item: Item): Promise<ContractItem> => {
        const contractAddress = item.created_contract.hash

        const txCount = await this.txCount(contractAddress)

        return {
          address: contractAddress,
          deployer: item.from.hash,
          txCount,
        }
      })

    return Promise.all(contracts)
  }
}