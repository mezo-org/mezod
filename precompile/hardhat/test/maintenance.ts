import { expect } from "chai";
import hre from "hardhat";
import deployed from "../ignition/deployments/chain-31611/deployed_addresses.json"

describe("Maintenance", function () {
    // get deployed contract address from ignition
    const contractAddress = deployed["Maintenance#MaintenanceCaller"];
   
    async function getContract() {
       const maintenance = await hre.ethers.getContractAt("MaintenanceCaller", contractAddress);
       const [ owner, validator1, validator2, validator3 ] = await hre.ethers.getSigners();
       return { maintenance, owner, validator1, validator2, validator3 }
    }
  
    describe("out of gas", function () {
        describe("setSupportNonEIP155Txs", function () {
            it("Should be rejected", async function () {
                const { maintenance } = await getContract();
                await expect(maintenance.setSupportNonEIP155Txs(true, {gasLimit: 10})).to.be.rejectedWith("out of gas");
            });
        })
        describe("setPrecompileByteCode", function () {
            it("Should be rejected", async function () {
                const { maintenance } = await getContract();
                await expect(maintenance.setPrecompileByteCode(contractAddress, "0x", {gasLimit: 10})).to.be.rejectedWith("out of gas");
            });
        });
    });
    describe("not owner", function () {
        describe("setSupportNonEIP155Txs", function () {
            it("Should be reverted", async function () {
                const { maintenance, validator1 } = await getContract();
                await expect(maintenance.connect(validator1).setSupportNonEIP155Txs(true)).to.be.rejectedWith("execution reverted");
            });
        })
        describe("setPrecompileByteCode", function () {
            it("Should be reverted", async function () {
                const { maintenance, validator1 } = await getContract();
                await expect(maintenance.connect(validator1).setPrecompileByteCode(contractAddress, "0x")).to.be.rejectedWith("execution reverted");
            });
        });
    });
});