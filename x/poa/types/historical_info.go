package types

import (
	"bytes"
	"sort"

	"github.com/cosmos/cosmos-sdk/codec"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

// NewHistoricalInfo will create a historical information struct from header and
// valset it will first sort valset before inclusion into historical info
func NewHistoricalInfo(header tmproto.Header, valSet []Validator) HistoricalInfo {
	// Sort in the same way that Tendermint does. Tendermint sorts by
	// voting power and address. In the PoA module, all validators have the
	// same voting power, so we sort by address only.
	sort.SliceStable(valSet, func(i, j int) bool {
		return bytes.Compare(
			valSet[i].GetOperator(),
			valSet[j].GetOperator(),
		) == -1
	})

	return HistoricalInfo{
		Header: header,
		Valset: valSet,
	}
}

// MustUnmarshalHistoricalInfo wll unmarshal historical info and panic on error
func MustUnmarshalHistoricalInfo(
	cdc codec.BinaryCodec,
	value []byte,
) HistoricalInfo {
	hi, err := UnmarshalHistoricalInfo(cdc, value)
	if err != nil {
		panic(err)
	}

	return hi
}

// UnmarshalHistoricalInfo will unmarshal historical info and return any error
func UnmarshalHistoricalInfo(
	cdc codec.BinaryCodec,
	value []byte,
) (hi HistoricalInfo, err error) {
	err = cdc.Unmarshal(value, &hi)
	return hi, err
}
