package mezotoken_test

import (
	"math/big"

	sdkmath "cosmossdk.io/math"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/mezo-org/mezod/precompile"
	"github.com/mezo-org/mezod/x/evm/statedb"
)

func (s *PrecompileTestSuite) TestGetMinterWhenNotSet() {
	method := s.mezoPrecompile.Abi.Methods["getMinter"]

	methodInputArgs, err := method.Inputs.Pack()
	s.Require().NoError(err)

	vmContract := vm.NewContract(&precompile.Contract{}, nil, nil, 0)
	//nolint:gocritic
	vmContract.Input = append(method.ID, methodInputArgs...)

	evm := &vm.EVM{
		StateDB: statedb.New(s.ctx, statedb.NewMockKeeper(), statedb.TxConfig{}),
	}

	output, err := s.mezoPrecompile.Run(evm, vmContract, false)
	s.Require().NoError(err)

	out, err := method.Outputs.Unpack(output)
	s.Require().NoError(err)
	s.Require().Len(out, 1)

	returnedAddress, ok := out[0].(common.Address)
	s.Require().True(ok)

	// Should return zero address when minter is not set
	s.Require().Equal(common.Address{}, returnedAddress)
}

func (s *PrecompileTestSuite) TestSetMinterByOwner() {
	// Set minter
	method := s.mezoPrecompile.Abi.Methods["setMinter"]

	methodInputArgs, err := method.Inputs.Pack(s.minter)
	s.Require().NoError(err)

	stateDB := statedb.New(s.ctx, s.app.EvmKeeper, statedb.TxConfig{})
	vmContract := vm.NewContract(&precompile.Contract{}, nil, nil, 0)
	//nolint:gocritic
	vmContract.Input = append(method.ID, methodInputArgs...)
	vmContract.CallerAddress = s.poaOwner

	evm := &vm.EVM{
		StateDB: stateDB,
	}

	_, err = s.mezoPrecompile.Run(evm, vmContract, false)
	s.Require().NoError(err)

	// Commit the statedb to persist changes
	err = stateDB.Commit()
	s.Require().NoError(err)

	// Verify minter was set by checking params
	params := s.app.EvmKeeper.GetParams(s.ctx)
	s.Require().Equal(s.minter.Hex(), params.MezoMinterAddress)
}

func (s *PrecompileTestSuite) TestSetMinterByNonOwner() {
	method := s.mezoPrecompile.Abi.Methods["setMinter"]

	methodInputArgs, err := method.Inputs.Pack(s.minter)
	s.Require().NoError(err)

	vmContract := vm.NewContract(&precompile.Contract{}, nil, nil, 0)
	//nolint:gocritic
	vmContract.Input = append(method.ID, methodInputArgs...)
	vmContract.CallerAddress = s.unauthorizedAddr

	evm := &vm.EVM{
		StateDB: statedb.New(s.ctx, statedb.NewMockKeeper(), statedb.TxConfig{}),
	}

	output, err := s.mezoPrecompile.Run(evm, vmContract, false)
	s.Require().Error(err)
	s.Require().Nil(output)
	s.Require().ErrorContains(err, "unauthorized")
}

func (s *PrecompileTestSuite) TestGetMinterAfterSet() {
	// First set minter
	setMethod := s.mezoPrecompile.Abi.Methods["setMinter"]
	methodInputArgs, err := setMethod.Inputs.Pack(s.minter)
	s.Require().NoError(err)

	vmContract := vm.NewContract(&precompile.Contract{}, nil, nil, 0)
	//nolint:gocritic
	vmContract.Input = append(setMethod.ID, methodInputArgs...)
	vmContract.CallerAddress = s.poaOwner

	evm := &vm.EVM{
		StateDB: statedb.New(s.ctx, statedb.NewMockKeeper(), statedb.TxConfig{}),
	}

	_, err = s.mezoPrecompile.Run(evm, vmContract, false)
	s.Require().NoError(err)

	// Now get minter
	getMethod := s.mezoPrecompile.Abi.Methods["getMinter"]
	methodInputArgs, err = getMethod.Inputs.Pack()
	s.Require().NoError(err)

	vmContract = vm.NewContract(&precompile.Contract{}, nil, nil, 0)
	//nolint:gocritic
	vmContract.Input = append(getMethod.ID, methodInputArgs...)

	output, err := s.mezoPrecompile.Run(evm, vmContract, false)
	s.Require().NoError(err)

	out, err := getMethod.Outputs.Unpack(output)
	s.Require().NoError(err)
	s.Require().Len(out, 1)

	returnedAddress, ok := out[0].(common.Address)
	s.Require().True(ok)
	s.Require().Equal(s.minter, returnedAddress)
}

