import { task, HardhatUserConfig } from "hardhat/config";
import { ethers } from "ethers";
import "@nomicfoundation/hardhat-toolbox";
import abi from "../abi.json";

import seed0 from "../../../build/node0/mezod/key_seed.json";
import seed1 from "../../../build/node1/mezod/key_seed.json";
import seed2 from "../../../build/node2/mezod/key_seed.json";
import seed3 from "../../../build/node3/mezod/key_seed.json";

const pk0 = ethers.Wallet.fromPhrase(seed0.secret).privateKey;
const pk1 = ethers.Wallet.fromPhrase(seed1.secret).privateKey;
const pk2 = ethers.Wallet.fromPhrase(seed2.secret).privateKey;
const pk3 = ethers.Wallet.fromPhrase(seed3.secret).privateKey;


const precompileAddress = "0x7b7c000000000000000000000000000000000011";

const config: HardhatUserConfig = {
  solidity: "0.8.24",
  defaultNetwork: 'localhost',
  networks: {
    localhost: {
      url: "http://localhost:8545",
      chainId: 31611,
      accounts: [pk0, pk1, pk2, pk3],
    },
  }
};

// task("accounts", "Returns node addresses", async (taskArguments, hre, runSuper) => {
//   const signers = await hre.ethers.getSigners();
//   // const [ node0 ] = await hre.ethers.getSigners()
//   let addresses = []
//   for (let s of signers) {
//     addresses.push(s.address);
//   }
//   console.log(addresses);
// });

task("owner", "Returns the current contract owner", async (taskArguments, hre, runSuper) => {
  const validatorPool = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
  let owner = await validatorPool.owner();
  if (owner) {
    console.log(owner);
    // 0xD9d322CA07ee2ABB0673c24F72c5eF1bD3A8733E
  }
});

task("candidateOwner", "Returns the current contract candidateOwner", async (taskArguments, hre, runSuper) => {
  const validatorPool = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
  let candidateOwner = await validatorPool.candidateOwner();
  if (candidateOwner) {
    console.log(candidateOwner);
    // 0x0000000000000000000000000000000000000000
  }
});

task("validators", "Returns an array of operator addresses for current validators", async (taskArguments, hre, runSuper) => {
  const validatorPool = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider);
  let validators = await validatorPool.validators();
  if (validators) {
    console.log(validators);
    // Result(4) [
    //   '0x61CB1489ea3F68Bbb1AEd3dFddcF4Dc196Ac6e9f',
    //   '0x64a49b408D1BE5BF5ECD3bc4C86136a2963e6359',
    //   '0xD9d322CA07ee2ABB0673c24F72c5eF1bD3A8733E',
    //   '0xf41dE72f3BA59a2d67c64793C1b817eDDf78618B'
    // ]
  }
});

task("validator", "Returns a validator's consensus public key & description")
  .addParam("operator", "The validator's operator address")
  .setAction(async (taskArguments, hre, runSuper) => {
    const validatorPool = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider);
    let validator = await validatorPool.validator(taskArguments.operator);
    if (validator) {
      console.log(validator);
      // Result(2) [
      //   '0xb9eb37a3e925696685af24bf7ebb48096daa683f01db2fce3822ee49cbb25618',
      //   Result(5) [ 'node1', '', '', '', '' ]
      // ]
    }
  });

task("application", "Returns an application's consensus public key & description")
  .addParam("operator", "The application's operator address")
  .setAction(async (taskArguments, hre, runSuper) => {
    const validatorPool = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider);
    let application = await validatorPool.application(taskArguments.operator);
    if (application) {
      console.log(application);
      // method errored out: [application does not exist]
    }
  });

task("applications", "Returns an array of operator addresses for current applications", async (taskArguments, hre, runSuper) => {
  const validatorPool = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider);
  let candidates = await validatorPool.applications();
  if (candidates) {
    console.log(candidates);
    // Result(0) []
  }
});
export default config;
