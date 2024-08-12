import { task } from "hardhat/config";
import "@nomicfoundation/hardhat-toolbox";

import abi from "../../validatorpool/abi.json";
const precompileAddress = "0x7b7c000000000000000000000000000000000011";

task("validatorPool:owner", "Returns the current contract owner", async (taskArguments, hre, runSuper) => {
    const validatorPool = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider);
    let owner = await validatorPool.owner();
    if (owner) {
      console.log(owner);
      // 0xD9d322CA07ee2ABB0673c24F72c5eF1bD3A8733E
    }
  });
  
task("validatorPool:candidateOwner", "Returns the current contract candidateOwner", async (taskArguments, hre, runSuper) => {
const validatorPool = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider);
let candidateOwner = await validatorPool.candidateOwner();
if (candidateOwner) {
    console.log(candidateOwner);
    // 0x0000000000000000000000000000000000000000
}
});

task("validatorPool:validators", "Returns an array of operator addresses for current validators", async (taskArguments, hre, runSuper) => {
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

task("validatorPool:validator", "Returns a validator's consensus public key & description")
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

task("validatorPool:application", "Returns an application's consensus public key & description")
.addParam("operator", "The application's operator address")
.setAction(async (taskArguments, hre, runSuper) => {
    const validatorPool = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider);
    let application = await validatorPool.application(taskArguments.operator);
    if (application) {
    console.log(application);
    // method errored out: [application does not exist]
    }
});

task("validatorPool:applications", "Returns an array of operator addresses for current applications", async (taskArguments, hre, runSuper) => {
const validatorPool = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider);
let candidates = await validatorPool.applications();
if (candidates) {
    console.log(candidates);
    // Result(0) []
}
});

task("validatorPool:leave", "Removes the signers validator from the pool")
.addParam("signer", "The signer address (msg.sender)")
.setAction(async (taskArguments, hre, runSuper) => {
    const signer = await hre.ethers.getSigner(taskArguments.signer);
    if (signer) {
    const validatorPool = new hre.ethers.Contract(precompileAddress, abi, signer);
    const result = await validatorPool.leave({gasLimit: 46128n});
    if (result) {
        console.log(result);
    }
    } else {
    console.log("Unknown signer");
    }
});

task("validatorPool:kick", "Kicks a validator from the pool")
.addParam("signer", "The signer address (msg.sender)")
.addParam("operator", "The operator address of the validator to be kicked")
.setAction(async (taskArguments, hre, runSuper) => {
const signer = await hre.ethers.getSigner(taskArguments.signer);
if (signer) {
    const validatorPool = new hre.ethers.Contract(precompileAddress, abi, signer);
    const result = await validatorPool.kick(taskArguments.operator);
    if (result) {
    console.log(result);
    }
} else {
    console.log("Unknown signer");
}
});

task("validatorPool:transferOwnership", "Begins the ownership transfer flow (owner)")
.addParam("signer", "The signer address (msg.sender)")
.addParam("to", "Address to transfer ownership to")
.setAction(async (taskArguments, hre, runSuper) => {
const signer = await hre.ethers.getSigner(taskArguments.signer);
if (signer) {
    const validatorPool = new hre.ethers.Contract(precompileAddress, abi, signer);
    const result = await validatorPool.transferOwnership(taskArguments.to);
    if (result) {
    console.log(result);
    }
} else {
    console.log("Unknown signer");
}
});

task("validatorPool:acceptOwnership", "Accepts a pending ownership transfer (candidateOwner)")
.addParam("signer", "The signer address (msg.sender)")
.setAction(async (taskArguments, hre, runSuper) => {
const signer = await hre.ethers.getSigner(taskArguments.signer);
if (signer) {
    const validatorPool = new hre.ethers.Contract(precompileAddress, abi, signer);
    const result = await validatorPool.acceptOwnership({gasLimit: 50000});
    if (result) {
    console.log(result);
    }
} else {
    console.log("Unknown signer");
}
});

task("validatorPool:submitApplication", "Submit a new validator application")
.addParam("signer", "The signer address (msg.sender)")
.addParam("conspubkey", "The validator's consensus pub key")
.addParam("moniker", "The validator's name")
.addOptionalParam("identity", "Optional identity signature (ex. UPort or Keybase)", "")
.addOptionalParam("website", "Optional website link", "")
.addOptionalParam("security", "Optional security contact information", "")
.addOptionalParam("details", "Optional details about the validator", "")
.setAction(async (taskArguments, hre, runSuper) => {
const signer = await hre.ethers.getSigner(taskArguments.signer);
if (signer) {
    const validatorPool = new hre.ethers.Contract(precompileAddress, abi, signer);
    const description: string[] = [
    taskArguments.moniker,
    taskArguments.identity,
    taskArguments.website,
    taskArguments.security,
    taskArguments.details,
    ];

    const result = await validatorPool.submitApplication(taskArguments.conspubkey, description, {gasLimit: 50000});
    if (result) {
    console.log(result);
    }
} else {
    console.log("Unknown signer");
}
});

task("validatorPool:approveApplication", "Approves a pending validator application (owner)")
.addParam("signer", "The signer address (msg.sender)")
.addParam("operator", "The validator's operator address of the application to approve")
.setAction(async (taskArguments, hre, runSuper) => {
const signer = await hre.ethers.getSigner(taskArguments.signer);
if (signer) {
    const validatorPool = new hre.ethers.Contract(precompileAddress, abi, signer);
    const result = await validatorPool.approveApplication(taskArguments.operator, {gasLimit: 50000});
    if (result) {
    console.log(result);
    }
} else {
    console.log("Unknown signer");
}
});