export type AddressItem = {
  address: string
  deployer?: string
}

export type TxItem = {
  hash: string
  from: string
  to: string
  value: string
}

export class BlockScoutAPI {
  readonly #apiUrl: string
  readonly #rateLimiter: RateLimiter

  constructor(readonly apiUrl: string) {
    this.#apiUrl = apiUrl
    this.#rateLimiter = new RateLimiter(25, 1000) // ~25 req/s (BlockScout limit is 50 req/s)
  }

  async #call(endpoint: string) {
    await this.#rateLimiter.wait()

    const response = await fetch(
      `${this.#apiUrl}/${endpoint}`,
      {
        headers: {
          "cache-control": "no-cache"
        }
      }
    )

    return response.json()
  }

  async externallyOwnedAccounts(): Promise<AddressItem[]> {
    type Item = {
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

      // We do not use tx_count fields associated with addresses returned by
      // this endpoint as they are not up-to-date. Transaction counts
      // are not updated by the indexer but instead, on demand, upon an
      // api call to addresses/<address>/counters or upon UI request that
      // is done from the address page.
      let responseJSON
      try {
        responseJSON = await this.#call(`addresses${queryString}`) as {
          items: Item[]
          next_page_params: NextPageParams
        }
      } catch (error) {
        // In case of error, break the loop and return the addresses fetched so far.
        console.error(`error fetching addresses page number: ${i+1}: ${error}`)
        break
      }

      // Filter out contracts to get EOAs. The above BlockScout endpoint
      // returns addresses that have a balance so we need to get contract
      // addresses in another way.
      const addressesPage = responseJSON.items
        .filter((item: Item) => {
          return !item.is_contract
        })
        .map((item: Item): AddressItem => {
          return {
            address: item.hash,
          }
        })

      addresses.push(...addressesPage)

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

  async txCount(address: string, filterMode?: "from" | "to"): Promise<number> {
    console.log(`fetch tx count of address ${address} in regular mode`)

    // First, try to fetch the transaction count directly from BlockScout
    // `/counters` endpoint. However, this count may not be up-to-date
    // and may return 0 for fresh addresses.
    const item = await this.#call(`addresses/${address}/counters`) as {
      transactions_count: string
    }

    // If the count is a parseable number...
    if (item.transactions_count && item.transactions_count.length > 0) {
      const txCount = parseInt(item.transactions_count)
      // ...and it is greater than 0, return it.
      if (txCount > 0) {
        return txCount
      }
    }

    // Otherwise, the count was likely not refreshed by BlockScout yet.
    // Fall back to fetching the transactions manually and counting them.

    type NextPageParams = {
      block_number: number,
      index: number,
      items_count: number
    }

    let txCount = 0

    let filterString: string = filterMode ? `filter=${filterMode}` : ""
    let queryString: string = filterString ? `?${filterString}` : ""

    for (let i = 0 ; ; i++) {
      console.log(`fetch tx count of address ${address} in fallback mode - page number: ${i + 1}`)

      const responseJSON = await this.#call(`addresses/${address}/transactions${queryString}`) as {
        items: any[]
        next_page_params: NextPageParams
      }

      txCount += responseJSON.items.length

      if (!responseJSON.next_page_params) {
        break
      }

      const {block_number, index, items_count} = responseJSON.next_page_params

      queryString =
        `?` +
        `${filterString ? filterString + "&" : ""}` +
        `block_number=${block_number}&` +
        `index=${index}&` +
        `items_count=${items_count}`
    }

    return txCount
  }

  async txs(address: string, filterMode?: "from" | "to"): Promise<TxItem[]> {
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

    type NextPageParams = {
      block_number: number,
      index: number,
      items_count: number
    }

    const txs: TxItem[] = []

    let filterString: string = filterMode ? `filter=${filterMode}` : ""
    let queryString: string = filterString ? `?${filterString}` : ""

    for (let i = 0 ; ; i++) {
      console.log(`fetch txs related to address ${address} page number: ${i + 1}`)

      const responseJSON = await this.#call(`addresses/${address}/transactions${queryString}`) as {
        items: Item[]
        next_page_params: NextPageParams
      }

      const txsPage = responseJSON.items
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

      txs.push(...txsPage)

      if (!responseJSON.next_page_params) {
        break
      }

      const {block_number, index, items_count} = responseJSON.next_page_params

      queryString =
        `?` +
        `${filterString ? filterString + "&" : ""}` +
        `block_number=${block_number}&` +
        `index=${index}&` +
        `items_count=${items_count}`
    }

    return txs
  }

  async contracts(): Promise<AddressItem[]> {
    type Item = {
      status: string
      from: {
        hash: string
      }
      created_contract: {
        hash: string
      }
    }

    type NextPageParams = {
      block_number: number,
      index: number,
      items_count: number
    }

    const contracts: AddressItem[] = []

    let queryString: string = ""
    for (let i = 0 ; ; i++) {
      console.log(`fetch contract creation txs page number: ${i + 1}`)

      let responseJSON
      try {
        responseJSON = await this.#call(`transactions?type=contract_creation${queryString}`) as {
          items: Item[]
          next_page_params: NextPageParams
        }
      } catch (error) {
        // In case of error, break the loop and return the contracts fetched so far.
        console.error(`error fetching contract creation txs page number: ${i + 1}: ${error}`)
        break
      }

      const contractsPage = responseJSON.items
        .filter((item: Item) => {
          return item.status === "ok"
        })
        .map(async (item: Item): Promise<AddressItem> => {
          return {
            address: item.created_contract.hash,
            deployer: item.from.hash,
          }
        })

      contracts.push(...await Promise.all(contractsPage))

      if (!responseJSON.next_page_params) {
        break
      }

      const {block_number, index, items_count} = responseJSON.next_page_params

      queryString = `&block_number=${block_number}&` +
        `index=${index}&` +
        `items_count=${items_count}`
    }

    return contracts
  }
}

/**
 * This is a simple rate limiter that works only if the `wait` method is
 * executed from a loop and one iteration blocks the next one.
 * It is straightforward and fits the purpose of the BlockScout API.
 * In case there is a need for a more sophisticated rate limiter, consider
 * the leaky bucket algorithm.
 */
class RateLimiter {
  readonly #requestsPerInterval: number
  readonly #intervalTime: number

  #queuedRequests = 0

  constructor(
    requestsPerInterval: number,
    intervalTime: number,
  ) {
    this.#requestsPerInterval = requestsPerInterval;
    this.#intervalTime = intervalTime;
  }

  public async wait() {
    let timeout = 0

    if (this.#queuedRequests >= this.#requestsPerInterval) {
      timeout = this.#intervalTime
      this.#queuedRequests = 0
    }

    return new Promise((resolve) => {
      setTimeout(
        () => {
          this.#queuedRequests++
          resolve(() => {})
        },
        timeout
      )
    });
  }
}