import { expect } from "chai";
import hre from "hardhat";
import deployed from "../ignition/deployments/chain-31611/deployed_addresses.json"

describe("AssetsBridge", function () {
    // get deployed contract address from ignition
    const contractAddress = deployed["AssetsBridge#AssetsBridgeCaller"];
    const sourceToken = "0xffffffffffffffffffffffffffffffffffffffff";
    const mezoToken = "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee";
   
    async function getContract() {
       const assetsbridge = await hre.ethers.getContractAt("AssetsBridgeCaller", contractAddress);
       const [ owner, validator1, validator2, validator3 ] = await hre.ethers.getSigners();
       return { assetsbridge, owner, validator1, validator2, validator3 }
    }
  
    describe("out of gas", function () {
        describe("createERC20TokenMapping", function () {
            it("Should be rejected", async function () {
                const { assetsbridge } = await getContract();
                await expect(assetsbridge.createERC20TokenMapping(sourceToken, mezoToken, {gasLimit: 10})).to.be.rejectedWith("out of gas");
            });
        });
        describe("deleteERC20TokenMapping", function () {
            it("Should be rejected", async function () {
                const { assetsbridge } = await getContract();
                await expect(assetsbridge.deleteERC20TokenMapping(sourceToken, {gasLimit: 10})).to.be.rejectedWith("out of gas");
            });
        });
    });
    describe("not owner", function () {
        describe("createERC20TokenMapping", function () {
            it("Should be reverted", async function () {
                const { assetsbridge, validator1 } = await getContract();
                await expect(assetsbridge.connect(validator1).createERC20TokenMapping(sourceToken, mezoToken)).to.be.rejectedWith("execution reverted");
            });
        });
        describe("deleteERC20TokenMapping", function () {
            it("Should be reverted", async function () {
                const { assetsbridge, validator1 } = await getContract();
                await expect(assetsbridge.connect(validator1).deleteERC20TokenMapping(sourceToken)).to.be.rejectedWith("execution reverted");
            });
        });
    });
});