func (s *PrecompileTestSuite) TestMintByMinter() {
	stateDB := statedb.New(s.ctx, s.app.EvmKeeper, statedb.TxConfig{})
	evm := &vm.EVM{
		StateDB: stateDB,
	}

	// First set minter
	setMethod := s.mezoPrecompile.Abi.Methods["setMinter"]
	methodInputArgs, err := setMethod.Inputs.Pack(s.minter)
	s.Require().NoError(err)

	vmContract := vm.NewContract(&precompile.Contract{}, nil, nil, 0)
	//nolint:gocritic
	vmContract.Input = append(setMethod.ID, methodInputArgs...)
	vmContract.CallerAddress = s.poaOwner

	_, err = s.mezoPrecompile.Run(evm, vmContract, false)
	s.Require().NoError(err)

	// Commit to persist minter setting
	err = stateDB.Commit()
	s.Require().NoError(err)

	// Check initial balance
	initialBalance := s.app.BankKeeper.GetBalance(s.ctx, s.recipientSDK, Denom)
	s.Require().Equal(sdkmath.NewInt(0), initialBalance.Amount)

	// Now mint tokens (recreate stateDB and EVM for fresh state)
	stateDB = statedb.New(s.ctx, s.app.EvmKeeper, statedb.TxConfig{})
	evm = &vm.EVM{
		StateDB: stateDB,
	}

	mintMethod := s.mezoPrecompile.Abi.Methods["mint"]
	mintAmount := big.NewInt(1000000000000000000) // 1 token with 18 decimals
	methodInputArgs, err = mintMethod.Inputs.Pack(s.recipient, mintAmount)
	s.Require().NoError(err)

	vmContract = vm.NewContract(&precompile.Contract{}, nil, nil, 0)
	//nolint:gocritic
	vmContract.Input = append(mintMethod.ID, methodInputArgs...)
	vmContract.CallerAddress = s.minter

	_, err = s.mezoPrecompile.Run(evm, vmContract, false)
	s.Require().NoError(err)

	// Commit to persist minting
	err = stateDB.Commit()
	s.Require().NoError(err)

	// Verify balance increased
	newBalance := s.app.BankKeeper.GetBalance(s.ctx, s.recipientSDK, Denom)
	s.Require().Equal(sdkmath.NewIntFromBigInt(mintAmount), newBalance.Amount)
}

func (s *PrecompileTestSuite) TestMintByNonMinter() {
	// First set minter
	setMethod := s.mezoPrecompile.Abi.Methods["setMinter"]
	methodInputArgs, err := setMethod.Inputs.Pack(s.minter)
	s.Require().NoError(err)

	vmContract := vm.NewContract(&precompile.Contract{}, nil, nil, 0)
	//nolint:gocritic
	vmContract.Input = append(setMethod.ID, methodInputArgs...)
	vmContract.CallerAddress = s.poaOwner

	evm := &vm.EVM{
		StateDB: statedb.New(s.ctx, statedb.NewMockKeeper(), statedb.TxConfig{}),
	}

	_, err = s.mezoPrecompile.Run(evm, vmContract, false)
	s.Require().NoError(err)

	// Try to mint tokens with unauthorized address
	mintMethod := s.mezoPrecompile.Abi.Methods["mint"]
	mintAmount := big.NewInt(1000000000000000000)
	methodInputArgs, err = mintMethod.Inputs.Pack(s.recipient, mintAmount)
	s.Require().NoError(err)

	vmContract = vm.NewContract(&precompile.Contract{}, nil, nil, 0)
	//nolint:gocritic
	vmContract.Input = append(mintMethod.ID, methodInputArgs...)
	vmContract.CallerAddress = s.unauthorizedAddr

	output, err := s.mezoPrecompile.Run(evm, vmContract, false)
	s.Require().Error(err)
	s.Require().Nil(output)
	s.Require().ErrorContains(err, "sender is not the minter")
}

func (s *PrecompileTestSuite) TestMintWhenMinterNotSet() {
	mintMethod := s.mezoPrecompile.Abi.Methods["mint"]
	mintAmount := big.NewInt(1000000000000000000)
	methodInputArgs, err := mintMethod.Inputs.Pack(s.recipient, mintAmount)
	s.Require().NoError(err)

	vmContract := vm.NewContract(&precompile.Contract{}, nil, nil, 0)
	//nolint:gocritic
	vmContract.Input = append(mintMethod.ID, methodInputArgs...)
	vmContract.CallerAddress = s.minter

	evm := &vm.EVM{
		StateDB: statedb.New(s.ctx, statedb.NewMockKeeper(), statedb.TxConfig{}),
	}

	output, err := s.mezoPrecompile.Run(evm, vmContract, false)
	s.Require().Error(err)
	s.Require().Nil(output)
	s.Require().ErrorContains(err, "minter not set")
}

