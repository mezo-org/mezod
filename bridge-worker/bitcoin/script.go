package bitcoin

import (
	"fmt"

	"github.com/btcsuite/btcd/txscript"
)

// Script represents an arbitrary Bitcoin script, NOT prepended with the
// byte-length of the script
type Script []byte

// NewScriptFromVarLenData construct a Script instance based on the provided
// variable length data prepended with a CompactSizeUint.
func NewScriptFromVarLenData(varLenData []byte) (Script, error) {
	// Extract the CompactSizeUint value that holds the byte length of the script.
	// Also, extract the byte length of the CompactSizeUint itself.
	scriptByteLength, compactByteLength, err := readCompactSizeUint(varLenData)
	if err != nil {
		return nil, fmt.Errorf("cannot read compact size uint: [%v]", err)
	}

	// Make sure the combined byte length of the script and the byte length
	// of the CompactSizeUint matches the total byte length of the variable
	// length data. Otherwise, the input data slice is malformed.
	// #nosec G115
	if uint64(scriptByteLength)+uint64(compactByteLength) != uint64(len(varLenData)) {
		return nil, fmt.Errorf("malformed var len data")
	}

	// Extract the actual script by omitting the leading CompactSizeUint.
	return varLenData[compactByteLength:], nil
}

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

// ToVarLenData converts the Script to a byte array prepended with a
// CompactSizeUint holding the script's byte length.
func (s Script) ToVarLenData() ([]byte, error) {
	compactBytes, err := writeCompactSizeUint(CompactSizeUint(len(s)))
	if err != nil {
		return nil, fmt.Errorf("cannot write compact size uint: [%v]", err)
	}

	return append(compactBytes, s...), nil
}
