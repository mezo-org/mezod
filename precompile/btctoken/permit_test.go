package btctoken_test

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/core/vm"

	"math/big"

	"github.com/evmos/evmos/v12/precompile"
	"github.com/evmos/evmos/v12/precompile/btctoken"
	"github.com/evmos/evmos/v12/x/evm/statedb"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func (s *PrecompileTestSuite) TestPermit() {
	amount := int64(100)

	testcases := []struct {
		name        string
		run         func() []interface{}
		postCheck   func()
		basicPass   bool
		errContains string
	}{
		// TODO: add more test cases
		{
			name: "successful permit",
			run: func() []interface{} {
				tmr := time.Now().Add(24 * time.Hour).Unix()

				// Create a message hash
				PermitTypehash := "Permit(address owner,address spender,uint256 value,uint256 nonce,uint256 deadline)"
				var PermitTypehashBytes32 [32]byte
				// Convert to bytes32 (32-byte array)
				copy(PermitTypehashBytes32[:], crypto.Keccak256([]byte(PermitTypehash))[:32])

				permit := crypto.Keccak256Hash(PermitTypehashBytes32[:])
				// TODO: add more params to the message
				message := permit[:]
				hashedMessage := crypto.Keccak256Hash(message)

				// Sign the hash with the private key
				signature, err := crypto.Sign(hashedMessage.Bytes(), s.account1.Priv)
				if err != nil {
					s.Require().NoError(err)
				}

				var r_component [32]byte
    		var s_component [32]byte

				// Extract r, s, v values from the signature
				// r_component := new(big.Int).SetBytes(signature[:32])
				copy(r_component[:], signature[:32])
				copy(s_component[:], signature[32:64])
				v := uint8(signature[64]) + 27 // Ethereum specific adjustment
				
				return []interface{}{
					s.account1.EvmAddr, s.account2.EvmAddr, big.NewInt(amount), big.NewInt(tmr), v, r_component, s_component,
				}
			},
			basicPass: true,
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			evm := &vm.EVM{
				StateDB: statedb.New(s.ctx, nil, statedb.TxConfig{}),
			}

			bankKeeper := s.app.BankKeeper
			authzKeeper := s.app.AuthzKeeper
			evmKeeper := *s.app.EvmKeeper

			btcTokenPrecompile, err := btctoken.NewPrecompile(bankKeeper, authzKeeper, evmKeeper)
			s.Require().NoError(err)
			s.btcTokenPrecompile = btcTokenPrecompile

			var methodInputs []interface{}
			if tc.run != nil {
				methodInputs = tc.run()
			}

			method := s.btcTokenPrecompile.Abi.Methods["permit"]
			var methodInputArgs []byte
			methodInputArgs, err = method.Inputs.Pack(methodInputs...)

			if tc.basicPass {
				s.Require().NoError(err, "expected no error")
			} else {
				s.Require().Error(err, "expected error")
				s.Require().ErrorContains(err, tc.errContains, "expected different error message")
				return
			}

			vmContract := vm.NewContract(&precompile.Contract{}, nil, nil, 0)
			// These first 4 bytes correspond to the method ID (first 4 bytes of the
			// Keccak-256 hash of the function signature).
			// In this case a function signature is 'function permit(address owner, address spender, uint256 amount, uint256 deadline, uint8 v, bytes32 r, bytes32 s)'
			vmContract.Input = append([]byte{0xd5, 0x05, 0xac, 0xcf}, methodInputArgs...)
			vmContract.CallerAddress = s.account2.EvmAddr

			output, err := s.btcTokenPrecompile.Run(evm, vmContract, false)
			s.Require().NoError(err, "expected no error")

			out, err := method.Outputs.Unpack(output)
			s.Require().NoError(err)
			s.Require().Equal(true, out[0], "expected different value")

			if tc.postCheck != nil {
				tc.postCheck()
			}
		})
	}
}

// This functions implements the EIP712 domain separator for the permit function
// and produces the same result as the Solidity code seen e.g. in the OpenZeppelin
// lib https://github.com/OpenZeppelin/openzeppelin-contracts/blob/master/contracts/utils/cryptography/EIP712.sol#L89
// that is used by tBTC token https://github.com/keep-network/tbtc-v2/blob/main/solidity/contracts/token/TBTC.sol#L8
// The result of this function is hardcoded in the production code (permit.go) to
// comply with the EVM implementation.
func buildDomainSeparator() ([]byte, error) {
	// Hash the domain type
	domainType := "EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"
	name := "BTC"
	version := "1"
	chainID := big.NewInt(31612)
	verifyingContract := common.HexToAddress("0x7b7c000000000000000000000000000000000000")

	var DomainTypeHashBytes32 [32]byte
	var NameHashBytes32 [32]byte
	var VersionHashBytes32 [32]byte

	// Convert to bytes32 (32-byte array)
	copy(DomainTypeHashBytes32[:], crypto.Keccak256([]byte(domainType))[:32])
	copy(NameHashBytes32[:], crypto.Keccak256([]byte(name))[:32])
	copy(VersionHashBytes32[:], crypto.Keccak256([]byte(version))[:32])

	bytes32Type, err := abi.NewType("bytes32", "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create new type: %v", err)
	}
	addressType, err := abi.NewType("address", "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create new type: %v", err)
	}
	uint256Type, err := abi.NewType("uint256", "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create new type: %v", err)
	}

	// Encode the permit parameters
	encodedDomainSeparator, err := abi.Arguments{
		{Type: bytes32Type},
		{Type: bytes32Type},
		{Type: bytes32Type},
		{Type: uint256Type},
		{Type: addressType},
	}.Pack(
		DomainTypeHashBytes32,
		NameHashBytes32,
		VersionHashBytes32,
		chainID,
		verifyingContract,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to encode domain separator: %v", err)
	}
	encodedDomainSeparatorHashed := crypto.Keccak256(encodedDomainSeparator)

	return encodedDomainSeparatorHashed, nil
}