func (s *PrecompileTestSuite) TestMintToZeroAddress() {
	// First set minter
	setMethod := s.mezoPrecompile.Abi.Methods["setMinter"]
	methodInputArgs, err := setMethod.Inputs.Pack(s.minter)
	s.Require().NoError(err)

	vmContract := vm.NewContract(&precompile.Contract{}, nil, nil, 0)
	//nolint:gocritic
	vmContract.Input = append(setMethod.ID, methodInputArgs...)
	vmContract.CallerAddress = s.poaOwner

	evm := &vm.EVM{
		StateDB: statedb.New(s.ctx, statedb.NewMockKeeper(), statedb.TxConfig{}),
	}

	_, err = s.mezoPrecompile.Run(evm, vmContract, false)
	s.Require().NoError(err)

	// Try to mint to zero address
	mintMethod := s.mezoPrecompile.Abi.Methods["mint"]
	mintAmount := big.NewInt(1000000000000000000)
	methodInputArgs, err = mintMethod.Inputs.Pack(common.Address{}, mintAmount)
	s.Require().NoError(err)

	vmContract = vm.NewContract(&precompile.Contract{}, nil, nil, 0)
	//nolint:gocritic
	vmContract.Input = append(mintMethod.ID, methodInputArgs...)
	vmContract.CallerAddress = s.minter

	output, err := s.mezoPrecompile.Run(evm, vmContract, false)
	s.Require().Error(err)
	s.Require().Nil(output)
	s.Require().ErrorContains(err, "cannot mint to zero address")
}

func (s *PrecompileTestSuite) TestMintZeroAmount() {
	// First set minter
	setMethod := s.mezoPrecompile.Abi.Methods["setMinter"]
	methodInputArgs, err := setMethod.Inputs.Pack(s.minter)
	s.Require().NoError(err)

	vmContract := vm.NewContract(&precompile.Contract{}, nil, nil, 0)
	//nolint:gocritic
	vmContract.Input = append(setMethod.ID, methodInputArgs...)
	vmContract.CallerAddress = s.poaOwner

	evm := &vm.EVM{
		StateDB: statedb.New(s.ctx, statedb.NewMockKeeper(), statedb.TxConfig{}),
	}

	_, err = s.mezoPrecompile.Run(evm, vmContract, false)
	s.Require().NoError(err)

	// Try to mint zero amount
	mintMethod := s.mezoPrecompile.Abi.Methods["mint"]
	mintAmount := big.NewInt(0)
	methodInputArgs, err = mintMethod.Inputs.Pack(s.recipient, mintAmount)
	s.Require().NoError(err)

	vmContract = vm.NewContract(&precompile.Contract{}, nil, nil, 0)
	//nolint:gocritic
	vmContract.Input = append(mintMethod.ID, methodInputArgs...)
	vmContract.CallerAddress = s.minter

	output, err := s.mezoPrecompile.Run(evm, vmContract, false)
	s.Require().Error(err)
	s.Require().Nil(output)
	s.Require().ErrorContains(err, "amount must be positive")
}

func (s *PrecompileTestSuite) TestSetMinterToZeroAddress() {
	method := s.mezoPrecompile.Abi.Methods["setMinter"]

	methodInputArgs, err := method.Inputs.Pack(common.Address{})
	s.Require().NoError(err)

	stateDB := statedb.New(s.ctx, s.app.EvmKeeper, statedb.TxConfig{})
	vmContract := vm.NewContract(&precompile.Contract{}, nil, nil, 0)
	//nolint:gocritic
	vmContract.Input = append(method.ID, methodInputArgs...)
	vmContract.CallerAddress = s.poaOwner

	evm := &vm.EVM{
		StateDB: stateDB,
	}

	_, err = s.mezoPrecompile.Run(evm, vmContract, false)
	s.Require().NoError(err)

	// Commit to persist the change
	err = stateDB.Commit()
	s.Require().NoError(err)

	// Verify minter was set to zero address
	params := s.app.EvmKeeper.GetParams(s.ctx)
	zeroAddr := common.Address{}
	s.Require().Equal(zeroAddr.Hex(), params.MezoMinterAddress)

	// Now verify getMinter returns zero address
	getMethod := s.mezoPrecompile.Abi.Methods["getMinter"]
	methodInputArgs, err = getMethod.Inputs.Pack()
	s.Require().NoError(err)

	vmContract = vm.NewContract(&precompile.Contract{}, nil, nil, 0)
	//nolint:gocritic
	vmContract.Input = append(getMethod.ID, methodInputArgs...)

	output, err := s.mezoPrecompile.Run(evm, vmContract, false)
	s.Require().NoError(err)

	out, err := getMethod.Outputs.Unpack(output)
	s.Require().NoError(err)
	s.Require().Len(out, 1)

	returnedAddress, ok := out[0].(common.Address)
	s.Require().True(ok)
	s.Require().Equal(zeroAddr, returnedAddress)
}
