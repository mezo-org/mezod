package precompile

type mockMethod struct {
	methodName  string
	methodType  MethodType
	requiredGas uint64
	payable     bool

	run	func(
		context *RunContext,
		inputs MethodInputs,
	) (MethodOutputs, error)
}

func (mm *mockMethod) MethodName() string {
	return mm.methodName
}

func (mm *mockMethod) MethodType() MethodType {
	return mm.methodType
}

func (mm *mockMethod) RequiredGas(methodInputArgs []byte) (uint64, bool) {
	if mm.requiredGas == 0 {
		return 0, false
	}

	return mm.requiredGas, true
}

func (mm *mockMethod) Payable() bool {
	return mm.payable
}

func (mm *mockMethod) Run(
	context *RunContext,
	inputs MethodInputs,
) (MethodOutputs, error) {
	return mm.run(context, inputs)
}
