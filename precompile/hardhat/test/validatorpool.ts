import { expect } from "chai";
import hre from "hardhat";
import deployed from "../ignition/deployments/chain-31611/deployed_addresses.json"

describe("ValidatorPool", function () {
    // get deployed contract address from ignition
    const contractAddress = deployed["ValidatorPool#ValidatorPoolCaller"];
   
    async function getContract() {
       const validatorPool = await hre.ethers.getContractAt("ValidatorPoolCaller", contractAddress);
       const [owner, validator1, validator2, validator3, user] = await hre.ethers.getSigners();
       return { validatorPool, owner, validator1, validator2, validator3, user}
    }
  
    describe("out of gas", function () {
        describe("leave", function () {
            it("Should be rejected", async function () {
                const { validatorPool } = await getContract();
                await expect(validatorPool.leave({gasLimit: 10})).to.be.rejectedWith("out of gas");
            });
        });
        describe("kick", function () {
            it("Should be rejected", async function () {
                const { validatorPool, validator1 } = await getContract();
                await expect(validatorPool.kick(validator1, {gasLimit: 10})).to.be.rejectedWith("out of gas");
            });
        });
        describe("approveApplication", function () {
            it("Should be rejected", async function () {
                const { validatorPool, validator1 } = await getContract();
                await expect(validatorPool.approveApplication(validator1, {gasLimit: 10})).to.be.rejectedWith("out of gas");
            });
        });
        describe("submitApplication", function () {
            it("Should be rejected", async function () {
                const { validatorPool } = await getContract();
                const description = {
                    moniker: "moniker",
                    identity: "identity",
                    website: "website",
                    securityContact: "security",
                    details: "details"
                }
                const consPubKey = "0x649e7bfb604b873b5646bf4b11356ca7a43a8b3f28cddc1406cdddcf44b31354"
                await expect(validatorPool.submitApplication(consPubKey, description, {gasLimit: 10})).to.be.rejectedWith("out of gas");
            });
        });
        describe("approveApplication", function () {
            it("Should be rejected", async function () {
                const { validatorPool, validator1 } = await getContract();
                await expect(validatorPool.cleanupApplications({gasLimit: 10})).to.be.rejectedWith("out of gas");
            });
        });
        describe("transferOwnership", function () {
            it("Should be rejected", async function () {
                const { validatorPool, validator1 } = await getContract();
                await expect(validatorPool.transferOwnership(validator1, {gasLimit: 10})).to.be.rejectedWith("out of gas");
            });
        });
        describe("addPrivilege", function () {
            it("Should be rejected", async function () {
                const { validatorPool, validator1 } = await getContract();
                await expect(validatorPool.addPrivilege([validator1.address], 10, {gasLimit: 10})).to.be.rejectedWith("out of gas");
            });
        });
        describe("removePrivilege", function () {
            it("Should be rejected", async function () {
                const { validatorPool, validator1 } = await getContract();
                await expect(validatorPool.removePrivilege([validator1.address], 10, {gasLimit: 10})).to.be.rejectedWith("out of gas");
            });
        });
    });
    describe("not owner", function () {
        describe("kick", function () {
            it("Should be reverted", async function () {
                const { validatorPool, owner, validator1 } = await getContract();
                await expect(validatorPool.connect(validator1).kick(owner)).to.be.rejectedWith("execution reverted");
            });
        });
        describe("approveApplication", function () {
            it("Should be reverted", async function () {
                const { validatorPool, owner, validator1 } = await getContract();
                await expect(validatorPool.connect(validator1).approveApplication(owner)).to.be.rejectedWith("execution reverted");
            });
        });
        describe("cleanupApplications", function () {
            it("Should be reverted", async function () {
                const { validatorPool, owner, validator1 } = await getContract();
                await expect(validatorPool.connect(validator1).cleanupApplications()).to.be.rejectedWith("execution reverted");
            });
        });
        describe("transferOwnership", function () {
            it("Should be reverted", async function () {
                const { validatorPool, owner, validator1 } = await getContract();
                await expect(validatorPool.connect(validator1).transferOwnership(validator1)).to.be.rejectedWith("execution reverted");
            });
        });   
        describe("addPrivilege", function () {
            it("Should be reverted", async function () {
                const { validatorPool, owner, validator1 } = await getContract();
                await expect(validatorPool.connect(validator1).addPrivilege([validator1.address], 10)).to.be.rejectedWith("execution reverted");
            });
        });     
        describe("removePrivilege", function () {
            it("Should be reverted", async function () {
                const { validatorPool, owner, validator1 } = await getContract();
                await expect(validatorPool.connect(validator1).removePrivilege([owner.address], 10)).to.be.rejectedWith("execution reverted");
            });
        });      
    });
});
