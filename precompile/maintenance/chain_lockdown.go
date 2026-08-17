package maintenance

import (
	"fmt"
	"regexp"
	"strconv"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/mezo-org/mezod/precompile"
	"github.com/mezo-org/mezod/x/evm/statedb"
)

// SetChainLockdownMethodName is the name of the setChainLockdown method. It
// matches the name of the method in the contract ABI.
const SetChainLockdownMethodName = "setChainLockdown"

// ChainLockdownSetEventName is the name of the ChainLockdownSet event. It
// matches the name of the event in the contract ABI.
const ChainLockdownSetEventName = "ChainLockdownSet"

// setChainLockdownMethod halts the chain. It schedules an upgrade plan whose
// name has no registered handler, at the next block height. The x/upgrade
// PreBlocker then panics on every validator. The chain cannot recover on chain.
type setChainLockdownMethod struct {
	poaKeeper     PoaKeeper
	upgradeKeeper UpgradeKeeper
}

func newSetChainLockdownMethod(
	poaKeeper PoaKeeper,
	upgradeKeeper UpgradeKeeper,
) *setChainLockdownMethod {
	return &setChainLockdownMethod{
		poaKeeper:     poaKeeper,
		upgradeKeeper: upgradeKeeper,
	}
}

func (m *setChainLockdownMethod) MethodName() string {
	return SetChainLockdownMethodName
}

func (m *setChainLockdownMethod) MethodType() precompile.MethodType {
	return precompile.Write
}

func (m *setChainLockdownMethod) RequiredGas(_ []byte) (uint64, bool) {
	return 0, false
}

func (m *setChainLockdownMethod) Payable() bool {
	return false
}

func (m *setChainLockdownMethod) Run(
	context *precompile.RunContext,
	inputs precompile.MethodInputs,
) (precompile.MethodOutputs, []statedb.StateChange, error) {
	if err := precompile.ValidateMethodInputsCount(inputs, 0); err != nil {
		return nil, nil, err
	}

	// This method is restricted to the PoA owner and the emergency team.
	err := m.poaKeeper.CheckOwnerOrEmergencyTeam(
		context.SdkCtx(),
		precompile.TypesConverter.Address.ToSDK(context.MsgSender()),
	)
	if err != nil {
		return nil, nil, err
	}

	lastCompleted, _, err := m.upgradeKeeper.GetLastCompletedUpgrade(context.SdkCtx())
	if err != nil {
		return nil, nil, err
	}

	haltName, err := nextUpgradeName(lastCompleted, m.upgradeKeeper.HasHandler)
	if err != nil {
		return nil, nil, err
	}

	// No handler is registered for the halt name. The x/upgrade PreBlocker
	// panics at this height and every validator stops.
	plan := upgradetypes.Plan{
		Name:   haltName,
		Height: context.SdkCtx().BlockHeight() + 1,
	}

	err = m.upgradeKeeper.ScheduleUpgrade(context.SdkCtx(), plan)
	if err != nil {
		return nil, nil, err
	}

	err = context.EventEmitter().Emit(
		NewChainLockdownSetEvent(haltName),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to emit ChainLockdownSet event: [%w]", err)
	}

	return precompile.MethodOutputs{true}, nil, nil
}

// upgradeNamePattern matches an upgrade name that carries a semantic version.
// Every historical mezod upgrade name matches it.
var upgradeNamePattern = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)$`)

// maxHaltNameCandidates caps the search for a name the binary cannot execute.
// The binary registers a handler for every historical upgrade, so the search
// skips a few names on a chain that never applied them. It never runs to the
// cap unless the binary registers a hundred handlers ahead of the chain.
const maxHaltNameCandidates = 100

// nextUpgradeName derives the halt name from the name of the last completed
// upgrade. The halt name bumps the major version and resets the rest. An empty
// last completed name derives from v0.0.0. A name that does not parse is an
// error.
//
// The running binary must not be able to execute the halt name, so the
// derivation skips every candidate that hasHandler reports as registered. On a
// release binary no future handler exists and the halt name is the next
// version. The skip loop has a cap and reaching the cap is an error.
func nextUpgradeName(
	lastCompleted string,
	hasHandler func(string) bool,
) (string, error) {
	derivationErr := fmt.Errorf(
		"cannot derive the halt name from the last completed upgrade %q",
		lastCompleted,
	)

	// A fresh chain has no completed upgrade.
	current := lastCompleted
	if current == "" {
		current = "v0.0.0"
	}

	matches := upgradeNamePattern.FindStringSubmatch(current)
	if matches == nil {
		return "", derivationErr
	}

	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return "", derivationErr
	}

	// A plan whose name has a registered handler upgrades the chain instead of
	// halting it. Skip every such name.
	firstCandidate := major + 1
	lastCandidate := major + maxHaltNameCandidates

	for candidate := firstCandidate; candidate <= lastCandidate; candidate++ {
		name := fmt.Sprintf("v%d.0.0", candidate)
		if !hasHandler(name) {
			return name, nil
		}
	}

	return "", fmt.Errorf(
		"the binary has an upgrade handler for every name from v%d.0.0 to v%d.0.0",
		firstCandidate,
		lastCandidate,
	)
}

// ChainLockdownSetEvent is emitted when the chain lockdown is set. It carries
// the name of the scheduled upgrade plan that halts the chain.
type ChainLockdownSetEvent struct {
	name string
}

func NewChainLockdownSetEvent(name string) *ChainLockdownSetEvent {
	return &ChainLockdownSetEvent{
		name: name,
	}
}

func (e *ChainLockdownSetEvent) EventName() string {
	return ChainLockdownSetEventName
}

func (e *ChainLockdownSetEvent) Arguments() []*precompile.EventArgument {
	return []*precompile.EventArgument{
		{Indexed: false, Value: e.name},
	}
}
