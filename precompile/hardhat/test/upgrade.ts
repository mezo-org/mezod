import { expect } from "chai";
import hre from "hardhat";
import deployed from "../ignition/deployments/chain-31611/deployed_addresses.json"

describe("Upgrade", function () {
    // get deployed contract address from ignition
    const contractAddress = deployed["Upgrade#UpgradeCaller"];
   
    async function getContract() {
       const upgrade = await hre.ethers.getContractAt("UpgradeCaller", contractAddress);
       const [ owner, validator1, validator2, validator3 ] = await hre.ethers.getSigners();
       return { upgrade, owner, validator1, validator2, validator3 }
    }
  
    describe("out of gas", function () {
        describe("submitPlan", function () {
            it("Should be rejected", async function () {
                const { upgrade } = await getContract();
                await expect(upgrade.submitPlan("v2.0.0", 20000, "", {gasLimit: 10})).to.be.rejectedWith("out of gas");
            });
        });
        describe("cancelPlan", function () {
            it("Should be rejected", async function () {
                const { upgrade } = await getContract();
                await expect(upgrade.cancelPlan({gasLimit: 10})).to.be.rejectedWith("out of gas");
            });
        });
    });
    describe("not owner", function () {
        describe("submitPlan", function () {
            it("Should be reverted", async function () {
                const { upgrade, validator1 } = await getContract();
                await expect(upgrade.connect(validator1).submitPlan("v2.0.0", 20000, "")).to.be.rejectedWith("execution reverted");
            });
        });
        describe("cancelPlan", function () {
            it("Should be reverted", async function () {
                const { upgrade, validator1 } = await getContract();
                await expect(upgrade.connect(validator1).cancelPlan()).to.be.rejectedWith("execution reverted");
            });
        });
    });
});
