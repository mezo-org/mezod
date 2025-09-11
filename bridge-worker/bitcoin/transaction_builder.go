package bitcoin

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

// internalTransaction is an internal utility representation of the Transaction
// that expose a lot of tools helpful during transaction manipulation.
type internalTransaction struct {
	*wire.MsgTx
}

func newInternalTransaction() *internalTransaction {
	msgTx := wire.NewMsgTx(wire.TxVersion)
	msgTx.LockTime = 0

	return &internalTransaction{msgTx}
}

func (it *internalTransaction) fromTransaction(transaction *Transaction) {
	it.Version = transaction.Version

	it.TxIn = make([]*wire.TxIn, len(transaction.Inputs))
	for i, input := range transaction.Inputs {
		it.TxIn[i] = &wire.TxIn{
			PreviousOutPoint: wire.OutPoint{
				Hash:  chainhash.Hash(input.Outpoint.TransactionHash),
				Index: input.Outpoint.OutputIndex,
			},
			SignatureScript: input.SignatureScript,
			Witness:         input.Witness,
			Sequence:        input.Sequence,
		}
	}

	it.TxOut = make([]*wire.TxOut, len(transaction.Outputs))
	for i, output := range transaction.Outputs {
		it.TxOut[i] = &wire.TxOut{
			Value:    output.Value,
			PkScript: output.PublicKeyScript,
		}
	}

	it.LockTime = transaction.Locktime
}
