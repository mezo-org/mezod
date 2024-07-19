package btctoken_test

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/evmos/v12/precompile/btctoken"
)

func (s *PrecompileTestSuite) TestEmitApprovalEvent() {
	testcases := []struct {
		name    string
		owner   common.Address
		spender common.Address
		amount  *big.Int
	}{
		{
			name:    "pass",
			owner:   s.account1.Addr,
			spender: s.account2.Addr,
			amount:  big.NewInt(100),
		},
	}

	for _, tc := range testcases {
		tc := tc
		s.Run(tc.name, func() {
			ae := btctoken.NewApprovalEvent(tc.owner, tc.spender, tc.amount)
			args := ae.Arguments()

			s.Require().Len(args, 3)

			// Check the first argument
			s.Require().True(args[0].Indexed)
			s.Require().Equal(tc.owner, args[0].Value)

			// Check the second argument
			s.Require().True(args[1].Indexed)
			s.Require().Equal(tc.spender, args[1].Value)

			// Check the third argument
			s.Require().False(args[2].Indexed)
			s.Require().Equal(tc.amount, args[2].Value)
		})
	}
}
