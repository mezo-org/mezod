package maintenance

// EvmByteCode is the EVM bytecode of the maintenance precompile. This code is
// returned by eth_getCode and ensures the precompile address is detected as a
// smart contract by external services. note: It should NOT contain a 0x prefix
//
// FIXME: As we already have a live chain at the moment of adding this precompile,
//
//	we cannot add a non-empty bytecode here. This bytecode is used
//	to create an EVM account for the precompile, in the InitChainer.
//	This account is needed to perform precompile verification in block explorers.
//	However, if we put a non-empty bytecode here we would cause troubles
//	for new full nodes trying to sync from genesis. This is because
//	the genesis resolved by new nodes would contain the maintenance
//	precompile bytecode while the actual genesis accepted by the
//	chain would not. Such a difference would cause an app state hash
//	mismatch and would crash the sync for new nodes on block 1.
//	This problem must be addressed as part of on-chain upgrades.
//	See: https://linear.app/thesis-co/issue/ENG-391/figure-out-how-to-handle-updates-of-byte-code-for-precompiles
//
//nolint:all
const EvmByteCode = ""
