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

  describe("receiveSendNative", function () {
    let initialSenderBalance: any;
    let initialRecipientBalance: any;
    let nativeAmount: any;
    let gasCost: any;

    beforeEach(async function () {
      nativeAmount = ethers.parseEther("10");

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await mezoTransfers.connect(signers[0]).receiveSendNative(recipientAddress, { value: nativeAmount });
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * tx.gasPrice;
    });

    it("should verify sender native balance", async function () {
      const currentSenderNativeBalance = await ethers.provider.getBalance(senderAddress);
      expect(initialSenderBalance - nativeAmount - gasCost).to.equal(currentSenderNativeBalance);
    });

    it("should verify recipient native balance", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
      expect(initialRecipientBalance).to.equal(0);
      expect(currentRecipientNativeBalance).to.equal(nativeAmount);
    });

    it("should verify MezoTransfers contract has zero balance", async function () {
      const mezoTransfersAddress = await mezoTransfers.getAddress();
      const currentContractNativeBalance = await ethers.provider.getBalance(mezoTransfersAddress);
      expect(currentContractNativeBalance).to.equal(0);
    });
  });

  describe("receiveSendBTCERC20", function () {
    let initialSenderBalance: any;
    let initialRecipientBalance: any;
    let nativeAmount: any;
    let gasCost: any;
    let mezoTransfersAddress: any;

    beforeEach(async function () {
      nativeAmount = ethers.parseEther("11");
      mezoTransfersAddress = await mezoTransfers.getAddress();

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await mezoTransfers.connect(signers[0]).receiveSendBTCERC20(recipientAddress, { value: nativeAmount });
      const receipt = await tx.wait();

      gasCost = receipt.gasUsed * tx.gasPrice;
    });

    it("should verify sender native balance", async function () {
      const currentSenderNativeBalance = await ethers.provider.getBalance(senderAddress);
      expect(initialSenderBalance - nativeAmount - gasCost).to.equal(currentSenderNativeBalance);
    });

    it("should verify recipient native balance", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
      expect(initialRecipientBalance).to.equal(0);
      expect(currentRecipientNativeBalance).to.equal(nativeAmount);
    });

    it("should verify MezoTransfers contract has zero balance", async function () {
      const currentContractBalance = await ethers.provider.getBalance(mezoTransfersAddress);
      expect(currentContractBalance).to.equal(0);
    });
  });

  describe("receiveSendNativeThenBTCERC20", function () {
    let initialSenderBalance: any;
    let initialRecipientBalance: any;
    let gasCost: any;
    let mezoTransfersAddress: any;
    const nativeAmount = ethers.parseEther("3");

    beforeEach(async function () {
      mezoTransfersAddress = await mezoTransfers.getAddress();

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await mezoTransfers.connect(signers[0]).receiveSendNativeThenBTCERC20(recipientAddress, { value: nativeAmount });
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * tx.gasPrice;
    });

    it("should verify sender native balance deduction", async function () {
      const currentSenderNativeBalance = await ethers.provider.getBalance(senderAddress);
      expect(initialSenderBalance - nativeAmount - gasCost).to.equal(currentSenderNativeBalance);
    });

    it("should verify recipient native balance received", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
      expect(initialRecipientBalance + nativeAmount).to.equal(currentRecipientNativeBalance);
    });

    it("should verify MezoTransfers contract has zero balance", async function () {
      const currentBalance = await ethers.provider.getBalance(mezoTransfersAddress);
      expect(currentBalance).to.equal(0);
    });
  });

  describe("receiveSendBTCERC20ThenNative", function () {
    const nativeAmount = ethers.parseEther("5");
    let initialSenderBalance: any;
    let initialRecipientBalance: any;
    let gasCost: any;
    let mezoTransfersAddress: any;

    beforeEach(async function () {
      mezoTransfersAddress = await mezoTransfers.getAddress();

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await mezoTransfers.connect(signers[0]).receiveSendBTCERC20ThenNative(recipientAddress, { value: nativeAmount });
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * tx.gasPrice;
    });

    it("should verify sender native balance deduction", async function () {
      const currentSenderNativeBalance = await ethers.provider.getBalance(senderAddress);
      expect(initialSenderBalance - nativeAmount - gasCost).to.equal(currentSenderNativeBalance);
    });

    it("should verify recipient native balance received", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
      expect(initialRecipientBalance + nativeAmount).to.equal(currentRecipientNativeBalance);
    });

    it("should verify MezoTransfers contract has zero balance", async function () {
      const currentBalance = await ethers.provider.getBalance(mezoTransfersAddress);
      expect(currentBalance).to.equal(0);
    });
  });

  describe("multipleBTCERC20AndNative", function () {
    let initialSenderBalance: any;
    let initialRecipientBalance: any;
    const nativeAmount = ethers.parseEther("8");
    let gasCost: any;
    let mezoTransfersAddress: any;

    beforeEach(async function () {
      mezoTransfersAddress = await mezoTransfers.getAddress();

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await mezoTransfers.connect(signers[0]).multipleBTCERC20AndNative(recipientAddress, { value: nativeAmount });
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * tx.gasPrice;
    });

    it("should verify sender native balance deduction", async function () {
      const currentSenderBalance = await ethers.provider.getBalance(senderAddress);
      expect(initialSenderBalance - nativeAmount - gasCost).to.equal(currentSenderBalance);
    });

    it("should verify recipient native balance received", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
      expect(initialRecipientBalance + nativeAmount).to.equal(currentRecipientNativeBalance);
    });

    it("should verify MezoTransfers contract has zero balance", async function () {
      const currentBalance = await ethers.provider.getBalance(mezoTransfersAddress);
      expect(currentBalance).to.equal(0);
    });
  });

  describe("multipleBTCERC20AndNative_tinyAmount", function () {
    let initialSenderBalance: any;
    let initialRecipientBalance: any;
    const nativeAmount = 4n; // this amount is split by 4 in contract
    let gasCost: any;
    let mezoTransfersAddress: any;

    beforeEach(async function () {
      mezoTransfersAddress = await mezoTransfers.getAddress();

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      // Transfer
      const tx = await mezoTransfers.connect(signers[0]).multipleBTCERC20AndNative(recipientAddress, { value: nativeAmount });
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * tx.gasPrice;
    });

    it("should verify sender native balance deduction", async function () {
      const currentSenderBalance = await ethers.provider.getBalance(senderAddress);
      expect(initialSenderBalance - nativeAmount - gasCost).to.equal(currentSenderBalance);
    });

    it("should verify recipient native balance received", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
      expect(initialRecipientBalance + nativeAmount).to.equal(currentRecipientNativeBalance);
    });

    it("should verify MezoTransfers contract has zero balance", async function () {
      const currentBalance = await ethers.provider.getBalance(mezoTransfersAddress);
      expect(currentBalance).to.equal(0);
    });
  });

  describe("multipleBTCERC20AndNative_hugeAmount", function () {
    let initialSenderBalance: any;
    let initialRecipientBalance: any;
    const nativeAmount = 5250000n; // this amount is split by 4 in contract
    let gasCost: any;
    let mezoTransfersAddress: any;

    beforeEach(async function () {
      mezoTransfersAddress = await mezoTransfers.getAddress();

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await mezoTransfers.connect(signers[0]).multipleBTCERC20AndNative(recipientAddress, { value: nativeAmount });
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * tx.gasPrice;
    });

    it("should verify sender native balance deduction", async function () {
      const currentSenderBalance = await ethers.provider.getBalance(senderAddress);
      expect(initialSenderBalance - nativeAmount - gasCost).to.equal(currentSenderBalance);
    });

    it("should verify recipient native balance received", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
      expect(initialRecipientBalance + nativeAmount).to.equal(currentRecipientNativeBalance);
    });

    it("should verify MezoTransfers contract has zero balance", async function () {
      const currentBalance = await ethers.provider.getBalance(mezoTransfersAddress);
      expect(currentBalance).to.equal(0);
    });
  });

  describe("transferWithRevert", function () {
    let initialSenderBalance: any;
    let initialRecipientBalance: any;
    const nativeAmount = 42;
    let mezoTransfersAddress: any;

    beforeEach(async function () {
      mezoTransfersAddress = await mezoTransfers.getAddress();

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      try {
        // Transfer
        await mezoTransfers
          .connect(signers[0])
          .transferWithRevert(recipientAddress, { value: nativeAmount });
      } catch (error) {
        expect(error.message).to.include("revert");
      }
    });

    it("should verify sender native balance deducted with gas cost", async function () {
      const currentSenderBalance = await ethers.provider.getBalance(senderAddress);
      expect(initialSenderBalance).to.equal(currentSenderBalance);
    });

    it("should verify recipient native balance remained the same", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
      expect(initialRecipientBalance).to.equal(currentRecipientNativeBalance);
    });

    it("should verify MezoTransfers contract has zero balance", async function () {
      const currentBalance = await ethers.provider.getBalance(mezoTransfersAddress);
      expect(currentBalance).to.equal(0);
    });
  });

  describe("transferNativeThenBTCERC20", function () {
    let initialSenderBalance: any;
    let initialRecipientBalance: any;
    let gasCost: any;
    let mezoTransfersAddress: any;
    const nativeAmount = ethers.parseEther("3");
    const tokenAmount = ethers.parseEther("1");

    beforeEach(async function () {
      mezoTransfersAddress = await mezoTransfers.getAddress();

      await btcErc20Token.connect(signers[0]).approve(mezoTransfersAddress, tokenAmount)
        .then(tx => tx.wait());

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await mezoTransfers.connect(signers[0]).transferNativeThenBTCERC20(recipientAddress, tokenAmount, { value: nativeAmount });
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * tx.gasPrice;
    });

    it("should verify sender native balance deduction", async function () {
      const currentSenderNativeBalance = await ethers.provider.getBalance(senderAddress);
      expect(initialSenderBalance - tokenAmount - nativeAmount - gasCost).to.equal(currentSenderNativeBalance);
    });

    it("should verify recipient native balance received", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
      expect(initialRecipientBalance + tokenAmount + nativeAmount).to.equal(currentRecipientNativeBalance);
    });

    it("should verify MezoTransfers contract has zero balance", async function () {
      const currentBalance = await ethers.provider.getBalance(mezoTransfersAddress);
      expect(currentBalance).to.equal(0);
    });
  });

  describe("transferBTCERC20ThenNative", function () {
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

      const tx = await mezoTransfers.connect(signers[0]).transferBTCERC20ThenNative(recipientAddress, tokenAmount, { value: nativeAmount });
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

  describe("storageUpdateAndRevert", function () {
    let initialContractBalance: any;
    let initialRecipientBalance: any;
    let amount: any;
    let gasCost: any;
    let mezoTransfersAddress: any;

    beforeEach(async function () {
      amount = ethers.parseEther("2");
      // Fund the contract first
      mezoTransfersAddress = await mezoTransfers.getAddress();
      await btcErc20Token.connect(signers[0]).transfer(mezoTransfersAddress, amount)
        .then(tx => tx.wait());

      initialContractBalance = await ethers.provider.getBalance(mezoTransfersAddress);
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await mezoTransfers.connect(signers[0]).storageStateTransition(recipientAddress);
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * tx.gasPrice;
    });

    it("should verify recipient native balance increased", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
      expect(initialRecipientBalance).to.equal(0);
      expect(currentRecipientNativeBalance).to.equal(amount);
    });

    it("should verify MezoTransfers contract has non-zero at the beginning of the call", async function () {
      expect(initialContractBalance).to.equal(amount);
    });

    it("should verify MezoTransfers contract has zero balance at the end of the call", async function () {
      const currentContractBalance = await ethers.provider.getBalance(mezoTransfersAddress);
      expect(currentContractBalance).to.equal(0);
    });

    it("should verify balanceTracker storage variable was unchanged", async function () {
      const balanceTracker = await mezoTransfers.balanceTracker();
      expect(balanceTracker).to.equal(0);
    });
  });
});