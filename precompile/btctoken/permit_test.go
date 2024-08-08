package btctoken_test

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/core/vm"

	"github.com/mezo-org/mezod/precompile"
	"github.com/mezo-org/mezod/precompile/btctoken"
	"github.com/mezo-org/mezod/x/evm/statedb"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	PermitTypehash = "Permit(address owner,address spender,uint256 value,uint256 nonce,uint256 deadline)"
	amount         = int64(100)
)

func (s *PrecompileTestSuite) TestPermit() {
	testcases := []struct {
		name        string
		run         func(nonce int64) []interface{}
		postCheck   func()
		basicPass   bool
		runs        int
		errContains string
	}{
		{
			name:        "empty args",
			run:         func(_ int64) []interface{} { return nil },
			runs:        1,
			errContains: "argument count mismatch",
		},
		{
			name: "argument count mismatch",
			run: func(_ int64) []interface{} {
				return []interface{}{
					1, 2, 3,
				}
			},
			runs:        1,
			errContains: "argument count mismatch",
		},
		{
			name: "invalid owner address",
			run: func(_ int64) []interface{} {
				return []interface{}{
					"invalid address", s.account1.EvmAddr, big.NewInt(1), big.NewInt(2), uint8(1), [32]byte{}, [32]byte{},
				}
			},
			runs:        1,
			errContains: "cannot use string as type array as argument",
		},
		{
			name: "invalid spender address",
			run: func(_ int64) []interface{} {
				return []interface{}{
					s.account1.EvmAddr, "invalid address", big.NewInt(1), big.NewInt(2), uint8(1), [32]byte{}, [32]byte{},
				}
			},
			runs:        1,
			errContains: "cannot use string as type array as argument",
		},
		{
			name: "invalid amount",
			run: func(_ int64) []interface{} {
				return []interface{}{
					s.account1.EvmAddr, s.account2.EvmAddr, "invalid amount", big.NewInt(2), uint8(1), [32]byte{}, [32]byte{},
				}
			},
			runs:        1,
			errContains: "cannot use string as type ptr as argument",
		},
		{
			name: "invalid permit typehash",
			run: func(nonce int64) []interface{} {
				deadline := time.Now().Add(24 * time.Hour).Unix() // tmr

				invalidPermitTypehash := "InvalidPermit(address owner,address spender,uint256 value,uint256 nonce,uint256 deadline)"
				digest := buildDigest(s, invalidPermitTypehash, s.account1.EvmAddr, s.account2.EvmAddr, amount, nonce, deadline)

				// Sign the hash with the private key
				// Extract r, s, v values from the signature
				// rComponent := new(big.Int).SetBytes(signature[:32])
				rComponent, sComponent, v := sign(digest, s) // Ethereum specific adjustment

				return []interface{}{
					s.account1.EvmAddr, s.account2.EvmAddr, big.NewInt(amount), big.NewInt(deadline), v, rComponent, sComponent,
				}
			},
			runs:        1,
			basicPass:   true,
			errContains: "verification failed over the signed message",
		},
		{
			name: "invalid owner in the hashed digest",
			run: func(nonce int64) []interface{} {
				deadline := time.Now().Add(24 * time.Hour).Unix() // tmr

				digest := buildDigest(s, PermitTypehash, s.account2.EvmAddr, s.account2.EvmAddr, amount, nonce, deadline)

				// Sign the hash with the private key
				// Extract r, s, v values from the signature
				// rComponent := new(big.Int).SetBytes(signature[:32])
				rComponent, sComponent, v := sign(digest, s) // Ethereum specific adjustment

				return []interface{}{
					s.account1.EvmAddr, s.account2.EvmAddr, big.NewInt(amount), big.NewInt(deadline), v, rComponent, sComponent,
				}
			},
			runs:        1,
			basicPass:   true,
			errContains: "verification failed over the signed message",
		},
		{
			name: "invalid spender in the hashed digest",
			run: func(nonce int64) []interface{} {
				deadline := time.Now().Add(24 * time.Hour).Unix() // tmr

				digest := buildDigest(s, PermitTypehash, s.account1.EvmAddr, s.account1.EvmAddr, amount, nonce, deadline)

				// Sign the hash with the private key
				// Extract r, s, v values from the signature
				// rComponent := new(big.Int).SetBytes(signature[:32])
				rComponent, sComponent, v := sign(digest, s) // Ethereum specific adjustment

				return []interface{}{
					s.account1.EvmAddr, s.account2.EvmAddr, big.NewInt(amount), big.NewInt(deadline), v, rComponent, sComponent,
				}
			},
			runs:        1,
			basicPass:   true,
			errContains: "verification failed over the signed message",
		},
		{
			name: "invalid amount in the hashed digest",
			run: func(nonce int64) []interface{} {
				deadline := time.Now().Add(24 * time.Hour).Unix() // tmr

				digest := buildDigest(s, PermitTypehash, s.account1.EvmAddr, s.account2.EvmAddr, int64(99), nonce, deadline)

				// Sign the hash with the private key
				// Extract r, s, v values from the signature
				// rComponent := new(big.Int).SetBytes(signature[:32])
				rComponent, sComponent, v := sign(digest, s) // Ethereum specific adjustment

				return []interface{}{
					s.account1.EvmAddr, s.account2.EvmAddr, big.NewInt(amount), big.NewInt(deadline), v, rComponent, sComponent,
				}
			},
			runs:        1,
			basicPass:   true,
			errContains: "verification failed over the signed message",
		},
		{
			name: "invalid nonce in the hashed digest",
			run: func(_ int64) []interface{} {
				deadline := time.Now().Add(24 * time.Hour).Unix() // tmr
				invalidNonce := int64(1)

				digest := buildDigest(s, PermitTypehash, s.account1.EvmAddr, s.account2.EvmAddr, amount, invalidNonce, deadline)

				// Sign the hash with the private key
				// Extract r, s, v values from the signature
				// rComponent := new(big.Int).SetBytes(signature[:32])
				rComponent, sComponent, v := sign(digest, s) // Ethereum specific adjustment

				return []interface{}{
					s.account1.EvmAddr, s.account2.EvmAddr, big.NewInt(amount), big.NewInt(deadline), v, rComponent, sComponent,
				}
			},
			runs:        1,
			basicPass:   true,
			errContains: "verification failed over the signed message",
		},
		{
			name: "invalid deadline in the hashed digest",
			run: func(nonce int64) []interface{} {
				deadline := time.Now().Add(25 * time.Hour).Unix() // tmr

				digest := buildDigest(s, PermitTypehash, s.account1.EvmAddr, s.account2.EvmAddr, amount, nonce, deadline)

				// Sign the hash with the private key
				// Extract r, s, v values from the signature
				// rComponent := new(big.Int).SetBytes(signature[:32])
				rComponent, sComponent, v := sign(digest, s) // Ethereum specific adjustment

				return []interface{}{
					s.account1.EvmAddr, s.account2.EvmAddr, big.NewInt(amount), big.NewInt(deadline), v, rComponent, sComponent,
				}
			},
			runs:        1,
			basicPass:   true,
			errContains: "verification failed over the signed message",
		},
		{
			name: "permit expired",
			run: func(nonce int64) []interface{} {
				expiredDeadline := time.Now().Add(-24 * time.Hour).Unix() // yesterday

				digest := buildDigest(s, PermitTypehash, s.account1.EvmAddr, s.account2.EvmAddr, amount, nonce, expiredDeadline)

				// Sign the hash with the private key
				// Extract r, s, v values from the signature
				// rComponent := new(big.Int).SetBytes(signature[:32])
				rComponent, sComponent, v := sign(digest, s) // Ethereum specific adjustment

				return []interface{}{
					s.account1.EvmAddr, s.account2.EvmAddr, big.NewInt(amount), big.NewInt(expiredDeadline), v, rComponent, sComponent,
				}
			},
			runs:        1,
			basicPass:   true,
			errContains: "permit expired",
		},
		{
			name: "successful permit",
			run: func(nonce int64) []interface{} {
				deadline := time.Now().Add(24 * time.Hour).Unix() // tmr

				digest := buildDigest(s, PermitTypehash, s.account1.EvmAddr, s.account2.EvmAddr, amount, nonce, deadline)

				// Sign the hash with the private key
				// Extract r, s, v values from the signature
				// rComponent := new(big.Int).SetBytes(signature[:32])
				rComponent, sComponent, v := sign(digest, s) // Ethereum specific adjustment

				return []interface{}{
					s.account1.EvmAddr, s.account2.EvmAddr, big.NewInt(amount), big.NewInt(deadline), v, rComponent, sComponent,
				}
			},
			runs:      1,
			basicPass: true,
			postCheck: func() {
				s.requireSendAuthz(
					s.account2.SdkAddr,
					s.account1.SdkAddr,
					sdk.NewCoins(sdk.NewInt64Coin("abtc", amount)),
				)
			},
		},
		{
			// This test is to check if the permit function can be executed twice with
			// different nonces. The second execution should also pass since the nonce
			// should be incremented by one in EVM storage for the given owner and in
			// this test as well.
			name: "successful permit executed twice",
			run: func(nonce int64) []interface{} {
				deadline := time.Now().Add(24 * time.Hour).Unix() // tmr

				digest := buildDigest(s, PermitTypehash, s.account1.EvmAddr, s.account2.EvmAddr, amount, nonce, deadline)

				// Sign the hash with the private key
				// Extract r, s, v values from the signature
				// rComponent := new(big.Int).SetBytes(signature[:32])
				rComponent, sComponent, v := sign(digest, s) // Ethereum specific adjustment

				return []interface{}{
					s.account1.EvmAddr, s.account2.EvmAddr, big.NewInt(amount), big.NewInt(deadline), v, rComponent, sComponent,
				}
			},
			runs:      2,
			basicPass: true,
			postCheck: func() {
				s.requireSendAuthz(
					s.account2.SdkAddr,
					s.account1.SdkAddr,
					sdk.NewCoins(sdk.NewInt64Coin("abtc", amount)),
				)
			},
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.SetupTest()
			evm := &vm.EVM{
				StateDB: statedb.New(s.ctx, nil, statedb.TxConfig{}),
			}

			bankKeeper := s.app.BankKeeper
			authzKeeper := s.app.AuthzKeeper
			evmKeeper := *s.app.EvmKeeper

			btcTokenPrecompile, err := btctoken.NewPrecompile(bankKeeper, authzKeeper, evmKeeper)
			s.Require().NoError(err)
			s.btcTokenPrecompile = btcTokenPrecompile

			// Default nonce for the tests starts with zero. Then we can increment it
			// by one to check the permit function with different nonces.
			nonce := int64(0)
			for i := 0; i < tc.runs; i++ {
				var methodInputs []interface{}
				if tc.run != nil {
					methodInputs = tc.run(nonce)
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
				if err != nil && tc.errContains != "" {
					s.Require().ErrorContains(err, tc.errContains, "expected different error message")
					return
				}
				s.Require().NoError(err, "expected no error")

				out, err := method.Outputs.Unpack(output)
				s.Require().NoError(err)
				s.Require().Equal(true, out[0], "expected different value")

				if tc.postCheck != nil {
					tc.postCheck()
				}

				nonce++
			}
		})
	}
}

