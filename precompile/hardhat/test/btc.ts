import { expect } from "chai";
import hre from "hardhat";
import deployed from "../ignition/deployments/chain-31611/deployed_addresses.json"
import { MaxUint256 } from "ethers";

describe("BTC", function () {
    // get deployed contract address from ignition
    const contractAddress = deployed["BTC#BTCCaller"];

    async function getContract() {
       const btc = await hre.ethers.getContractAt("BTCCaller", contractAddress);
       const [owner, validator1, validator2, validator3 ] = await hre.ethers.getSigners();
       return { btc, owner, validator1, validator2, validator3 }
    }
  
    describe("out of gas", function () {
        describe("transfer", function() {
            it("Should be rejected", async function () {
                const { btc, validator1 } = await getContract();
                await expect(btc.transfer(validator1, 20000, {gasLimit: 10})).to.be.rejectedWith("out of gas");
            });
        });
        describe("transferFrom", function() {
            it("Should be rejected", async function () {
                const { btc, owner, validator1 } = await getContract();
                await expect(btc.transferFrom(validator1, owner, 20000, {gasLimit: 10})).to.be.rejectedWith("out of gas");
            });
        });
        describe("approve", function() {
            it("Should be rejected", async function () {
                const { btc, validator1 } = await getContract();
                await expect(btc.approve(validator1, 20000, {gasLimit: 10})).to.be.rejectedWith("out of gas");
            });
        });
        describe("permit", function() {
            it("Should be rejected", async function () {
                const { btc, owner, validator1 } = await getContract();
                await expect(btc.permit(validator1, owner, 20000, 1893456000, 27, "0xb3b6b3e4ad32709cac57c5f4bcd86fb63ca0b292c6bd027874b68d36a8794168", "0x0c3dc80556d8b8866fb5a5b757ed6536177fc95c2b4f36a89e5c67a714ba98a2", {gasLimit: 10})).to.be.rejectedWith("out of gas");
            });
        });
    });
    describe("out of bounds", function () {
        describe("transfer", function() {
            it("Should be rejected (negative)", async function () {
                const { btc, validator1 } = await getContract();
                await expect(btc.transfer(validator1, -20000)).to.be.rejectedWith("value out-of-bounds");
            });
        });
        describe("approve", function() {
            it("Should be rejected (negative)", async function () {
                const { btc, validator1 } = await getContract();
                await expect(btc.approve(validator1, -20000)).to.be.rejectedWith("value out-of-bounds");
            });
            it("Should be rejected (overflow)", async function () {
                const { btc, validator1 } = await getContract();
                await expect(btc.approve(validator1, MaxUint256+1n)).to.be.rejectedWith("value out-of-bounds");
            });
        });
    });
});
