package bridgeworker

import "context"

func (bw *BridgeWorker) handleBitcoinWithdrawing(ctx context.Context) error {
	bw.logger.Info("Inside Bitcoin withdrawal logic")
	// TODO: Continue with implementation.

	<-ctx.Done()
	return ctx.Err()
}
