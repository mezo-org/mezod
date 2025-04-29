import { MEZOTransfers } from '../typechain-types/MEZOTransfers';
import { expect } from "chai";
import hre from "hardhat";
import { ethers } from "hardhat"
import { getDeployedContract } from "./helpers/contract"
import btcabi from '../../../precompile/btctoken/abi.json'
import mezoabi from '../../../precompile/mezotoken/abi.json'

const btcPrecompileAddress = '0x7b7c000000000000000000000000000000000000';
const mezoPrecompileAddress = '0x7b7c000000000000000000000000000000000001';

describe("MEZOTransfers", function () {
  const { deployments } = hre;
  let btcErc20Token: any;
  let mezoErc20Token: any;
  let mezoTransfers: MEZOTransfers;
  let sender: any;
  let recipientAddress: string;

  const fixture = (async function () {
    await deployments.fixture();
    btcErc20Token = new hre.ethers.Contract(btcPrecompileAddress, btcabi, ethers.provider);
    mezoErc20Token = new hre.ethers.Contract(mezoPrecompileAddress, mezoabi, ethers.provider);
    mezoTransfers = await getDeployedContract("MEZOTransfers");
    const signers = await ethers.getSigners();
    sender = signers[0];
    recipientAddress = ethers.Wallet.createRandom().address;
  });

  describe("transferSomeMEZO", function () {
    let mezoTransfersAddress: string;
    let mezoAmount: any;

    let initialSenderMEZOBalance: any;
    let initialSenderBTCBalance: any;
    let initialRecipientMEZOBalance: any;
    let initialRecipientBTCBalance: any;
    let initialContractMEZOBalance: any;
    let initialContractBTCBalance: any;

    let gasCost: any;

    before(async function () {
      await fixture();

      mezoTransfersAddress = await mezoTransfers.getAddress();
      mezoAmount = ethers.parseEther("8");

      const transferTx = await mezoErc20Token.connect(sender).transfer(mezoTransfersAddress, mezoAmount);
      await transferTx.wait();

      initialSenderMEZOBalance = await mezoErc20Token.balanceOf(sender.address);
      initialSenderBTCBalance = await ethers.provider.getBalance(sender.address);

      initialRecipientMEZOBalance = await mezoErc20Token.balanceOf(recipientAddress);
      initialRecipientBTCBalance = await ethers.provider.getBalance(recipientAddress);

      initialContractMEZOBalance = await mezoErc20Token.balanceOf(mezoTransfersAddress);
      initialContractBTCBalance = await ethers.provider.getBalance(mezoTransfersAddress);

      const tx = await mezoTransfers.connect(sender).transferSomeMEZO(recipientAddress);
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * receipt.gasPrice;
    });

    it("should decrease sender BTC balance by gas cost", async function () {
        const current = await ethers.provider.getBalance(sender.address);
        expect(current).to.equal(initialSenderBTCBalance - gasCost);
    });

    it("should not change sender MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(sender.address);
        expect(current).to.equal(initialSenderMEZOBalance);
    });

    it("should not change recipient BTC balance", async function () {
        const current = await ethers.provider.getBalance(recipientAddress);
        expect(current).to.equal(initialRecipientBTCBalance);
    });

    it("should properly increase recipient MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(recipientAddress);
        expect(current).to.equal(initialRecipientMEZOBalance + ethers.parseEther("4"));
    });

    it("should not change contract BTC balance", async function () {
        const current = await ethers.provider.getBalance(mezoTransfersAddress);
        expect(current).to.equal(initialContractBTCBalance);
    });

    it("should properly decrease contract MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(mezoTransfersAddress);
        expect(current).to.equal(initialContractMEZOBalance - ethers.parseEther("4"));
    });
  });

  describe("transferAllMEZO", function () {
    let mezoTransfersAddress: string;
    let mezoAmount: any;

    let initialSenderMEZOBalance: any;
    let initialSenderBTCBalance: any;
    let initialRecipientMEZOBalance: any;
    let initialRecipientBTCBalance: any;
    let initialContractMEZOBalance: any;
    let initialContractBTCBalance: any;

    let gasCost: any;

    before(async function () {
      await fixture();

      mezoTransfersAddress = await mezoTransfers.getAddress();
      mezoAmount = ethers.parseEther("8");

      const transferTx = await mezoErc20Token.connect(sender).transfer(mezoTransfersAddress, mezoAmount);
      await transferTx.wait();

      initialSenderMEZOBalance = await mezoErc20Token.balanceOf(sender.address);
      initialSenderBTCBalance = await ethers.provider.getBalance(sender.address);

      initialRecipientMEZOBalance = await mezoErc20Token.balanceOf(recipientAddress);
      initialRecipientBTCBalance = await ethers.provider.getBalance(recipientAddress);

      initialContractMEZOBalance = await mezoErc20Token.balanceOf(mezoTransfersAddress);
      initialContractBTCBalance = await ethers.provider.getBalance(mezoTransfersAddress);

      const tx = await mezoTransfers.connect(sender).transferAllMEZO(recipientAddress);
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * receipt.gasPrice;
    });

    it("should decrease sender BTC balance by gas cost", async function () {
        const current = await ethers.provider.getBalance(sender.address);
        expect(current).to.equal(initialSenderBTCBalance - gasCost);
    });

    it("should not change sender MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(sender.address);
        expect(current).to.equal(initialSenderMEZOBalance);
    });

    it("should not change recipient BTC balance", async function () {
        const current = await ethers.provider.getBalance(recipientAddress);
        expect(current).to.equal(initialRecipientBTCBalance);
    });

    it("should properly increase recipient MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(recipientAddress);
        expect(current).to.equal(initialRecipientMEZOBalance + ethers.parseEther("8"));
    });

    it("should not change contract BTC balance", async function () {
        const current = await ethers.provider.getBalance(mezoTransfersAddress);
        expect(current).to.equal(initialContractBTCBalance);
    });

    it("should properly decrease contract MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(mezoTransfersAddress);
        expect(current).to.equal(initialContractMEZOBalance - ethers.parseEther("8"));
    });
  });

  describe("pullMEZO", function () {
    let mezoTransfersAddress: string;
    let mezoAmount: any;

    let initialSenderMEZOBalance: any;
    let initialSenderBTCBalance: any;
    let initialContractMEZOBalance: any;
    let initialContractBTCBalance: any;

    let gasCost: any;

    before(async function () {
      await fixture();

      mezoTransfersAddress = await mezoTransfers.getAddress();
      mezoAmount = ethers.parseEther("8");

      const approveTx = await mezoErc20Token.connect(sender).approve(mezoTransfersAddress, mezoAmount);
      await approveTx.wait();

      initialSenderMEZOBalance = await mezoErc20Token.balanceOf(sender.address);
      initialSenderBTCBalance = await ethers.provider.getBalance(sender.address);

      initialContractMEZOBalance = await mezoErc20Token.balanceOf(mezoTransfersAddress);
      initialContractBTCBalance = await ethers.provider.getBalance(mezoTransfersAddress);

      const tx = await mezoTransfers.connect(sender).pullMEZO(sender.address, mezoAmount);
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * receipt.gasPrice;
    });

    it("should decrease sender BTC balance by gas cost", async function () {
        const current = await ethers.provider.getBalance(sender.address);
        expect(current).to.equal(initialSenderBTCBalance - gasCost);
    });

    it("should properly decrease sender MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(sender.address);
        expect(current).to.equal(initialSenderMEZOBalance - ethers.parseEther("8"));
    });

    it("should not change contract BTC balance", async function () {
        const current = await ethers.provider.getBalance(mezoTransfersAddress);
        expect(current).to.equal(initialContractBTCBalance);
    });

    it("should properly increase contract MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(mezoTransfersAddress);
        expect(current).to.equal(initialContractMEZOBalance + ethers.parseEther("8"));
    });
  });

  describe("pullMEZOToRecipient", function () {
    let mezoTransfersAddress: string;
    let mezoAmount: any;

    let initialSenderMEZOBalance: any;
    let initialSenderBTCBalance: any;
    let initialRecipientMEZOBalance: any;
    let initialRecipientBTCBalance: any;
    let initialContractMEZOBalance: any;
    let initialContractBTCBalance: any;

    let gasCost: any;

    before(async function () {
      await fixture();

      mezoTransfersAddress = await mezoTransfers.getAddress();
      mezoAmount = ethers.parseEther("8");

      const approveTx = await mezoErc20Token.connect(sender).approve(mezoTransfersAddress, mezoAmount);
      await approveTx.wait();

      initialSenderMEZOBalance = await mezoErc20Token.balanceOf(sender.address);
      initialSenderBTCBalance = await ethers.provider.getBalance(sender.address);

      initialRecipientMEZOBalance = await mezoErc20Token.balanceOf(recipientAddress);
      initialRecipientBTCBalance = await ethers.provider.getBalance(recipientAddress);

      initialContractMEZOBalance = await mezoErc20Token.balanceOf(mezoTransfersAddress);
      initialContractBTCBalance = await ethers.provider.getBalance(mezoTransfersAddress);

      const tx = await mezoTransfers.connect(sender).pullMEZOToRecipient(sender.address, recipientAddress, mezoAmount);
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * receipt.gasPrice;
    });

    it("should decrease sender BTC balance by gas cost", async function () {
        const current = await ethers.provider.getBalance(sender.address);
        expect(current).to.equal(initialSenderBTCBalance - gasCost);
    });

    it("should properly decrease sender MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(sender.address);
        expect(current).to.equal(initialSenderMEZOBalance - ethers.parseEther("8"));
    });

    it("should not change recipient BTC balance", async function () {
        const current = await ethers.provider.getBalance(recipientAddress);
        expect(current).to.equal(initialRecipientBTCBalance);
    });

    it("should properly increase recipient MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(recipientAddress);
        expect(current).to.equal(initialRecipientMEZOBalance + ethers.parseEther("8"));
    });

    it("should not change contract BTC balance", async function () {
        const current = await ethers.provider.getBalance(mezoTransfersAddress);
        expect(current).to.equal(initialContractBTCBalance);
    });

    it("should not change contract MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(mezoTransfersAddress);
        expect(current).to.equal(initialContractMEZOBalance);
    });
  });

  describe("receiveNativeThenTransferMEZO", function () {
    let mezoTransfersAddress: string;
    let mezoAmount: any;
    let btcAmount: any;
    
    let initialSenderMEZOBalance: any;
    let initialSenderBTCBalance: any;
    let initialRecipientMEZOBalance: any;
    let initialRecipientBTCBalance: any;
    let initialContractMEZOBalance: any;
    let initialContractBTCBalance: any;

    let gasCost: any;

    before(async function () {
      await fixture();

      mezoTransfersAddress = await mezoTransfers.getAddress();
      mezoAmount = ethers.parseEther("8");
      btcAmount = ethers.parseEther("6");

      const transferTx = await mezoErc20Token.connect(sender).transfer(mezoTransfersAddress, mezoAmount);
      await transferTx.wait();

      initialSenderMEZOBalance = await mezoErc20Token.balanceOf(sender.address);
      initialSenderBTCBalance = await ethers.provider.getBalance(sender.address);

      initialRecipientMEZOBalance = await mezoErc20Token.balanceOf(recipientAddress);
      initialRecipientBTCBalance = await ethers.provider.getBalance(recipientAddress);

      initialContractMEZOBalance = await mezoErc20Token.balanceOf(mezoTransfersAddress);
      initialContractBTCBalance = await ethers.provider.getBalance(mezoTransfersAddress);

      const tx = await mezoTransfers.connect(sender).receiveNativeThenTransferMEZO(recipientAddress, {value: btcAmount});
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * receipt.gasPrice;
    });

    it("should properly decrease sender BTC balance", async function () {
        const current = await ethers.provider.getBalance(sender.address);
        expect(current).to.equal(initialSenderBTCBalance - ethers.parseEther("6") - gasCost);
    });

    it("should not change sender MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(sender.address);
        expect(current).to.equal(initialSenderMEZOBalance);
    });

    it("should not change recipient BTC balance", async function () {
        const current = await ethers.provider.getBalance(recipientAddress);
        expect(current).to.equal(initialRecipientBTCBalance);
    });

    it("should properly increase recipient MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(recipientAddress);
        expect(current).to.equal(initialRecipientMEZOBalance + ethers.parseEther("8"));
    });

    it("should properly increase contract BTC balance", async function () {
        const current = await ethers.provider.getBalance(mezoTransfersAddress);
        expect(current).to.equal(initialContractBTCBalance + ethers.parseEther("6"));
    });

    it("should properly decrease contract MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(mezoTransfersAddress);
        expect(current).to.equal(initialContractMEZOBalance - ethers.parseEther("8"));
    });
  });

  describe("receiveNativeThenPullMEZO", function () {
    let mezoTransfersAddress: string;
    let mezoAmount: any;
    let btcAmount: any;
    
    let initialSenderMEZOBalance: any;
    let initialSenderBTCBalance: any;
    let initialContractMEZOBalance: any;
    let initialContractBTCBalance: any;

    let gasCost: any;

    before(async function () {
      await fixture();

      mezoTransfersAddress = await mezoTransfers.getAddress();
      mezoAmount = ethers.parseEther("8");
      btcAmount = ethers.parseEther("6");

      const approveTx = await mezoErc20Token.connect(sender).approve(mezoTransfersAddress, mezoAmount);
      await approveTx.wait();

      initialSenderMEZOBalance = await mezoErc20Token.balanceOf(sender.address);
      initialSenderBTCBalance = await ethers.provider.getBalance(sender.address);

      initialContractMEZOBalance = await mezoErc20Token.balanceOf(mezoTransfersAddress);
      initialContractBTCBalance = await ethers.provider.getBalance(mezoTransfersAddress);

      const tx = await mezoTransfers.connect(sender).receiveNativeThenPullMEZO(sender.address, mezoAmount, {value: btcAmount});
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * receipt.gasPrice;
    });

    it("should properly decrease sender BTC balance", async function () {
        const current = await ethers.provider.getBalance(sender.address);
        expect(current).to.equal(initialSenderBTCBalance - ethers.parseEther("6") - gasCost);
    });

    it("should properly decrease sender MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(sender.address);
        expect(current).to.equal(initialSenderMEZOBalance - ethers.parseEther("8"));
    });

    it("should properly increase contract BTC balance", async function () {
        const current = await ethers.provider.getBalance(mezoTransfersAddress);
        expect(current).to.equal(initialContractBTCBalance + ethers.parseEther("6"));
    });

    it("should properly increase contract MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(mezoTransfersAddress);
        expect(current).to.equal(initialContractMEZOBalance + ethers.parseEther("8"));
    });
  });

  describe("sendNativeThenTransferMEZO", function () {
    let mezoTransfersAddress: string;
    let mezoAmount: any;
    let btcAmount: any;

    let initialSenderMEZOBalance: any;
    let initialSenderBTCBalance: any;
    let initialRecipientMEZOBalance: any;
    let initialRecipientBTCBalance: any;
    let initialContractMEZOBalance: any;
    let initialContractBTCBalance: any;

    let gasCost: any;

    before(async function () {
      await fixture();

      mezoTransfersAddress = await mezoTransfers.getAddress();
      mezoAmount = ethers.parseEther("8");
      btcAmount = ethers.parseEther("6");

      const mezoTransferTx = await mezoErc20Token.connect(sender).transfer(mezoTransfersAddress, mezoAmount);
      await mezoTransferTx.wait();

      const btcTransferTx = await btcErc20Token.connect(sender).transfer(mezoTransfersAddress, btcAmount);
      await btcTransferTx.wait();

      initialSenderMEZOBalance = await mezoErc20Token.balanceOf(sender.address);
      initialSenderBTCBalance = await ethers.provider.getBalance(sender.address);

      initialRecipientMEZOBalance = await mezoErc20Token.balanceOf(recipientAddress);
      initialRecipientBTCBalance = await ethers.provider.getBalance(recipientAddress);

      initialContractMEZOBalance = await mezoErc20Token.balanceOf(mezoTransfersAddress);
      initialContractBTCBalance = await ethers.provider.getBalance(mezoTransfersAddress);

      const tx = await mezoTransfers.connect(sender).sendNativeThenTransferMEZO(recipientAddress);
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * receipt.gasPrice;
    });

    it("should decrease sender BTC balance by gas cost", async function () {
        const current = await ethers.provider.getBalance(sender.address);
        expect(current).to.equal(initialSenderBTCBalance - gasCost);
    });

    it("should not change sender MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(sender.address);
        expect(current).to.equal(initialSenderMEZOBalance);
    });

    it("should properly increase recipient BTC balance", async function () {
        const current = await ethers.provider.getBalance(recipientAddress);
        expect(current).to.equal(initialRecipientBTCBalance + ethers.parseEther("6"));
    });

    it("should properly increase recipient MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(recipientAddress);
        expect(current).to.equal(initialRecipientMEZOBalance + ethers.parseEther("8"));
    });

    it("should properly decrease contract BTC balance", async function () {
        const current = await ethers.provider.getBalance(mezoTransfersAddress);
        expect(current).to.equal(initialContractBTCBalance - ethers.parseEther("6"));
    });

    it("should properly decrease contract MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(mezoTransfersAddress);
        expect(current).to.equal(initialContractMEZOBalance - ethers.parseEther("8"));
    });
  });

  describe("transferMEZOThenSendNative", function () {
    let mezoTransfersAddress: string;
    let mezoAmount: any;
    let btcAmount: any;

    let initialSenderMEZOBalance: any;
    let initialSenderBTCBalance: any;
    let initialRecipientMEZOBalance: any;
    let initialRecipientBTCBalance: any;
    let initialContractMEZOBalance: any;
    let initialContractBTCBalance: any;

    let gasCost: any;

    before(async function () {
      await fixture();

      mezoTransfersAddress = await mezoTransfers.getAddress();
      mezoAmount = ethers.parseEther("8");
      btcAmount = ethers.parseEther("6");

      const mezoTransferTx = await mezoErc20Token.connect(sender).transfer(mezoTransfersAddress, mezoAmount);
      await mezoTransferTx.wait();

      const btcTransferTx = await btcErc20Token.connect(sender).transfer(mezoTransfersAddress, btcAmount);
      await btcTransferTx.wait();

      initialSenderMEZOBalance = await mezoErc20Token.balanceOf(sender.address);
      initialSenderBTCBalance = await ethers.provider.getBalance(sender.address);

      initialRecipientMEZOBalance = await mezoErc20Token.balanceOf(recipientAddress);
      initialRecipientBTCBalance = await ethers.provider.getBalance(recipientAddress);

      initialContractMEZOBalance = await mezoErc20Token.balanceOf(mezoTransfersAddress);
      initialContractBTCBalance = await ethers.provider.getBalance(mezoTransfersAddress);

      const tx = await mezoTransfers.connect(sender).transferMEZOThenSendNative(recipientAddress);
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * receipt.gasPrice;
    });

    it("should decrease sender BTC balance by gas cost", async function () {
        const current = await ethers.provider.getBalance(sender.address);
        expect(current).to.equal(initialSenderBTCBalance - gasCost);
    });

    it("should not change sender MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(sender.address);
        expect(current).to.equal(initialSenderMEZOBalance);
    });

    it("should properly increase recipient BTC balance", async function () {
        const current = await ethers.provider.getBalance(recipientAddress);
        expect(current).to.equal(initialRecipientBTCBalance + ethers.parseEther("6"));
    });

    it("should properly increase recipient MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(recipientAddress);
        expect(current).to.equal(initialRecipientMEZOBalance + ethers.parseEther("8"));
    });

    it("should properly decrease contract BTC balance", async function () {
        const current = await ethers.provider.getBalance(mezoTransfersAddress);
        expect(current).to.equal(initialContractBTCBalance - ethers.parseEther("6"));
    });

    it("should properly decrease contract MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(mezoTransfersAddress);
        expect(current).to.equal(initialContractMEZOBalance - ethers.parseEther("8"));
    });
  });

  describe("sendNativeThenPullMEZO", function () {
    let mezoTransfersAddress: string;
    let mezoAmount: any;
    let btcAmount: any;

    let initialSenderMEZOBalance: any;
    let initialSenderBTCBalance: any;
    let initialRecipientMEZOBalance: any;
    let initialRecipientBTCBalance: any;
    let initialContractMEZOBalance: any;
    let initialContractBTCBalance: any;

    let gasCost: any;

    before(async function () {
      await fixture();

      mezoTransfersAddress = await mezoTransfers.getAddress();
      mezoAmount = ethers.parseEther("8");
      btcAmount = ethers.parseEther("6");

      const mezoApproveTx = await mezoErc20Token.connect(sender).approve(mezoTransfersAddress, mezoAmount);
      await mezoApproveTx.wait();

      const btcTransferTx = await btcErc20Token.connect(sender).transfer(mezoTransfersAddress, btcAmount);
      await btcTransferTx.wait();

      initialSenderMEZOBalance = await mezoErc20Token.balanceOf(sender.address);
      initialSenderBTCBalance = await ethers.provider.getBalance(sender.address);

      initialRecipientMEZOBalance = await mezoErc20Token.balanceOf(recipientAddress);
      initialRecipientBTCBalance = await ethers.provider.getBalance(recipientAddress);

      initialContractMEZOBalance = await mezoErc20Token.balanceOf(mezoTransfersAddress);
      initialContractBTCBalance = await ethers.provider.getBalance(mezoTransfersAddress);

      const tx = await mezoTransfers.connect(sender).sendNativeThenPullMEZO(recipientAddress, sender.address, mezoAmount);
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * receipt.gasPrice;
    });

    it("should decrease sender BTC balance by gas cost", async function () {
        const current = await ethers.provider.getBalance(sender.address);
        expect(current).to.equal(initialSenderBTCBalance - gasCost);
    });

    it("should properly decrease sender MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(sender.address);
        expect(current).to.equal(initialSenderMEZOBalance - ethers.parseEther("8"));
    });

    it("should properly increase recipient BTC balance", async function () {
        const current = await ethers.provider.getBalance(recipientAddress);
        expect(current).to.equal(initialRecipientBTCBalance + ethers.parseEther("6"));
    });

    it("should not change recipient MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(recipientAddress);
        expect(current).to.equal(initialRecipientMEZOBalance);
    });

    it("should properly decrease contract BTC balance", async function () {
        const current = await ethers.provider.getBalance(mezoTransfersAddress);
        expect(current).to.equal(initialContractBTCBalance - ethers.parseEther("6"));
    });

    it("should properly increase contract MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(mezoTransfersAddress);
        expect(current).to.equal(initialContractMEZOBalance + ethers.parseEther("8"));
    });
  });

  describe("pullMEZOThenSendNative", function () {
    let mezoTransfersAddress: string;
    let mezoAmount: any;
    let btcAmount: any;

    let initialSenderMEZOBalance: any;
    let initialSenderBTCBalance: any;
    let initialRecipientMEZOBalance: any;
    let initialRecipientBTCBalance: any;
    let initialContractMEZOBalance: any;
    let initialContractBTCBalance: any;

    let gasCost: any;

    before(async function () {
      await fixture();

      mezoTransfersAddress = await mezoTransfers.getAddress();
      mezoAmount = ethers.parseEther("8");
      btcAmount = ethers.parseEther("6");

      const mezoApproveTx = await mezoErc20Token.connect(sender).approve(mezoTransfersAddress, mezoAmount);
      await mezoApproveTx.wait();

      const btcTransferTx = await btcErc20Token.connect(sender).transfer(mezoTransfersAddress, btcAmount);
      await btcTransferTx.wait();

      initialSenderMEZOBalance = await mezoErc20Token.balanceOf(sender.address);
      initialSenderBTCBalance = await ethers.provider.getBalance(sender.address);

      initialRecipientMEZOBalance = await mezoErc20Token.balanceOf(recipientAddress);
      initialRecipientBTCBalance = await ethers.provider.getBalance(recipientAddress);

      initialContractMEZOBalance = await mezoErc20Token.balanceOf(mezoTransfersAddress);
      initialContractBTCBalance = await ethers.provider.getBalance(mezoTransfersAddress);

      const tx = await mezoTransfers.connect(sender).pullMEZOThenSendNative(sender.address, mezoAmount, recipientAddress);
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * receipt.gasPrice;
    });

    it("should decrease sender BTC balance by gas cost", async function () {
        const current = await ethers.provider.getBalance(sender.address);
        expect(current).to.equal(initialSenderBTCBalance - gasCost);
    });

    it("should properly decrease sender MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(sender.address);
        expect(current).to.equal(initialSenderMEZOBalance - ethers.parseEther("8"));
    });

    it("should properly increase recipient BTC balance", async function () {
        const current = await ethers.provider.getBalance(recipientAddress);
        expect(current).to.equal(initialRecipientBTCBalance + ethers.parseEther("6"));
    });

    it("should not change recipient MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(recipientAddress);
        expect(current).to.equal(initialRecipientMEZOBalance);
    });

    it("should properly decrease contract BTC balance", async function () {
        const current = await ethers.provider.getBalance(mezoTransfersAddress);
        expect(current).to.equal(initialContractBTCBalance - ethers.parseEther("6"));
    });

    it("should properly increase contract MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(mezoTransfersAddress);
        expect(current).to.equal(initialContractMEZOBalance + ethers.parseEther("8"));
    });
  });

  describe("transferBTCThenTransferMEZO", function () {
    let mezoTransfersAddress: string;
    let mezoAmount: any;
    let btcAmount: any;

    let initialSenderMEZOBalance: any;
    let initialSenderBTCBalance: any;
    let initialRecipientMEZOBalance: any;
    let initialRecipientBTCBalance: any;
    let initialContractMEZOBalance: any;
    let initialContractBTCBalance: any;

    let gasCost: any;

    before(async function () {
      await fixture();

      mezoTransfersAddress = await mezoTransfers.getAddress();
      mezoAmount = ethers.parseEther("8");
      btcAmount = ethers.parseEther("6");

      const mezoTransferTx = await mezoErc20Token.connect(sender).transfer(mezoTransfersAddress, mezoAmount);
      await mezoTransferTx.wait();

      const btcTransferTx = await btcErc20Token.connect(sender).transfer(mezoTransfersAddress, btcAmount);
      await btcTransferTx.wait();

      initialSenderMEZOBalance = await mezoErc20Token.balanceOf(sender.address);
      initialSenderBTCBalance = await ethers.provider.getBalance(sender.address);

      initialRecipientMEZOBalance = await mezoErc20Token.balanceOf(recipientAddress);
      initialRecipientBTCBalance = await ethers.provider.getBalance(recipientAddress);

      initialContractMEZOBalance = await mezoErc20Token.balanceOf(mezoTransfersAddress);
      initialContractBTCBalance = await ethers.provider.getBalance(mezoTransfersAddress);

      const tx = await mezoTransfers.connect(sender).transferBTCThenTransferMEZO(recipientAddress);
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * receipt.gasPrice;
    });

    it("should decrease sender BTC balance by gas cost", async function () {
        const current = await ethers.provider.getBalance(sender.address);
        expect(current).to.equal(initialSenderBTCBalance - gasCost);
    });

    it("should not change sender MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(sender.address);
        expect(current).to.equal(initialSenderMEZOBalance);
    });

    it("should properly increase recipient BTC balance", async function () {
        const current = await ethers.provider.getBalance(recipientAddress);
        expect(current).to.equal(initialRecipientBTCBalance + ethers.parseEther("6"));
    });

    it("should properly increase recipient MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(recipientAddress);
        expect(current).to.equal(initialRecipientMEZOBalance + ethers.parseEther("8"));
    });

    it("should properly decrease contract BTC balance", async function () {
        const current = await ethers.provider.getBalance(mezoTransfersAddress);
        expect(current).to.equal(initialContractBTCBalance - ethers.parseEther("6"));
    });

    it("should properly decrease contract MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(mezoTransfersAddress);
        expect(current).to.equal(initialContractMEZOBalance - ethers.parseEther("8"));
    });
  });

  describe("transferMEZOThenTransferBTC", function () {
    let mezoTransfersAddress: string;
    let mezoAmount: any;
    let btcAmount: any;

    let initialSenderMEZOBalance: any;
    let initialSenderBTCBalance: any;
    let initialRecipientMEZOBalance: any;
    let initialRecipientBTCBalance: any;
    let initialContractMEZOBalance: any;
    let initialContractBTCBalance: any;

    let gasCost: any;

    before(async function () {
      await fixture();

      mezoTransfersAddress = await mezoTransfers.getAddress();
      mezoAmount = ethers.parseEther("8");
      btcAmount = ethers.parseEther("6");

      const mezoTransferTx = await mezoErc20Token.connect(sender).transfer(mezoTransfersAddress, mezoAmount);
      await mezoTransferTx.wait();

      const btcTransferTx = await btcErc20Token.connect(sender).transfer(mezoTransfersAddress, btcAmount);
      await btcTransferTx.wait();

      initialSenderMEZOBalance = await mezoErc20Token.balanceOf(sender.address);
      initialSenderBTCBalance = await ethers.provider.getBalance(sender.address);

      initialRecipientMEZOBalance = await mezoErc20Token.balanceOf(recipientAddress);
      initialRecipientBTCBalance = await ethers.provider.getBalance(recipientAddress);

      initialContractMEZOBalance = await mezoErc20Token.balanceOf(mezoTransfersAddress);
      initialContractBTCBalance = await ethers.provider.getBalance(mezoTransfersAddress);

      const tx = await mezoTransfers.connect(sender).transferMEZOThenTransferBTC(recipientAddress);
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * receipt.gasPrice;
    });

    it("should decrease sender BTC balance by gas cost", async function () {
        const current = await ethers.provider.getBalance(sender.address);
        expect(current).to.equal(initialSenderBTCBalance - gasCost);
    });

    it("should not change sender MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(sender.address);
        expect(current).to.equal(initialSenderMEZOBalance);
    });

    it("should properly increase recipient BTC balance", async function () {
        const current = await ethers.provider.getBalance(recipientAddress);
        expect(current).to.equal(initialRecipientBTCBalance + ethers.parseEther("6"));
    });

    it("should properly increase recipient MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(recipientAddress);
        expect(current).to.equal(initialRecipientMEZOBalance + ethers.parseEther("8"));
    });

    it("should properly decrease contract BTC balance", async function () {
        const current = await ethers.provider.getBalance(mezoTransfersAddress);
        expect(current).to.equal(initialContractBTCBalance - ethers.parseEther("6"));
    });

    it("should properly decrease contract MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(mezoTransfersAddress);
        expect(current).to.equal(initialContractMEZOBalance - ethers.parseEther("8"));
    });
  });

  describe("pullBTCThenPullMEZO", function () {
    let mezoTransfersAddress: string;
    let mezoAmount: any;
    let btcAmount: any;

    let initialSenderMEZOBalance: any;
    let initialSenderBTCBalance: any;
    let initialContractMEZOBalance: any;
    let initialContractBTCBalance: any;

    let gasCost: any;

    before(async function () {
      await fixture();

      mezoTransfersAddress = await mezoTransfers.getAddress();
      mezoAmount = ethers.parseEther("8");
      btcAmount = ethers.parseEther("6");

      const mezoApproveTx = await mezoErc20Token.connect(sender).approve(mezoTransfersAddress, mezoAmount);
      await mezoApproveTx.wait();

      const btcApproveTx = await btcErc20Token.connect(sender).approve(mezoTransfersAddress, btcAmount);
      await btcApproveTx.wait();

      initialSenderMEZOBalance = await mezoErc20Token.balanceOf(sender.address);
      initialSenderBTCBalance = await ethers.provider.getBalance(sender.address);

      initialContractMEZOBalance = await mezoErc20Token.balanceOf(mezoTransfersAddress);
      initialContractBTCBalance = await ethers.provider.getBalance(mezoTransfersAddress);

      const tx = await mezoTransfers.connect(sender).pullBTCThenPullMEZO(sender, btcAmount, mezoAmount);
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * receipt.gasPrice;
    });

    it("should properly decrease sender BTC balance", async function () {
        const current = await ethers.provider.getBalance(sender.address);
        expect(current).to.equal(initialSenderBTCBalance - ethers.parseEther("6") - gasCost);
    });

    it("should properly decrease sender MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(sender.address);
        expect(current).to.equal(initialSenderMEZOBalance - ethers.parseEther("8"));
    });

    it("should properly increase contract BTC balance", async function () {
        const current = await ethers.provider.getBalance(mezoTransfersAddress);
        expect(current).to.equal(initialContractBTCBalance + ethers.parseEther("6"));
    });

    it("should properly increase contract MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(mezoTransfersAddress);
        expect(current).to.equal(initialContractMEZOBalance + ethers.parseEther("8"));
    });
  });

  describe("pullMEZOThenPullBTC", function () {
    let mezoTransfersAddress: string;
    let mezoAmount: any;
    let btcAmount: any;

    let initialSenderMEZOBalance: any;
    let initialSenderBTCBalance: any;
    let initialContractMEZOBalance: any;
    let initialContractBTCBalance: any;

    let gasCost: any;

    before(async function () {
      await fixture();

      mezoTransfersAddress = await mezoTransfers.getAddress();
      mezoAmount = ethers.parseEther("8");
      btcAmount = ethers.parseEther("6");

      const mezoApproveTx = await mezoErc20Token.connect(sender).approve(mezoTransfersAddress, mezoAmount);
      await mezoApproveTx.wait();

      const btcApproveTx = await btcErc20Token.connect(sender).approve(mezoTransfersAddress, btcAmount);
      await btcApproveTx.wait();

      initialSenderMEZOBalance = await mezoErc20Token.balanceOf(sender.address);
      initialSenderBTCBalance = await ethers.provider.getBalance(sender.address);

      initialContractMEZOBalance = await mezoErc20Token.balanceOf(mezoTransfersAddress);
      initialContractBTCBalance = await ethers.provider.getBalance(mezoTransfersAddress);

      const tx = await mezoTransfers.connect(sender).pullMEZOThenPullBTC(sender, mezoAmount, btcAmount);
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * receipt.gasPrice;
    });

    it("should properly decrease sender BTC balance", async function () {
        const current = await ethers.provider.getBalance(sender.address);
        expect(current).to.equal(initialSenderBTCBalance - ethers.parseEther("6") - gasCost);
    });

    it("should properly decrease sender MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(sender.address);
        expect(current).to.equal(initialSenderMEZOBalance - ethers.parseEther("8"));
    });

    it("should properly increase contract BTC balance", async function () {
        const current = await ethers.provider.getBalance(mezoTransfersAddress);
        expect(current).to.equal(initialContractBTCBalance + ethers.parseEther("6"));
    });

    it("should properly increase contract MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(mezoTransfersAddress);
        expect(current).to.equal(initialContractMEZOBalance + ethers.parseEther("8"));
    });
  });

  describe("sendNativeThenTransferBTCThenTransferMEZO", function () {
    let mezoTransfersAddress: string;
    let mezoAmount: any;
    let btcAmount: any;

    let initialSenderMEZOBalance: any;
    let initialSenderBTCBalance: any;
    let initialRecipientMEZOBalance: any;
    let initialRecipientBTCBalance: any;
    let initialContractMEZOBalance: any;
    let initialContractBTCBalance: any;

    let gasCost: any;

    before(async function () {
      await fixture();

      mezoTransfersAddress = await mezoTransfers.getAddress();
      mezoAmount = ethers.parseEther("8");
      btcAmount = ethers.parseEther("6");

      const mezoTransferTx = await mezoErc20Token.connect(sender).transfer(mezoTransfersAddress, mezoAmount);
      await mezoTransferTx.wait();

      const btcTransferTx = await btcErc20Token.connect(sender).transfer(mezoTransfersAddress, btcAmount);
      await btcTransferTx.wait();

      initialSenderMEZOBalance = await mezoErc20Token.balanceOf(sender.address);
      initialSenderBTCBalance = await ethers.provider.getBalance(sender.address);

      initialRecipientMEZOBalance = await mezoErc20Token.balanceOf(recipientAddress);
      initialRecipientBTCBalance = await ethers.provider.getBalance(recipientAddress);

      initialContractMEZOBalance = await mezoErc20Token.balanceOf(mezoTransfersAddress);
      initialContractBTCBalance = await ethers.provider.getBalance(mezoTransfersAddress);

      const tx = await mezoTransfers.connect(sender).sendNativeThenTransferBTCThenTransferMEZO(recipientAddress);
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * receipt.gasPrice;
    });

    it("should decrease sender BTC balance by gas cost", async function () {
        const current = await ethers.provider.getBalance(sender.address);
        expect(current).to.equal(initialSenderBTCBalance - gasCost);
    });

    it("should not change sender MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(sender.address);
        expect(current).to.equal(initialSenderMEZOBalance);
    });

    it("should properly increase recipient BTC balance", async function () {
        const current = await ethers.provider.getBalance(recipientAddress);
        expect(current).to.equal(initialRecipientBTCBalance + ethers.parseEther("6"));
    });

    it("should properly increase recipient MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(recipientAddress);
        expect(current).to.equal(initialRecipientMEZOBalance + ethers.parseEther("8"));
    });

    it("should properly decrease contract BTC balance", async function () {
        const current = await ethers.provider.getBalance(mezoTransfersAddress);
        expect(current).to.equal(initialContractBTCBalance - ethers.parseEther("6"));
    });

    it("should properly decrease contract MEZO balance", async function () {
        const current = await mezoErc20Token.balanceOf(mezoTransfersAddress);
        expect(current).to.equal(initialContractMEZOBalance - ethers.parseEther("8"));
    });
  });
})
