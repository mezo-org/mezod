package bitcoin

import (
	"github.com/btcsuite/btcd/txscript"
)

// Script represents an arbitrary Bitcoin script, NOT prepended with the
// byte-length of the script
type Script []byte

// PayToPublicKeyHash constructs a P2PKH script for the provided 20-byte public
// key hash. The function assumes the provided public key hash is valid.
func PayToPublicKeyHash(publicKeyHash [20]byte) (Script, error) {
	return txscript.NewScriptBuilder().
		AddOp(txscript.OP_DUP).
		AddOp(txscript.OP_HASH160).
		AddData(publicKeyHash[:]).
		AddOp(txscript.OP_EQUALVERIFY).
		AddOp(txscript.OP_CHECKSIG).
		Script()
}

// PayToWitnessPublicKeyHash constructs a P2WPKH script for the provided
// 20-byte public key hash. The function assumes the provided public key hash
// is valid.
func PayToWitnessPublicKeyHash(publicKeyHash [20]byte) (Script, error) {
	return txscript.NewScriptBuilder().
		AddOp(txscript.OP_0).
		AddData(publicKeyHash[:]).
		Script()
}
