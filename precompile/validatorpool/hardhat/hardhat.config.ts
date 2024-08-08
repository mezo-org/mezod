import { task, vars, HardhatUserConfig } from "hardhat/config";
import "@nomicfoundation/hardhat-toolbox";
import abi from "../abi.json";

const precompileAddress = "0x7b7c000000000000000000000000000000000011";

const getPrivKeys = (varname: string) : string[] => {
  const strings: string[] = vars.get(varname, "").split(",");
  const keys: string[] = []
  for (let i = 0; i < keys.length; i++) {
    if (strings[i] !== "") {
      keys.push(strings[i]);
    } 
  }
  return keys;
}

const config: HardhatUserConfig = {
  solidity: "0.8.24",
  defaultNetwork: 'localhost',
  networks: {
    localhost: {
      url: "http://localhost:8545",
      chainId: 31611,
      accounts: getPrivKeys("MEZO_LOCALHOST_PRIVKEYS")
    },
    mezo_testnet: {
      url: "http://mezo-node-0.test.mezo.org:8545",
      chainId: 31611,
      accounts: getPrivKeys("MEZO_TESTNET_PRIVKEYS")
    }
  }
};

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

task("leave", "Removes the signers validator from the pool")
  .addParam("signer", "The signer address (msg.sender)")
  .setAction(async (taskArguments, hre, runSuper) => {
    const signer = await hre.ethers.getSigner(taskArguments.signer);
    if (signer) {
      const validatorPool = new hre.ethers.Contract(precompileAddress, abi, signer);
      const result = await validatorPool.leave({gasLimit: 5000000})
      if (result) {
        console.log(result);
      }
    } else {
      console.log("Unknown signer")
    }
  });

task("kick", "Kicks a validator from the pool")
.addParam("signer", "The signer address (msg.sender)")
.addParam("operator", "The operator address of the validator to be kicked")
.setAction(async (taskArguments, hre, runSuper) => {
  const signer = await hre.ethers.getSigner(taskArguments.signer);
  if (signer) {
    const validatorPool = new hre.ethers.Contract(precompileAddress, abi, signer);
    const result = await validatorPool.kick(taskArguments.operator)
    if (result) {
      console.log(result);
    }
  } else {
    console.log("Unknown signer")
  }
});
export default config;
