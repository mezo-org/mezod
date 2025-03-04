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
  let senderSigner: any;
  let senderAddress: string;
  let recipientAddress: string;

  const fixture = (async function () {
    await deployments.fixture();
    btcErc20Token = new hre.ethers.Contract(precompileAddress, abi, ethers.provider);
    mezoTransfers = await getDeployedContract("MezoTransfers");
    signers = await ethers.getSigners();
    senderSigner = signers[0];
    senderAddress = senderSigner.address;
    recipientAddress = ethers.Wallet.createRandom().address;
  });

  describe("erc20WithTestbedPrecompileTransfer", function () {
    let initialRecipientBalance: any;
    let tokenAmount: any;

    before(async function () {
      await fixture();

      tokenAmount = ethers.parseEther("12");
      const mezoTransfersAddress = await mezoTransfers.getAddress();

      const transferTx = await btcErc20Token.connect(senderSigner).transfer(mezoTransfersAddress, tokenAmount);
      await transferTx.wait();

      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await mezoTransfers.connect(senderSigner).erc20WithTestbedPrecompileTransfer(recipientAddress);
      await tx.wait();
    });

    it("should verify initial recipient balance is zero", async function () {
      expect(initialRecipientBalance).to.equal(0);
    });

    it("should verify current recipient native balance", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
      expect(currentRecipientNativeBalance).to.equal(tokenAmount / 2n);
    });

    it("should verify MezoTransfers contract has half balance", async function () {
      const mezoTransfersAddress = await mezoTransfers.getAddress();
      const currentContractNativeBalance = await ethers.provider.getBalance(mezoTransfersAddress);
      expect(currentContractNativeBalance).to.equal(tokenAmount / 2n);
    });
  });


  describe("erc20ThenRevertingExternalCallWithMultiplePrecompile", function () {
    let initialRecipientBalance: any;
    let tokenAmount: any;

    before(async function () {
      await fixture();

      tokenAmount = ethers.parseEther("8");
      const mezoTransfersAddress = await mezoTransfers.getAddress();

      const transferTx = await btcErc20Token.connect(senderSigner).transfer(mezoTransfersAddress, tokenAmount);
      await transferTx.wait();

      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await mezoTransfers.connect(senderSigner).erc20ThenRevertingExternalCallWithMultiplePrecompile(recipientAddress);
      await tx.wait();
    });

    it("should verify initial recipient balance is zero", async function () {
      expect(initialRecipientBalance).to.equal(0);
    });

    it("should verify current recipient balance", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
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

    it("should verify MezoTransfers contract has zero balance", async function () {
      const mezoTransfersAddress = await mezoTransfers.getAddress();
      const currentContractNativeBalance = await ethers.provider.getBalance(mezoTransfersAddress);
      expect(currentContractNativeBalance).to.equal(0);
    });
  });

  describe("revertingExternalCallInPrecompileThenERC20", function () {
    let initialRecipientBalance: any;
    let tokenAmount: any;

    before(async function () {
      await fixture();

      tokenAmount = ethers.parseEther("8");
      const mezoTransfersAddress = await mezoTransfers.getAddress();

      const transferTx = await btcErc20Token.connect(senderSigner).transfer(mezoTransfersAddress, tokenAmount);
      await transferTx.wait();

      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await mezoTransfers.connect(senderSigner).revertingExternalCallInPrecompileThenERC20(recipientAddress);
      await tx.wait();
    });

    it("should verify initial recipient balance is zero", async function () {
      expect(initialRecipientBalance).to.equal(0);
    });

    it("should verify current recipient balance", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
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

    it("should verify MezoTransfers contract has zero balance", async function () {
      const mezoTransfersAddress = await mezoTransfers.getAddress();
      const currentContractNativeBalance = await ethers.provider.getBalance(mezoTransfersAddress);
      expect(currentContractNativeBalance).to.equal(0);
    });
  });

  describe("revertingExternalCallThenERC20Transfer", function () {
    let initialRecipientBalance: any;
    let tokenAmount: any;

    before(async function () {
      await fixture();

      tokenAmount = ethers.parseEther("8");
      const mezoTransfersAddress = await mezoTransfers.getAddress();

      const transferTx = await btcErc20Token.connect(senderSigner).transfer(mezoTransfersAddress, tokenAmount);
      await transferTx.wait();

      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await mezoTransfers.connect(senderSigner).revertingExternalCallThenERC20Transfer(recipientAddress);
      await tx.wait();
    });


    it("should verify initial recipient balance is zero", async function () {
      expect(initialRecipientBalance).to.equal(0);
    });

    it("should verify current recipient balance", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
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

    it("should verify MezoTransfers contract has zero balance", async function () {
      const mezoTransfersAddress = await mezoTransfers.getAddress();
      const currentContractNativeBalance = await ethers.provider.getBalance(mezoTransfersAddress);
      expect(currentContractNativeBalance).to.equal(0);
    });
  });

  describe("erc20ThenRevertingExternalCallInPrecompile", function () {
    let initialRecipientBalance: any;
    let tokenAmount: any;

    before(async function () {
      await fixture();

      tokenAmount = ethers.parseEther("8");
      const mezoTransfersAddress = await mezoTransfers.getAddress();

      const transferTx = await btcErc20Token.connect(senderSigner).transfer(mezoTransfersAddress, tokenAmount);
      await transferTx.wait();

      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await mezoTransfers.connect(senderSigner).erc20ThenRevertingExternalCallInPrecompile(recipientAddress);
      await tx.wait();
    });


    it("should verify initial recipient balance is zero", async function () {
      expect(initialRecipientBalance).to.equal(0);
    });

    it("should verify current recipient balance", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
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

    it("should verify MezoTransfers contract has zero balance", async function () {
      const mezoTransfersAddress = await mezoTransfers.getAddress();
      const currentContractNativeBalance = await ethers.provider.getBalance(mezoTransfersAddress);
      expect(currentContractNativeBalance).to.equal(0);
    });
  });

  describe("erc20ThenRevertingInPrecompile", function () {
    let initialRecipientBalance: any;
    let tokenAmount: any;

    before(async function () {
      await fixture();

      tokenAmount = ethers.parseEther("8");
      const mezoTransfersAddress = await mezoTransfers.getAddress();

      const transferTx = await btcErc20Token.connect(senderSigner).transfer(mezoTransfersAddress, tokenAmount);
      await transferTx.wait();

      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      try {
        const tx = await mezoTransfers.connect(senderSigner).erc20ThenRevertingInPrecompile(recipientAddress, {gasLimit: 10000000});
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

    it("should verify native and BTC ERC20 balance equivalence", async function () {
      const currentSenderBTCERC20Balance = await btcErc20Token.balanceOf(senderAddress);
      const currentSenderNativeBalance = await ethers.provider.getBalance(senderAddress);
      expect(currentSenderNativeBalance).to.equal(currentSenderBTCERC20Balance);

      const currentRecipientBTCERC20Balance = await btcErc20Token.balanceOf(recipientAddress);
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
      expect(currentRecipientNativeBalance).to.equal(currentRecipientBTCERC20Balance);
    });

    it("should verify MezoTransfers contract has full balance", async function () {
      const mezoTransfersAddress = await mezoTransfers.getAddress();
      const currentContractNativeBalance = await ethers.provider.getBalance(mezoTransfersAddress);
      expect(currentContractNativeBalance).to.equal(tokenAmount);
    });
  });

  describe("erc20ThenRevertingExternalCall", function () {
    let initialRecipientBalance: any;
    let tokenAmount: any;

    before(async function () {
      await fixture();

      tokenAmount = ethers.parseEther("8");
      const mezoTransfersAddress = await mezoTransfers.getAddress();

      const transferTx = await btcErc20Token.connect(senderSigner).transfer(mezoTransfersAddress, tokenAmount);
      await transferTx.wait();
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await mezoTransfers.connect(senderSigner).erc20ThenRevertingExternalCall(recipientAddress);
      await tx.wait();
    });


    it("should verify initial recipient balance is zero", async function () {
      expect(initialRecipientBalance).to.equal(0);
    });

    it("should verify current recipient balance", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
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

    it("should verify MezoTransfers contract has zero balance", async function () {
      const mezoTransfersAddress = await mezoTransfers.getAddress();
      const currentContractNativeBalance = await ethers.provider.getBalance(mezoTransfersAddress);
      expect(currentContractNativeBalance).to.equal(0);
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
      const mezoTransfersAddress = await mezoTransfers.getAddress();

      const transferTx = await btcErc20Token.connect(senderSigner).transfer(mezoTransfersAddress, tokenAmount);
      await transferTx.wait();

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await mezoTransfers.connect(senderSigner).nativeThenERC20(recipientAddress);
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

    it("should verify MezoTransfers contract has zero balance", async function () {
      const mezoTransfersAddress = await mezoTransfers.getAddress();
      const currentContractNativeBalance = await ethers.provider.getBalance(mezoTransfersAddress);
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
      const mezoTransfersAddress = await mezoTransfers.getAddress();

      const transferTx = await btcErc20Token.connect(senderSigner).transfer(mezoTransfersAddress, tokenAmount);
      await transferTx.wait();

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await mezoTransfers.connect(senderSigner).erc20ThenNative(recipientAddress);
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

    before(async function () {
      await fixture();

      nativeAmount = ethers.parseEther("10");

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await mezoTransfers.connect(senderSigner).receiveSendNative(recipientAddress, { value: nativeAmount });
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

    it("should verify MezoTransfers contract has zero balance", async function () {
      const mezoTransfersAddress = await mezoTransfers.getAddress();
      const currentContractNativeBalance = await ethers.provider.getBalance(mezoTransfersAddress);
      expect(currentContractNativeBalance).to.equal(0);
    });
  });

  describe("receiveSendERC20", function () {
    let initialSenderBalance: any;
    let initialRecipientBalance: any;
    let nativeAmount: any;
    let gasCost: any;
    let mezoTransfersAddress: any;

    before(async function () {
      await fixture();

      nativeAmount = ethers.parseEther("11");
      mezoTransfersAddress = await mezoTransfers.getAddress();

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await mezoTransfers.connect(senderSigner).receiveSendERC20(recipientAddress, { value: nativeAmount });
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

    it("should verify MezoTransfers contract has zero balance", async function () {
      const currentContractBalance = await ethers.provider.getBalance(mezoTransfersAddress);
      expect(currentContractBalance).to.equal(0);
    });
  });

  describe("receiveSendNativeThenERC20", function () {
    let initialSenderBalance: any;
    let initialRecipientBalance: any;
    let gasCost: any;
    let mezoTransfersAddress: any;
    const nativeAmount = ethers.parseEther("3");

    before(async function () {
      await fixture();

      mezoTransfersAddress = await mezoTransfers.getAddress();

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await mezoTransfers.connect(senderSigner).receiveSendNativeThenERC20(recipientAddress, { value: nativeAmount });
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

    it("should verify MezoTransfers contract has zero balance", async function () {
      const currentBalance = await ethers.provider.getBalance(mezoTransfersAddress);
      expect(currentBalance).to.equal(0);
    });
  });

  describe("receiveSendERC20ThenNative", function () {
    const nativeAmount = ethers.parseEther("5");
    let initialSenderBalance: any;
    let initialRecipientBalance: any;
    let gasCost: any;
    let mezoTransfersAddress: any;

    before(async function () {
      await fixture();

      mezoTransfersAddress = await mezoTransfers.getAddress();

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await mezoTransfers.connect(senderSigner).receiveSendERC20ThenNative(recipientAddress, { value: nativeAmount });
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

    it("should verify MezoTransfers contract has zero balance", async function () {
      const currentBalance = await ethers.provider.getBalance(mezoTransfersAddress);
      expect(currentBalance).to.equal(0);
    });
  });

  describe("receiveSendMultiple", function () {
    let initialSenderBalance: any;
    let initialRecipientBalance: any;
    const nativeAmount = ethers.parseEther("8");
    let gasCost: any;
    let mezoTransfersAddress: any;

    before(async function () {
      await fixture();

      mezoTransfersAddress = await mezoTransfers.getAddress();

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await mezoTransfers.connect(senderSigner).receiveSendMultiple(recipientAddress, { value: nativeAmount });
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

    it("should verify MezoTransfers contract has zero balance", async function () {
      const currentBalance = await ethers.provider.getBalance(mezoTransfersAddress);
      expect(currentBalance).to.equal(0);
    });
  });

  describe("receiveSendMultiple_tinyAmount", function () {
    let initialSenderBalance: any;
    let initialRecipientBalance: any;
    const nativeAmount = 4n; // this amount is split by 4 in contract
    let gasCost: any;
    let mezoTransfersAddress: any;

    before(async function () {
      await fixture();

      mezoTransfersAddress = await mezoTransfers.getAddress();

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      // Transfer
      const tx = await mezoTransfers.connect(senderSigner).receiveSendMultiple(recipientAddress, { value: nativeAmount });
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

    it("should verify MezoTransfers contract has zero balance", async function () {
      const currentBalance = await ethers.provider.getBalance(mezoTransfersAddress);
      expect(currentBalance).to.equal(0);
    });
  });

  describe("receiveSendMultiple_hugeAmount", function () {
    const nativeAmount = ethers.parseEther("21000000"); // this amount is split by 4 in contract
    let initialSenderBalance: any;
    let initialRecipientBalance: any;
    let gasCost: any;
    let mezoTransfersAddress: any;

    before(async function () {
      await fixture();

      mezoTransfersAddress = await mezoTransfers.getAddress();

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await mezoTransfers.connect(senderSigner).receiveSendMultiple(recipientAddress, { value: nativeAmount });
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

    it("should verify MezoTransfers contract has zero balance", async function () {
      const currentBalance = await ethers.provider.getBalance(mezoTransfersAddress);
      expect(currentBalance).to.equal(0);
    });
  });

  describe("receiveSendRevert", function () {
    let initialSenderBalance: any;
    let initialRecipientBalance: any;
    const nativeAmount = 42;
    let mezoTransfersAddress: any;

    before(async function () {
      await fixture();

      mezoTransfersAddress = await mezoTransfers.getAddress();

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      try {
        // Transfer
        await mezoTransfers
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

    it("should verify MezoTransfers contract has zero balance", async function () {
      const currentBalance = await ethers.provider.getBalance(mezoTransfersAddress);
      expect(currentBalance).to.equal(0);
    });
  });

  describe("receiveSendNativeThenPullERC20", function () {
    let initialSenderBalance: any;
    let initialRecipientBalance: any;
    let gasCost: any;
    let mezoTransfersAddress: any;
    const nativeAmount = ethers.parseEther("3");
    const tokenAmount = ethers.parseEther("1");

    before(async function () {
      await fixture();

      mezoTransfersAddress = await mezoTransfers.getAddress();

      await btcErc20Token.connect(senderSigner).approve(mezoTransfersAddress, tokenAmount)
        .then(tx => tx.wait());

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await mezoTransfers.connect(senderSigner).receiveSendNativeThenPullERC20(recipientAddress, tokenAmount, { value: nativeAmount });
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

    it("should verify MezoTransfers contract has zero balance", async function () {
      const currentBalance = await ethers.provider.getBalance(mezoTransfersAddress);
      expect(currentBalance).to.equal(0);
    });
  });

  describe("receivePullERC20ThenNative", function () {
    let initialSenderBalance: any;
    let initialRecipientBalance: any;
    const tokenAmount = ethers.parseEther("4");
    const nativeAmount = ethers.parseEther("5");
    let gasCost: any;
    let mezoTransfersAddress: any;

    before(async function () {
      await fixture();

      mezoTransfersAddress = await mezoTransfers.getAddress();

      await btcErc20Token.connect(senderSigner).approve(mezoTransfersAddress, tokenAmount)
        .then(tx => tx.wait());

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await mezoTransfers.connect(senderSigner).receivePullERC20ThenNative(recipientAddress, tokenAmount, { value: nativeAmount });
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

    it("should verify MezoTransfers contract has zero balance", async function () {
      const currentBalance = await ethers.provider.getBalance(mezoTransfersAddress);
      expect(currentBalance).to.equal(0);
    });
  });

  describe("storageStateTransition", function () {
    let initialContractBalance: any;
    let initialRecipientBalance: any;
    let initialSenderBalance: any;
    let amount: any;
    let gasCost: any;
    let mezoTransfersAddress: any;

    before(async function () {
      await fixture();

      amount = ethers.parseEther("2");
      mezoTransfersAddress = await mezoTransfers.getAddress();

      await btcErc20Token.connect(senderSigner).approve(mezoTransfersAddress, amount)
        .then(tx => tx.wait());

      initialContractBalance = await ethers.provider.getBalance(mezoTransfersAddress);
      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

      const tx = await mezoTransfers.connect(senderSigner).storageStateTransition(recipientAddress, amount);
      const receipt = await tx.wait();
      gasCost = receipt.gasUsed * receipt.gasPrice;
    });

    it("should verify sender balance deduction", async function () {
      const currentSenderNativeBalance = await ethers.provider.getBalance(senderAddress);
      expect(initialSenderBalance - amount - gasCost).to.equal(currentSenderNativeBalance);
    });

    it("should verify recipient native balance is zero before the call", async function () {
      expect(initialRecipientBalance).to.equal(0);
    });

    it("should verify recipient native balance increased after the call", async function () {
      const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
      expect(currentRecipientNativeBalance).to.equal(amount);
    });

    it("should verify MezoTransfers contract has zero balance before the call", async function () {
      expect(initialContractBalance).to.equal(0);
    });

    it("should verify MezoTransfers contract has zero balance after the call", async function () {
      const currentContractBalance = await ethers.provider.getBalance(mezoTransfersAddress);
      expect(currentContractBalance).to.equal(0);
    });

    it("should verify balanceTracker storage variable was unchanged", async function () {
      const balanceTracker = await mezoTransfers.balanceTracker();
      expect(balanceTracker).to.equal(0);
    });
  });
});
