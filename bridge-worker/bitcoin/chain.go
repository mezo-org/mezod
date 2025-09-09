package bitcoin

// Chain defines an interface meant to be used for interaction with the
// Bitcoin chain.
type Chain interface {
	// GetTransaction gets the transaction with the given transaction hash.
	// If the transaction with the given hash was not found on the chain,
	// this function returns an error.
	GetTransaction(transactionHash Hash) (*Transaction, error)

	// GetTxHashesForPublicKeyHash gets hashes of confirmed transactions that pays
	// the given public key hash using either a P2PKH or P2WPKH script. The returned
	// transactions hashes are ordered by block height in the ascending order, i.e.
	// the latest transaction hash is at the end of the list. The returned list does
	// not contain unconfirmed transactions hashes living in the mempool at the
	// moment of request.
	GetTxHashesForPublicKeyHash(
		publicKeyHash [20]byte,
	) ([]Hash, error)
}
