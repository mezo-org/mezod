import { MezoTransfers } from '../typechain-types/contracts/MezoTransfers';
import { expect } from "chai";
import hre from "hardhat";
import { ethers } from "hardhat"
import { getDeployedContract } from "./helpers/contract"
import abi from '../../../precompile/btctoken/abi.json'

const precompileAddress = '0x7b7c000000000000000000000000000000000000';

describe("MezoTransfers", function () {
  const { deployments } = hre;
  let btcErc20Token: any;
  let mezoTransfers: MezoTransfers;
  let signers: any;
  let senderAddress: string;
  let recipientAddress: string;

  beforeEach(async function () {
    await deployments.fixture();
    btcErc20Token = new hre.ethers.Contract(precompileAddress, abi, ethers.provider);
    mezoTransfers = await getDeployedContract("MezoTransfers");
    signers = await ethers.getSigners();
    senderAddress = signers[0].address;
    recipientAddress = ethers.Wallet.createRandom().address;
  });

  describe("nativeThenBTCERC20", function () {
    let initialSenderBalance: any;
    let initialRecipientBalance: any;
    let tokenAmount: any;
    let gasCost: any;

    beforeEach(async function () {
      tokenAmount = ethers.parseEther("8");
      const mezoTransfersAddress = await mezoTransfers.getAddress();

      const transferTx = await btcErc20Token.connect(signers[0]).transfer(mezoTransfersAddress, tokenAmount);
      await transferTx.wait();

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await mezoTransfers.connect(signers[0]).nativeThenBTCERC20(recipientAddress);
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * tx.gasPrice;
    });

    it("should verify sender native balance", async function () {
      const currentSenderNativeBalance = await ethers.provider.getBalance(senderAddress);
      expect(initialSenderBalance - gasCost).to.equal(currentSenderNativeBalance);
    });

    it("should verify recipient native balance", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
      expect(initialRecipientBalance).to.equal(0);
      expect(currentRecipientNativeBalance).to.equal(tokenAmount);
    });

    it("should verify native and BTC ERC20 balance equivalence", async function () {
      const currentSenderBTCERC20Balance = await btcErc20Token.balanceOf(senderAddress);
      const currentSenderNativeBalance = await ethers.provider.getBalance(senderAddress);
      expect(currentSenderNativeBalance).to.equal(currentSenderBTCERC20Balance);

      const currentRecipientBTCERC20Balance = await btcErc20Token.balanceOf(recipientAddress);
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
      expect(currentRecipientNativeBalance).to.equal(currentRecipientBTCERC20Balance);
    });

    it("should verify MezoTransfers contract has zero balance", async function () {
      const mezoTransfersAddress = await mezoTransfers.getAddress();
      const currentContractNativeBalance = await ethers.provider.getBalance(mezoTransfersAddress);
      expect(currentContractNativeBalance).to.equal(0);
    });
  });

  describe("btcERC20ThenNative", function () {
    let initialSenderBalance: any;
    let initialRecipientBalance: any;
    let tokenAmount: any;
    let gasCost: any;

    beforeEach(async function () {
      tokenAmount = ethers.parseEther("12");
      const mezoTransfersAddress = await mezoTransfers.getAddress();

      const transferTx = await btcErc20Token.connect(signers[0]).transfer(mezoTransfersAddress, tokenAmount);
      await transferTx.wait();

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await mezoTransfers.connect(signers[0]).btcERC20ThenNative(recipientAddress);
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * tx.gasPrice;
    });

    it("should verify sender native balance", async function () {
      const currentSenderNativeBalance = await ethers.provider.getBalance(senderAddress);
      expect(initialSenderBalance - gasCost).to.equal(currentSenderNativeBalance);
    });

    it("should verify recipient native balance", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
      expect(initialRecipientBalance).to.equal(0);
      expect(currentRecipientNativeBalance).to.equal(tokenAmount);
    });

    it("should verify MezoTransfers contract has zero balance", async function () {
      const mezoTransfersAddress = await mezoTransfers.getAddress();
      const currentContractNativeBalance = await ethers.provider.getBalance(mezoTransfersAddress);
      expect(currentContractNativeBalance).to.equal(0);
    });
  });

  describe("receiveAndSendNativeThenBTCERC20", function () {
    let initialSenderBalance: any;
    let initialRecipientBalance: any;
    const tokenAmount = ethers.parseEther("2");
    const nativeAmount = ethers.parseEther("3");
    let gasCost: any;
    let mezoTransfersAddress: any;

    beforeEach(async function () {
      mezoTransfersAddress = await mezoTransfers.getAddress();

      await btcErc20Token.connect(signers[0]).approve(mezoTransfersAddress, tokenAmount)
        .then(tx => tx.wait());

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      // Transfer
      const tx = await mezoTransfers.connect(signers[0]).receiveAndSendNativeThenBTCERC20(recipientAddress, tokenAmount, { value: nativeAmount });
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * tx.gasPrice;
    });

    it("should verify sender native balance deduction", async function () {
      const currentSenderNativeBalance = await ethers.provider.getBalance(senderAddress);
      expect(initialSenderBalance - nativeAmount - tokenAmount - gasCost).to.equal(currentSenderNativeBalance);
    });

    it("should verify recipient native balance received", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
      expect(initialRecipientBalance + nativeAmount + tokenAmount).to.equal(currentRecipientNativeBalance);
    });

    it("should verify MezoTransfers contract has zero balance", async function () {
      const currentBalance = await ethers.provider.getBalance(mezoTransfersAddress);
      expect(currentBalance).to.equal(0);
    });
  });

  describe("receiveAndSendBtcERC20ThenNative", function () {
    let initialSenderBalance: any;
    let initialRecipientBalance: any;
    const tokenAmount = ethers.parseEther("4");
    const nativeAmount = ethers.parseEther("5");
    let gasCost: any;
    let mezoTransfersAddress: any;

    beforeEach(async function () {
      mezoTransfersAddress = await mezoTransfers.getAddress();

      await btcErc20Token.connect(signers[0]).approve(mezoTransfersAddress, tokenAmount)
        .then(tx => tx.wait());

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      // Transfer
      const tx = await mezoTransfers.connect(signers[0]).receiveAndSendBtcERC20ThenNative(recipientAddress, tokenAmount, { value: nativeAmount });
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * tx.gasPrice;
    });

    it("should verify sender native balance deduction", async function () {
      const currentSenderNativeBalance = await ethers.provider.getBalance(senderAddress);
      expect(initialSenderBalance - nativeAmount- tokenAmount - gasCost).to.equal(currentSenderNativeBalance);
    });

    it("should verify recipient native balance received", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
      expect(initialRecipientBalance + nativeAmount + tokenAmount).to.equal(currentRecipientNativeBalance);
    });

    it("should verify MezoTransfers contract has zero balance", async function () {
      const currentBalance = await ethers.provider.getBalance(mezoTransfersAddress);
      expect(currentBalance).to.equal(0);
    });
  });
});