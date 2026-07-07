import type { HardhatRuntimeEnvironment } from "hardhat/types";
import type { DeployFunction } from "hardhat-deploy/types";

const func: DeployFunction = async (hre: HardhatRuntimeEnvironment) => {
  const { getNamedAccounts, deployments } = hre;
  const { deployer } = await getNamedAccounts();

  await deployments.deploy("BridgeOutCaller", {
    from: deployer,
    args: [],
    log: true,
    waitConfirmations: 1,
  });
};

export default func;

func.tags = ["IndirectBridgeOut"];
