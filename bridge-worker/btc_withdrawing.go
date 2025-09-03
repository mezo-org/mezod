package bridgeworker

import (
	"context"
	"encoding/hex"
)

func (bw *BridgeWorker) handleBitcoinWithdrawing(ctx context.Context) error {
	bw.logger.Info("Inside Bitcoin withdrawal logic")
	activeWalletPublicKeyHash, err := bw.tbtcBridgeContract.ActiveWalletPubKeyHash()
	if err != nil {
		panic(err)
	}
	bw.logger.Info(
		"tBTC bridge connectivity check",
		"active_wallet_public_key_hash", hex.EncodeToString(activeWalletPublicKeyHash[:]),
	)
	// TODO: Continue with implementation.

	<-ctx.Done()
	return ctx.Err()
}
