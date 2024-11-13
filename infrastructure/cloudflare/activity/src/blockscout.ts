export type AddressItem = {
  address: string
  txCount: number
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

    const responseJSON = await this.#call("addresses") as { items: Item[] }
    const items = responseJSON.items

    return items
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
  }
}