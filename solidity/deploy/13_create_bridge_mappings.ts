import type { DeployFunction } from "hardhat-deploy/types"
import type { HardhatRuntimeEnvironment } from "hardhat/types"

// ethereum -> mezo mappings
const mappings: Record<string, string> = {
    "SolvBTC": "mSolvBTC"
}

const func: DeployFunction = async (hre: HardhatRuntimeEnvironment) => {
    const { ethers, getNamedAccounts, deployments, helpers, companionNetworks } = hre
    const { poaOwner } = await getNamedAccounts()
    const { log, read, execute } = deployments

    for (const [ethereumToken, mezoToken] of Object.entries(mappings)) {
        const ethereumTokenDeployment = await companionNetworks['ethereum'].deployments.getOrNull(ethereumToken);
        const isValidDeployment = ethereumTokenDeployment && helpers.address.isValid(ethereumTokenDeployment.address)

        if (!isValidDeployment) {
            log(`${ethereumToken} deployment artifact for ethereum not found; skipping...`)
            continue
        }

        const mezoTokenDeployment = await deployments.get(mezoToken)

        log(`creating mapping: ${ethereumToken}/${ethereumTokenDeployment.address} -> ${mezoToken}/${mezoTokenDeployment.address}`)
        
        const existingMapping = (await read("AssetsBridge", "getERC20TokenMapping", ethereumTokenDeployment.address)).mezoToken
        if (existingMapping !== ethers.constants.AddressZero) {
            if (existingMapping === mezoTokenDeployment.address) {
                log("mapping already exists; skipping...")
            } else {
                // Do not automatically overwrite existing mapping.
                // This situation is exceptional and requires manual intervention.
                log(
                    "mapping already exists but for different mezo token; " + 
                    "manual intervention required; skipping..."
                )
            }
            continue
        }
        
        const tx = await execute(
            "AssetsBridge",
            { from: poaOwner, log: true },
            "createERC20TokenMapping",
            ethereumTokenDeployment.address,
            mezoTokenDeployment.address
        )

        const receipt = await tx.wait()
        log(`mapping created; tx: ${receipt.transactionHash}`)
    }
}

export default func

func.tags = ["CreateBridgeMappings"]
func.dependencies = Object.values(mappings)
func.runAtTheEnd = true
