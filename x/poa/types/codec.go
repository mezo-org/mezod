package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types on codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgSubmitApplication{}, "poa/MsgSubmitApplication", nil)
	cdc.RegisterConcrete(MsgVote{}, "poa/MsgVote", nil)
	cdc.RegisterConcrete(MsgProposeKick{}, "poa/MsgProposeKick", nil)
	cdc.RegisterConcrete(MsgLeaveValidatorSet{}, "poa/MsgLeaveValidatorSet", nil)
}

// ModuleCdc defines the module codec
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
