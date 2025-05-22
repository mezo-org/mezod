import { BTCTransfers } from '../typechain-types/BTCTransfers.sol/BTCTransfers';
import { expect } from "chai";
import hre from "hardhat";
import { ethers } from "hardhat"
import { getDeployedContract } from "./helpers/contract"
import abi from '../../../precompile/btctoken/abi.json'

const precompileAddress = '0x7b7c000000000000000000000000000000000000';

describe("BTCTransfers", function () {
  const { deployments } = hre;
  let btcErc20Token: any;
  let btcTransfers: BTCTransfers;
  let otherSpender: string;
  let signers: any;
  let senderSigner: any;
  let senderAddress: string;
  let recipientAddress: string;

  const fixture = (async function () {
    await deployments.fixture();
    btcErc20Token = new hre.ethers.Contract(precompileAddress, abi, ethers.provider);
    btcTransfers = await getDeployedContract("BTCTransfers");
    otherSpender = await(await getDeployedContract("OtherSpender")).getAddress();
    signers = await ethers.getSigners();
    senderSigner = signers[0];
    senderAddress = senderSigner.address;
    recipientAddress = ethers.Wallet.createRandom().address;
  });

  describe("approveZeroBeforeUserHaveEverApproved", function () {
    let receipt: any;
    let tx: any;

    before(async function () {
      await fixture();

      // we do a first approve, before the user ever
      // tried to approve for an actual amount
      tx =  await btcErc20Token.connect(senderSigner)
        .approve(otherSpender, 0, {gasLimit: 1000000});
      await tx.wait();

      receipt = await ethers.provider.getTransactionReceipt(tx.hash);
    });

    it("should verify the transaction didn't revert", async function () {
	expect(receipt!.status).to.equal(1);
    });
  });

  describe("multipleApproveInARow", function () {
    let receipt1: any;
    let receipt2: any;
    let receipt3: any;
    let tx1: any;
    let tx2: any;
    let tx3: any;

    before(async function () {
      await fixture();

      // Here we try to first do an approve, then reset it with 0, then approve again with 0.
      // All transactions should pass.
      tx1 =  await btcErc20Token.connect(senderSigner)
        .approve(otherSpender, 10, {gasLimit: 1000000});
      await tx1.wait();
      tx2 =  await btcErc20Token.connect(senderSigner)
        .approve(otherSpender, 0, {gasLimit: 1000000});
      await tx2.wait();
      tx3 =  await btcErc20Token.connect(senderSigner)
	.approve(otherSpender, 0, {gasLimit: 1000000});
      await tx3.wait();

      receipt1 = await ethers.provider.getTransactionReceipt(tx1.hash);
      receipt2 = await ethers.provider.getTransactionReceipt(tx2.hash);
      receipt3 = await ethers.provider.getTransactionReceipt(tx3.hash);
    });

    it("should verify the transactions didn't revert", async function () {
	expect(receipt1!.status).to.equal(1);
	expect(receipt2!.status).to.equal(1);
	expect(receipt3!.status).to.equal(1);
    });
  });

  describe("basicSpendTo0", function () {
    let initialSenderBalance: any;
    let initialRecipientBalance: any;
    let tokenAmount: any;
    let gasCost: any;

    before(async function () {
      await fixture();

      tokenAmount = ethers.parseEther("8");
      const btcTransfersAddress = await btcTransfers.getAddress();

      const transferTx = await btcErc20Token.connect(senderSigner).transfer(btcTransfersAddress, tokenAmount);
      await transferTx.wait();

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(otherSpender);

      const tx = await btcTransfers.connect(senderSigner).basicSpendTo0(otherSpender);
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * receipt.gasPrice;
    });

    it("should verify initial recipient balance is zero", async function () {
	expect(initialRecipientBalance).to.equal(0);
    });

    it("should verify current recipient balance", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(otherSpender);
      expect(currentRecipientNativeBalance).to.equal(tokenAmount / 2n);
    });

    it("should verify native and BTC ERC20 balance equivalence", async function () {
      const currentSenderBTCERC20Balance = await btcErc20Token.balanceOf(senderAddress);
      const currentSenderNativeBalance = await ethers.provider.getBalance(senderAddress);
      expect(currentSenderNativeBalance).to.equal(currentSenderBTCERC20Balance);

      const currentRecipientBTCERC20Balance = await btcErc20Token.balanceOf(recipientAddress);
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
      expect(currentRecipientNativeBalance).to.equal(currentRecipientBTCERC20Balance);
    });

    it("should verify BTCTransfers contract has half balance", async function () {
      const btcTransfersAddress = await btcTransfers.getAddress();
      const currentContractNativeBalance = await ethers.provider.getBalance(btcTransfersAddress);
      expect(currentContractNativeBalance).to.equal(tokenAmount / 2n);
    });
  });

  describe("basicSendThenSpendAndReturn", function () {
    let initialSenderBalance: any;
    let initialRecipientBalance: any;
    let tokenAmount: any;
    let gasCost: any;

    before(async function () {
      await fixture();

      tokenAmount = ethers.parseEther("8");
      const btcTransfersAddress = await btcTransfers.getAddress();

      const transferTx = await btcErc20Token.connect(senderSigner).transfer(btcTransfersAddress, tokenAmount);
      await transferTx.wait();

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(otherSpender);

      const tx = await btcTransfers.connect(senderSigner).basicSendThenSpendAndReturn(otherSpender);
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * receipt.gasPrice;
    });

    it("should verify initial recipient balance is zero", async function () {
      expect(initialRecipientBalance).to.equal(0);
    });

    it("should verify current recipient balance", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(otherSpender);
      expect(currentRecipientNativeBalance).to.equal(0n);
    });

    it("should verify native and BTC ERC20 balance equivalence", async function () {
      const currentSenderBTCERC20Balance = await btcErc20Token.balanceOf(senderAddress);
      const currentSenderNativeBalance = await ethers.provider.getBalance(senderAddress);
      expect(currentSenderNativeBalance).to.equal(currentSenderBTCERC20Balance);

      const currentRecipientBTCERC20Balance = await btcErc20Token.balanceOf(recipientAddress);
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
      expect(currentRecipientNativeBalance).to.equal(currentRecipientBTCERC20Balance);
    });

    it("should verify BTCTransfers contract has full balance", async function () {
      const btcTransfersAddress = await btcTransfers.getAddress();
      const currentContractNativeBalance = await ethers.provider.getBalance(btcTransfersAddress);
      expect(currentContractNativeBalance).to.equal(tokenAmount);
    });
  });

  describe("basicSendThenSpendThenSendAndReturn", function () {
    let initialSenderBalance: any;
    let initialRecipientBalance: any;
    let tokenAmount: any;
    let gasCost: any;

    before(async function () {
      await fixture();

      tokenAmount = ethers.parseEther("8");
      const btcTransfersAddress = await btcTransfers.getAddress();

      const transferTx = await btcErc20Token.connect(senderSigner).transfer(btcTransfersAddress, tokenAmount);
      await transferTx.wait();

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(otherSpender);


      const tx = await btcTransfers.connect(senderSigner).basicSendThenSpendThenSendAndReturn(otherSpender);
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * receipt.gasPrice;
    });

    it("should verify initial recipient balance is zero", async function () {
      expect(initialRecipientBalance).to.equal(0);
    });

    it("should verify current recipient balance", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(otherSpender);
      expect(currentRecipientNativeBalance).to.equal(tokenAmount / 4n);
    });

    it("should verify native and BTC ERC20 balance equivalence", async function () {
      const currentSenderBTCERC20Balance = await btcErc20Token.balanceOf(senderAddress);
      const currentSenderNativeBalance = await ethers.provider.getBalance(senderAddress);
      expect(currentSenderNativeBalance).to.equal(currentSenderBTCERC20Balance);

      const currentRecipientBTCERC20Balance = await btcErc20Token.balanceOf(recipientAddress);
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
      expect(currentRecipientNativeBalance).to.equal(currentRecipientBTCERC20Balance);
    });

    it("should verify BTCTransfers contract has 3/4 of balance", async function () {
      const btcTransfersAddress = await btcTransfers.getAddress();
      const currentContractNativeBalance = await ethers.provider.getBalance(btcTransfersAddress);
      expect(currentContractNativeBalance).to.equal(tokenAmount / 4n * 3n);
    });
  });

  describe("nativeThenERC20", function () {
    let initialSenderBalance: any;
    let initialRecipientBalance: any;
    let tokenAmount: any;
    let gasCost: any;

    before(async function () {
      await fixture();

      tokenAmount = ethers.parseEther("8");
      const btcTransfersAddress = await btcTransfers.getAddress();

      const transferTx = await btcErc20Token.connect(senderSigner).transfer(btcTransfersAddress, tokenAmount);
      await transferTx.wait();

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await btcTransfers.connect(senderSigner).nativeThenERC20(recipientAddress);
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * receipt.gasPrice;
    });

    it("should verify sender native balance", async function () {
      const currentSenderNativeBalance = await ethers.provider.getBalance(senderAddress);
      expect(initialSenderBalance - gasCost).to.equal(currentSenderNativeBalance);
    });

    it("should verify initial recipient balance is zero", async function () {
      expect(initialRecipientBalance).to.equal(0);
    });

    it("should verify current recipient balance", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
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

    it("should verify BTCTransfers contract has zero balance", async function () {
      const btcTransfersAddress = await btcTransfers.getAddress();
      const currentContractNativeBalance = await ethers.provider.getBalance(btcTransfersAddress);
      expect(currentContractNativeBalance).to.equal(0);
    });
  });

  describe("erc20ThenNative", function () {
    let initialSenderBalance: any;
    let initialRecipientBalance: any;
    let tokenAmount: any;
    let gasCost: any;

    before(async function () {
      await fixture();

      tokenAmount = ethers.parseEther("12");
      const btcTransfersAddress = await btcTransfers.getAddress();

      const transferTx = await btcErc20Token.connect(senderSigner).transfer(btcTransfersAddress, tokenAmount);
      await transferTx.wait();

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await btcTransfers.connect(senderSigner).erc20ThenNative(recipientAddress);
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * receipt.gasPrice;
    });

    it("should verify sender native balance", async function () {
      const currentSenderNativeBalance = await ethers.provider.getBalance(senderAddress);
      expect(initialSenderBalance - gasCost).to.equal(currentSenderNativeBalance);
    });

    it("should verify initial recipient balance is zero", async function () {
      expect(initialRecipientBalance).to.equal(0);
    });

    it("should verify current recipient native balance", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
      expect(currentRecipientNativeBalance).to.equal(tokenAmount);
    });

    it("should verify BTCTransfers contract has zero balance", async function () {
      const btcTransfersAddress = await btcTransfers.getAddress();
      const currentContractNativeBalance = await ethers.provider.getBalance(btcTransfersAddress);
      expect(currentContractNativeBalance).to.equal(0);
    });
  });

  describe("receiveSendNative", function () {
    let initialSenderBalance: any;
    let initialRecipientBalance: any;
    let nativeAmount: any;
    let gasCost: any;

    before(async function () {
      await fixture();

      nativeAmount = ethers.parseEther("10");

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await btcTransfers.connect(senderSigner).receiveSendNative(recipientAddress, { value: nativeAmount });
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * receipt.gasPrice;
    });

    it("should verify sender native balance", async function () {
      const currentSenderNativeBalance = await ethers.provider.getBalance(senderAddress);
      expect(initialSenderBalance - nativeAmount - gasCost).to.equal(currentSenderNativeBalance);
    });

    it("should verify initial recipient balance is zero", async function () {
      expect(initialRecipientBalance).to.equal(0);
    });

    it("should verify recipient native balance", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
      expect(currentRecipientNativeBalance).to.equal(nativeAmount);
    });

    it("should verify BTCTransfers contract has zero balance", async function () {
      const btcTransfersAddress = await btcTransfers.getAddress();
      const currentContractNativeBalance = await ethers.provider.getBalance(btcTransfersAddress);
      expect(currentContractNativeBalance).to.equal(0);
    });
  });

  describe("receiveSendERC20", function () {
    let initialSenderBalance: any;
    let initialRecipientBalance: any;
    let nativeAmount: any;
    let gasCost: any;
    let btcTransfersAddress: any;

    before(async function () {
      await fixture();

      nativeAmount = ethers.parseEther("11");
      btcTransfersAddress = await btcTransfers.getAddress();

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await btcTransfers.connect(senderSigner).receiveSendERC20(recipientAddress, { value: nativeAmount });
      const receipt = await tx.wait();

      gasCost = receipt.gasUsed * receipt.gasPrice;
    });

    it("should verify sender native balance", async function () {
      const currentSenderNativeBalance = await ethers.provider.getBalance(senderAddress);
      expect(initialSenderBalance - nativeAmount - gasCost).to.equal(currentSenderNativeBalance);
    });

    it("should verify initial recipient balance is zero", async function () {
      expect(initialRecipientBalance).to.equal(0);
    });

    it("should verify recipient native balance", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
      expect(currentRecipientNativeBalance).to.equal(nativeAmount);
    });

    it("should verify BTCTransfers contract has zero balance", async function () {
      const currentContractBalance = await ethers.provider.getBalance(btcTransfersAddress);
      expect(currentContractBalance).to.equal(0);
    });
  });

  describe("receiveSendNativeThenERC20", function () {
    let initialSenderBalance: any;
    let initialRecipientBalance: any;
    let gasCost: any;
    let btcTransfersAddress: any;
    const nativeAmount = ethers.parseEther("3");

    before(async function () {
      await fixture();

      btcTransfersAddress = await btcTransfers.getAddress();

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await btcTransfers.connect(senderSigner).receiveSendNativeThenERC20(recipientAddress, { value: nativeAmount });
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * receipt.gasPrice;
    });

    it("should verify sender native balance deduction", async function () {
      const currentSenderNativeBalance = await ethers.provider.getBalance(senderAddress);
      expect(initialSenderBalance - nativeAmount - gasCost).to.equal(currentSenderNativeBalance);
    });

    it("should verify recipient native balance received", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
      expect(initialRecipientBalance + nativeAmount).to.equal(currentRecipientNativeBalance);
    });

    it("should verify BTCTransfers contract has zero balance", async function () {
      const currentBalance = await ethers.provider.getBalance(btcTransfersAddress);
      expect(currentBalance).to.equal(0);
    });
  });

  describe("receiveSendERC20ThenNative", function () {
    const nativeAmount = ethers.parseEther("5");
    let initialSenderBalance: any;
    let initialRecipientBalance: any;
    let gasCost: any;
    let btcTransfersAddress: any;

    before(async function () {
      await fixture();

      btcTransfersAddress = await btcTransfers.getAddress();

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await btcTransfers.connect(senderSigner).receiveSendERC20ThenNative(recipientAddress, { value: nativeAmount });
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * receipt.gasPrice;
    });

    it("should verify sender native balance deduction", async function () {
      const currentSenderNativeBalance = await ethers.provider.getBalance(senderAddress);
      expect(initialSenderBalance - nativeAmount - gasCost).to.equal(currentSenderNativeBalance);
    });

    it("should verify recipient native balance received", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
      expect(initialRecipientBalance + nativeAmount).to.equal(currentRecipientNativeBalance);
    });

    it("should verify BTCTransfers contract has zero balance", async function () {
      const currentBalance = await ethers.provider.getBalance(btcTransfersAddress);
      expect(currentBalance).to.equal(0);
    });
  });

  describe("receiveSendMultiple", function () {
    let initialSenderBalance: any;
    let initialRecipientBalance: any;
    const nativeAmount = ethers.parseEther("8");
    let gasCost: any;
    let btcTransfersAddress: any;

    before(async function () {
      await fixture();

      btcTransfersAddress = await btcTransfers.getAddress();

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await btcTransfers.connect(senderSigner).receiveSendMultiple(recipientAddress, { value: nativeAmount });
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * receipt.gasPrice;
    });

    it("should verify sender native balance deduction", async function () {
      const currentSenderBalance = await ethers.provider.getBalance(senderAddress);
      expect(initialSenderBalance - nativeAmount - gasCost).to.equal(currentSenderBalance);
    });

    it("should verify recipient native balance received", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
      expect(initialRecipientBalance + nativeAmount).to.equal(currentRecipientNativeBalance);
    });

    it("should verify BTCTransfers contract has zero balance", async function () {
      const currentBalance = await ethers.provider.getBalance(btcTransfersAddress);
      expect(currentBalance).to.equal(0);
    });
  });

  describe("receiveSendMultiple_tinyAmount", function () {
    let initialSenderBalance: any;
    let initialRecipientBalance: any;
    const nativeAmount = 4n; // this amount is split by 4 in contract
    let gasCost: any;
    let btcTransfersAddress: any;

    before(async function () {
      await fixture();

      btcTransfersAddress = await btcTransfers.getAddress();

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      // Transfer
      const tx = await btcTransfers.connect(senderSigner).receiveSendMultiple(recipientAddress, { value: nativeAmount });
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * receipt.gasPrice;
    });

    it("should verify sender native balance deduction", async function () {
      const currentSenderBalance = await ethers.provider.getBalance(senderAddress);
      expect(initialSenderBalance - nativeAmount - gasCost).to.equal(currentSenderBalance);
    });

    it("should verify recipient native balance received", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
      expect(initialRecipientBalance + nativeAmount).to.equal(currentRecipientNativeBalance);
    });

    it("should verify BTCTransfers contract has zero balance", async function () {
      const currentBalance = await ethers.provider.getBalance(btcTransfersAddress);
      expect(currentBalance).to.equal(0);
    });
  });

  describe("receiveSendMultiple_hugeAmount", function () {
    const nativeAmount = ethers.parseEther("21000000"); // this amount is split by 4 in contract
    let initialSenderBalance: any;
    let initialRecipientBalance: any;
    let gasCost: any;
    let btcTransfersAddress: any;

    before(async function () {
      await fixture();

      btcTransfersAddress = await btcTransfers.getAddress();

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await btcTransfers.connect(senderSigner).receiveSendMultiple(recipientAddress, { value: nativeAmount });
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * receipt.gasPrice;
    });

    it("should verify sender native balance deduction", async function () {
      const currentSenderBalance = await ethers.provider.getBalance(senderAddress);
      expect(initialSenderBalance - nativeAmount - gasCost).to.equal(currentSenderBalance);
    });

    it("should verify recipient native balance received", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
      expect(initialRecipientBalance + nativeAmount).to.equal(currentRecipientNativeBalance);
    });

    it("should verify BTCTransfers contract has zero balance", async function () {
      const currentBalance = await ethers.provider.getBalance(btcTransfersAddress);
      expect(currentBalance).to.equal(0);
    });
  });

  describe("receiveSendRevert", function () {
    let initialSenderBalance: any;
    let initialRecipientBalance: any;
    const nativeAmount = 42;
    let btcTransfersAddress: any;

    before(async function () {
      await fixture();

      btcTransfersAddress = await btcTransfers.getAddress();

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      try {
        // Transfer
        await btcTransfers
          .connect(senderSigner)
          .receiveSendRevert(recipientAddress, { value: nativeAmount });
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

    it("should verify BTCTransfers contract has zero balance", async function () {
      const currentBalance = await ethers.provider.getBalance(btcTransfersAddress);
      expect(currentBalance).to.equal(0);
    });
  });

  describe("receiveSendNativeThenPullERC20", function () {
    let initialSenderBalance: any;
    let initialRecipientBalance: any;
    let gasCost: any;
    let btcTransfersAddress: any;
    const nativeAmount = ethers.parseEther("3");
    const tokenAmount = ethers.parseEther("1");

    before(async function () {
      await fixture();

      btcTransfersAddress = await btcTransfers.getAddress();

      await btcErc20Token.connect(senderSigner).approve(btcTransfersAddress, tokenAmount)
        .then(tx => tx.wait());

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await btcTransfers.connect(senderSigner).receiveSendNativeThenPullERC20(recipientAddress, tokenAmount, { value: nativeAmount });
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * receipt.gasPrice;
    });

    it("should verify sender native balance deduction", async function () {
      const currentSenderNativeBalance = await ethers.provider.getBalance(senderAddress);
      expect(initialSenderBalance - tokenAmount - nativeAmount - gasCost).to.equal(currentSenderNativeBalance);
    });

    it("should verify recipient native balance received", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
      expect(initialRecipientBalance + tokenAmount + nativeAmount).to.equal(currentRecipientNativeBalance);
    });

    it("should verify BTCTransfers contract has zero balance", async function () {
      const currentBalance = await ethers.provider.getBalance(btcTransfersAddress);
      expect(currentBalance).to.equal(0);
    });
  });

  describe("receivePullERC20ThenNative", function () {
    let initialSenderBalance: any;
    let initialRecipientBalance: any;
    const tokenAmount = ethers.parseEther("4");
    const nativeAmount = ethers.parseEther("5");
    let gasCost: any;
    let btcTransfersAddress: any;

    before(async function () {
      await fixture();

      btcTransfersAddress = await btcTransfers.getAddress();

      await btcErc20Token.connect(senderSigner).approve(btcTransfersAddress, tokenAmount)
        .then(tx => tx.wait());

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await btcTransfers.connect(senderSigner).receivePullERC20ThenNative(recipientAddress, tokenAmount, { value: nativeAmount });
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * receipt.gasPrice;
    });

    it("should verify sender native balance deduction", async function () {
      const currentSenderNativeBalance = await ethers.provider.getBalance(senderAddress);
      expect(initialSenderBalance - nativeAmount- tokenAmount - gasCost).to.equal(currentSenderNativeBalance);
    });

    it("should verify recipient native balance received", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
      expect(initialRecipientBalance + nativeAmount + tokenAmount).to.equal(currentRecipientNativeBalance);
    });

    it("should verify BTCTransfers contract has zero balance", async function () {
      const currentBalance = await ethers.provider.getBalance(btcTransfersAddress);
      expect(currentBalance).to.equal(0);
    });
  });

  describe("stateChangeThenPullERC20", function () {
    let initialContractBalance: any;
    let initialSenderBalance: any;
    let amount: any;
    let gasCost: any;
    let btcTransfersAddress: any;

    before(async function () {
      await fixture();

      amount = ethers.parseEther("2");
      btcTransfersAddress = await btcTransfers.getAddress();

      await btcErc20Token.connect(senderSigner).approve(btcTransfersAddress, amount)
        .then(tx => tx.wait());

      initialContractBalance = await ethers.provider.getBalance(btcTransfersAddress);
      initialSenderBalance = await ethers.provider.getBalance(senderAddress);

      const tx = await btcTransfers.connect(senderSigner).stateChangeThenPullERC20(amount, true);
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * receipt.gasPrice;
    });

    it("should verify sender balance deduction", async function () {
      const expectedBalance = initialSenderBalance - amount - gasCost;

      // Just in case, verify native and ERC-20 balances equivalence.
      const currentSenderNativeBalance = await ethers.provider.getBalance(senderAddress);
      const currentSenderERC20Balance = await btcErc20Token.balanceOf(senderAddress);

      expect(currentSenderNativeBalance).to.equal(expectedBalance);
      expect(currentSenderERC20Balance).to.equal(expectedBalance);
    });

    it("should verify BTCTransfers contract has zero balance before the call", async function () {
      expect(initialContractBalance).to.equal(0);
    });

    it("should verify BTCTransfers contract balance increased after the call", async function () {
      // Just in case, verify native and ERC-20 balances equivalence.
      const currentContractNativeBalance = await ethers.provider.getBalance(btcTransfersAddress);
      const currentContractERC20Balance = await btcErc20Token.balanceOf(btcTransfersAddress);

      expect(currentContractNativeBalance).to.equal(amount);
      expect(currentContractERC20Balance).to.equal(amount);
    });

    it("should verify balanceTracker storage variable was unchanged", async function () {
      const balanceTracker = await btcTransfers.balanceTracker();
      expect(balanceTracker).to.equal(1); // 1 is the original value of the storage variable.
    });
  });

  describe("stateChangeThenRevertingPullERC20", function () {
    let initialContractBalance: any;
    let initialSenderBalance: any;
    let amount: any;
    let btcTransfersAddress: any;
    let tx: any;
    let receipt: any;

    before(async function () {
      await fixture();

      amount = ethers.parseEther("2");
      btcTransfersAddress = await btcTransfers.getAddress();

      // Deliberately omit the call where sender approves the BTCTransfers contract to pull ERC-20 BTC.
      // This will cause the IBTC(precompile).transferFrom call to revert.

      initialContractBalance = await ethers.provider.getBalance(btcTransfersAddress);
      initialSenderBalance = await ethers.provider.getBalance(senderAddress);

      try {
        // The second argument is false, so the storage variable is not reset by the contractlogic.
        // This is needed to ensure the only way to reset the storage variable is through a tx revert.
        tx = await btcTransfers.connect(senderSigner).stateChangeThenPullERC20(amount, false, {gasLimit: 1000000});
        await tx.wait();
      } catch (error) {
        expect((error as Error).message).to.contain("execution reverted");
      }

      receipt = await ethers.provider.getTransactionReceipt(tx.hash);
    });

    it("should verify the transaction reverted", async function () {
      expect(receipt!.status).to.equal(0);
    });

    it("should verify sender balance decreased by gas cost only", async function () {
      const gasCost = receipt!.gasUsed * receipt!.gasPrice;
      const expectedBalance = initialSenderBalance - gasCost;

      // Just in case, verify native and ERC-20 balances equivalence.
      const currentSenderNativeBalance = await ethers.provider.getBalance(senderAddress);
      const currentSenderERC20Balance = await btcErc20Token.balanceOf(senderAddress);

      expect(currentSenderNativeBalance).to.equal(expectedBalance);
      expect(currentSenderERC20Balance).to.equal(expectedBalance);
    });

    it("should verify BTCTransfers contract has zero balance before the call", async function () {
      expect(initialContractBalance).to.equal(0);
    });

    it("should verify BTCTransfers contract has zero balance after the call", async function () {
      // Just in case, verify native and ERC-20 balances equivalence.
      const currentContractNativeBalance = await ethers.provider.getBalance(btcTransfersAddress);
      const currentContractERC20Balance = await btcErc20Token.balanceOf(btcTransfersAddress);

      expect(currentContractNativeBalance).to.equal(0);
      expect(currentContractERC20Balance).to.equal(0);
    });

    it("should verify balanceTracker storage variable was unchanged", async function () {
      const balanceTracker = await btcTransfers.balanceTracker();
      expect(balanceTracker).to.equal(1); // 1 is the original value of the storage variable.
    });
  });

  describe("erc20RevertsWhenExceedMaxPrecompileCalls", function () {
    let initialRecipientBalance: any;
    let tokenAmount: any;

    before(async function () {
      await fixture();

      tokenAmount = ethers.parseEther("8");
      const btcTransfersAddress = await btcTransfers.getAddress();

      const transferTx = await btcErc20Token.connect(senderSigner).transfer(btcTransfersAddress, tokenAmount);
      await transferTx.wait();

      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      try {
        const tx = await btcTransfers.connect(senderSigner).erc20RevertsWhenExceedMaxPrecompileCalls(recipientAddress, {gasLimit: 10000000});
        await tx.wait();
      } catch (err) {
      }
    });

    it("should verify initial recipient balance is zero", async function () {
      expect(initialRecipientBalance).to.equal(0);
    });

    it("should verify current recipient balance is zero", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
      const currentRecipientERC20Balance = await btcErc20Token.balanceOf(recipientAddress);
      expect(currentRecipientNativeBalance).to.equal(0n);
      expect(currentRecipientERC20Balance).to.equal(0n);
    });

    it("should verify BTCTransfers contract has full balance", async function () {
      const btcTransfersAddress = await btcTransfers.getAddress();
      const currentContractNativeBalance = await ethers.provider.getBalance(btcTransfersAddress);
      expect(currentContractNativeBalance).to.equal(tokenAmount);
    });
  });


  describe("erc20ThenRevertingExternalCallWithMultiplePrecompile", function () {
    let initialRecipientBalance: any;
    let tokenAmount: any;

    before(async function () {
      await fixture();

      tokenAmount = ethers.parseEther("8");
      const btcTransfersAddress = await btcTransfers.getAddress();

      const transferTx = await btcErc20Token.connect(senderSigner).transfer(btcTransfersAddress, tokenAmount);
      await transferTx.wait();

      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await btcTransfers.connect(senderSigner).erc20ThenRevertingExternalCallWithMultiplePrecompile(recipientAddress);
      await tx.wait();
    });

    it("should verify initial recipient balance is zero", async function () {
      expect(initialRecipientBalance).to.equal(0);
    });

    it("should verify current recipient balance", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
      expect(currentRecipientNativeBalance).to.equal(tokenAmount / 2n);
    });

    it("should verify BTCTransfers contract has zero balance", async function () {
      const btcTransfersAddress = await btcTransfers.getAddress();
      const currentContractNativeBalance = await ethers.provider.getBalance(btcTransfersAddress);
      expect(currentContractNativeBalance).to.equal(0);
    });
  });

  describe("revertingExternalCallInPrecompileThenERC20", function () {
    let initialRecipientBalance: any;
    let tokenAmount: any;

    before(async function () {
      await fixture();

      tokenAmount = ethers.parseEther("8");
      const btcTransfersAddress = await btcTransfers.getAddress();

      const transferTx = await btcErc20Token.connect(senderSigner).transfer(btcTransfersAddress, tokenAmount);
      await transferTx.wait();

      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await btcTransfers.connect(senderSigner).revertingExternalCallInPrecompileThenERC20(recipientAddress);
      await tx.wait();
    });

    it("should verify initial recipient balance is zero", async function () {
      expect(initialRecipientBalance).to.equal(0);
    });

    it("should verify current recipient balance", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
      expect(currentRecipientNativeBalance).to.equal(tokenAmount / 2n);
    });

    it("should verify BTCTransfers contract has zero balance", async function () {
      const btcTransfersAddress = await btcTransfers.getAddress();
      const currentContractNativeBalance = await ethers.provider.getBalance(btcTransfersAddress);
      expect(currentContractNativeBalance).to.equal(0);
    });
  });

  describe("revertingExternalCallThenERC20Transfer", function () {
    let initialRecipientBalance: any;
    let tokenAmount: any;

    before(async function () {
      await fixture();

      tokenAmount = ethers.parseEther("8");
      const btcTransfersAddress = await btcTransfers.getAddress();

      const transferTx = await btcErc20Token.connect(senderSigner).transfer(btcTransfersAddress, tokenAmount);
      await transferTx.wait();

      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await btcTransfers.connect(senderSigner).revertingExternalCallThenERC20Transfer(recipientAddress);
      await tx.wait();
    });

    it("should verify initial recipient balance is zero", async function () {
      expect(initialRecipientBalance).to.equal(0);
    });

    it("should verify current recipient balance", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
      expect(currentRecipientNativeBalance).to.equal(tokenAmount / 2n);
    });

    it("should verify BTCTransfers contract has zero balance", async function () {
      const btcTransfersAddress = await btcTransfers.getAddress();
      const currentContractNativeBalance = await ethers.provider.getBalance(btcTransfersAddress);
      expect(currentContractNativeBalance).to.equal(0);
    });
  });

  describe("erc20ThenRevertingExternalCallInPrecompile", function () {
    let initialRecipientBalance: any;
    let tokenAmount: any;

    before(async function () {
      await fixture();

      tokenAmount = ethers.parseEther("8");
      const btcTransfersAddress = await btcTransfers.getAddress();

      const transferTx = await btcErc20Token.connect(senderSigner).transfer(btcTransfersAddress, tokenAmount);
      await transferTx.wait();

      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await btcTransfers.connect(senderSigner).erc20ThenRevertingExternalCallInPrecompile(recipientAddress);
      await tx.wait();
    });

    it("should verify initial recipient balance is zero", async function () {
      expect(initialRecipientBalance).to.equal(0);
    });

    it("should verify current recipient balance", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
      expect(currentRecipientNativeBalance).to.equal(tokenAmount / 2n);
    });

    it("should verify BTCTransfers contract has zero balance", async function () {
      const btcTransfersAddress = await btcTransfers.getAddress();
      const currentContractNativeBalance = await ethers.provider.getBalance(btcTransfersAddress);
      expect(currentContractNativeBalance).to.equal(0);
    });
  });

  describe("revertingInPrecompile", function () {
    let initialRecipientBalance: any;
    let tokenAmount: any;

    before(async function () {
      await fixture();

      tokenAmount = ethers.parseEther("8");
      const btcTransfersAddress = await btcTransfers.getAddress();

      const transferTx = await btcErc20Token.connect(senderSigner).transfer(btcTransfersAddress, tokenAmount);
      await transferTx.wait();

      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      try {
        const tx = await btcTransfers.connect(senderSigner).revertingInPrecompile(recipientAddress, {gasLimit: 10000000});
        await tx.wait();
      } catch (err) {}
    });

    it("should verify initial recipient balance is zero", async function () {
      expect(initialRecipientBalance).to.equal(0);
    });

    it("should verify current recipient balance is zero", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
      const currentRecipientERC20Balance = await btcErc20Token.balanceOf(recipientAddress);

      expect(currentRecipientNativeBalance).to.equal(0n);
      expect(currentRecipientERC20Balance).to.equal(0n);
    });

    it("should verify BTCTransfers contract has full balance", async function () {
      const btcTransfersAddress = await btcTransfers.getAddress();

      const currentContractNativeBalance = await ethers.provider.getBalance(btcTransfersAddress);
      const currentContractERC20Balance = await btcErc20Token.balanceOf(btcTransfersAddress);

      expect(currentContractNativeBalance).to.equal(tokenAmount);
      expect(currentContractERC20Balance).to.equal(tokenAmount);
    });
  });

  describe("erc20ThenRevertingInPrecompile", function () {
    let initialRecipientBalance: any;
    let tokenAmount: any;

    before(async function () {
      await fixture();

      tokenAmount = ethers.parseEther("8");
      const btcTransfersAddress = await btcTransfers.getAddress();

      const transferTx = await btcErc20Token.connect(senderSigner).transfer(btcTransfersAddress, tokenAmount);
      await transferTx.wait();

      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      try {
        const tx = await btcTransfers.connect(senderSigner).erc20ThenRevertingInPrecompile(recipientAddress, {gasLimit: 10000000});
        await tx.wait();
      } catch (err) {}
    });

    it("should verify initial recipient balance is zero", async function () {
      expect(initialRecipientBalance).to.equal(0);
    });

    it("should verify current recipient balance is zero", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
      const currentRecipientERC20Balance = await btcErc20Token.balanceOf(recipientAddress);
      expect(currentRecipientNativeBalance).to.equal(0n);
      expect(currentRecipientERC20Balance).to.equal(0n);
    });

    it("should verify BTCTransfers contract has full balance", async function () {
      const btcTransfersAddress = await btcTransfers.getAddress();
      const currentContractNativeBalance = await ethers.provider.getBalance(btcTransfersAddress);
      expect(currentContractNativeBalance).to.equal(tokenAmount);
    });
  });

  describe("erc20ThenRevertingExternalCall", function () {
    let initialRecipientBalance: any;
    let tokenAmount: any;

    before(async function () {
      await fixture();

      tokenAmount = ethers.parseEther("8");
      const btcTransfersAddress = await btcTransfers.getAddress();

      const transferTx = await btcErc20Token.connect(senderSigner).transfer(btcTransfersAddress, tokenAmount);
      await transferTx.wait();
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await btcTransfers.connect(senderSigner).erc20ThenRevertingExternalCall(recipientAddress);
      await tx.wait();
    });

    it("should verify initial recipient balance is zero", async function () {
      expect(initialRecipientBalance).to.equal(0);
    });

    it("should verify current recipient balance", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
      expect(currentRecipientNativeBalance).to.equal(tokenAmount / 2n);
    });

    it("should verify BTCTransfers contract has zero balance", async function () {
      const btcTransfersAddress = await btcTransfers.getAddress();
      const currentContractNativeBalance = await ethers.provider.getBalance(btcTransfersAddress);
      expect(currentContractNativeBalance).to.equal(0);
    });
  });
});