func sign(digest common.Hash, s *PrecompileTestSuite) ([32]byte, [32]byte, uint8) {
	signature, err := crypto.Sign(digest.Bytes(), s.account1.Priv)
	if err != nil {
		s.Require().NoError(err)
	}

	var rComponent [32]byte
	var sComponent [32]byte

	copy(rComponent[:], signature[:32])
	copy(sComponent[:], signature[32:64])
	v := signature[64] + 27
	return rComponent, sComponent, v
}

func buildDigest(s *PrecompileTestSuite, permitTypehash string, owner, spender common.Address, amount, nonce, deadline int64) common.Hash {
	var PermitTypehashBytes32 [32]byte
	copy(PermitTypehashBytes32[:], crypto.Keccak256([]byte(permitTypehash))[:32])

	bytes32Type, _ := abi.NewType("bytes32", "", nil)
	addressType, _ := abi.NewType("address", "", nil)
	uint256Type, _ := abi.NewType("uint256", "", nil)

	message, err := abi.Arguments{
		{Type: bytes32Type},
		{Type: addressType},
		{Type: addressType},
		{Type: uint256Type},
		{Type: uint256Type},
		{Type: uint256Type},
	}.Pack(
		PermitTypehashBytes32,
		owner,
		spender,
		big.NewInt(amount),
		big.NewInt(nonce),
		big.NewInt(deadline),
	)
	s.Require().NoError(err)

	hashedMessage := crypto.Keccak256Hash(message)

	var DomainSeparatorBytes32 [32]byte
	domainSeparator, err := buildDomainSeparator()
	if err != nil {
		s.Require().NoError(err)
	}

	copy(DomainSeparatorBytes32[:], domainSeparator[:32])

	encodedData := append([]byte("\x19\x01"), DomainSeparatorBytes32[:]...)
	encodedData = append(encodedData, hashedMessage.Bytes()...)

	return crypto.Keccak256Hash(encodedData)
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

	return crypto.Keccak256(encodedDomainSeparator), nil
}
