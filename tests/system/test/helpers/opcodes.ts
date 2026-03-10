import hre from "hardhat"

/**
 * Get deployed opcode listing for a compiled contract from Hardhat build info.
 * @param contractName Name of the compiled contract.
 * @returns Deployed opcodes emitted by solc.
 */
export async function getDeployedOpcodes(
  contractName: string,
): Promise<string[]> {
  const artifact = await hre.artifacts.readArtifact(contractName)
  const fullyQualifiedName = `${artifact.sourceName}:${artifact.contractName}`
  const buildInfo = await hre.artifacts.getBuildInfo(fullyQualifiedName)

  if (buildInfo === undefined) {
    throw new Error(`build info not found for ${fullyQualifiedName}`)
  }

  const contractOutput =
    buildInfo.output.contracts?.[artifact.sourceName]?.[artifact.contractName]
  const opcodes = contractOutput?.evm?.deployedBytecode?.opcodes

  if (opcodes === undefined || opcodes.length === 0) {
    throw new Error(`deployed opcodes not found for ${fullyQualifiedName}`)
  }

  return opcodes.split(/\s+/).filter(Boolean)
}
