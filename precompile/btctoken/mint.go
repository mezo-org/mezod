package btctoken

import (
	"github.com/evmos/evmos/v12/precompile"
)

type mintMethod struct {

}

func newMintMethod() *mintMethod {
	return &mintMethod{}
}

func (mm *mintMethod) MethodName() string {
	//TODO implement me
	panic("implement me")
}

func (mm *mintMethod) MethodType() precompile.MethodType {
	//TODO implement me
	panic("implement me")
}

func (mm *mintMethod) RequiredGas(methodInputArgs []byte) (uint64, bool) {
	//TODO implement me
	panic("implement me")
}

func (mm *mintMethod) Payable() bool {
	//TODO implement me
	panic("implement me")
}

func (mm *mintMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, error) {
	//TODO implement me
	panic("implement me")
}

