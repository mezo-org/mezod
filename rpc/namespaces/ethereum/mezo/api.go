package mezo

import (
	"cosmossdk.io/log"

	"github.com/mezo-org/mezod/rpc/backend"
	rpctypes "github.com/mezo-org/mezod/rpc/types"
	evmtypes "github.com/mezo-org/mezod/x/evm/types"
)

// PublicAPI is the custom set of methods prefixed with mezo_ in the EVM JSON-RPC API.
type PublicAPI struct {
	logger  log.Logger
	backend backend.EVMBackend
}

func NewPublicAPI(logger log.Logger, backend backend.EVMBackend) *PublicAPI {
	api := &PublicAPI{
		logger:  logger.With("api", "mezo"),
		backend: backend,
	}

	return api
}

// EstimateCost returns the estimated cost of a transaction.
func (e *PublicAPI) EstimateCost(args evmtypes.TransactionArgs, blockNrOptional *rpctypes.BlockNumber) (*rpctypes.EstimateCostResult, error) {
	e.logger.Debug("mezo_estimateCost")
	return e.backend.EstimateCost(args, blockNrOptional)
}
