import { task } from "hardhat/config";
import "@nomicfoundation/hardhat-toolbox";

import abi from "../../btctoken/abi.json";
const precompileAddress = "0x7b7c000000000000000000000000000000000001";

task("btctoken:name", "Returns the token name", async (taskArguments, hre, runSuper) => {
    const btctoken = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider);
    let name = await btctoken.name();
    if (name) {
      console.log(name);
      // BTC
    }
});