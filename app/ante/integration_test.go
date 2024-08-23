package ante_test

import (
	"time"

	sdkmath "cosmossdk.io/math"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	testutiltx "github.com/mezo-org/mezod/testutil/tx"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/mezo-org/mezod/crypto/ethsecp256k1"
	"github.com/mezo-org/mezod/testutil"
	"github.com/mezo-org/mezod/utils"
)

var _ = Describe("when sending a Cosmos transaction", func() {
	var (
		addr sdk.AccAddress
		priv *ethsecp256k1.PrivKey
		msg  sdk.Msg
	)

	Context("and the sender account has enough balance to pay for the transaction cost", Ordered, func() {
		balance := sdkmath.NewInt(1e18)

		BeforeEach(func() {
			addr, priv = testutiltx.NewAccAddressAndKey()

			msg = &banktypes.MsgSend{
				FromAddress: addr.String(),
				ToAddress:   "mezo1dx67l23hz9l0k9hcher8xz04uj7wf3yulxfqd2",
				Amount:      sdk.Coins{sdk.Coin{Amount: sdkmath.NewInt(1e14), Denom: utils.BaseDenom}},
			}

			err := testutil.PrepareAccount(s.ctx, s.app.AccountKeeper, s.app.BankKeeper, addr, balance)
			Expect(err).To(BeNil())

			s.ctx, err = testutil.Commit(s.ctx, s.app, time.Second*0, nil)
			Expect(err).To(BeNil())
		})

		It("should succeed", func() {
			res, err := testutil.DeliverTx(s.ctx, s.app, priv, nil, msg)
			Expect(err).To(BeNil())
			Expect(res.IsOK()).To(BeTrue())
		})
	})

	Context("and the sender account has not enough balance to pay for the transaction cost", func() {
		balance := sdkmath.NewInt(0)

		BeforeEach(func() {
			addr, priv = testutiltx.NewAccAddressAndKey()

			msg = &banktypes.MsgSend{
				FromAddress: addr.String(),
				ToAddress:   "mezo1dx67l23hz9l0k9hcher8xz04uj7wf3yulxfqd2",
				Amount:      sdk.Coins{sdk.Coin{Amount: sdkmath.NewInt(1e14), Denom: utils.BaseDenom}},
			}

			err := testutil.PrepareAccount(s.ctx, s.app.AccountKeeper, s.app.BankKeeper, addr, balance)
			Expect(err).To(BeNil())

			s.ctx, err = testutil.Commit(s.ctx, s.app, time.Second*0, nil)
			Expect(err).To(BeNil())
		})

		It("should fail", func() {
			res, err := testutil.DeliverTx(s.ctx, s.app, priv, nil, msg)
			Expect(err).To(BeNil())
			Expect(res.IsOK()).To(BeFalse())
		})
	})
})
