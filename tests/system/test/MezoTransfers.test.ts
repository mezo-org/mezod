import { MezoTransfers } from '../typechain-types/contracts/MezoTransfers';
import { expect } from "chai";
import hre, { ethers } from "hardhat";
import { getDeployedContract } from "./helpers/contract";
import abi from '../../../precompile/btctoken/abi.json';

const precompileAddress = '0x7b7c000000000000000000000000000000000000';

describe("MezoTransfers", function () {
  const { deployments } = hre;
  let btcErc20Token: any;
  let mezoTransfers: MezoTransfers;
  let signers: any;
  let senderAddress: string;
  let recipientAddress: string;
  
  async function fixture() {
    await deployments.fixture();
    btcErc20Token = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider);
    mezoTransfers = await getDeployedContract("MezoTransfers");
    signers = await ethers.getSigners();
    // TODO: Use a random address for sender and add some funds to it.
    senderAddress = signers[0].address;
    recipientAddress = ethers.Wallet.createRandom().address;
  }

  before(async function () {
    await fixture();
  });

  const testCases = [
    {
      // WARNING: Sender's and recipient's balances are wrong.
      description: "nativeThenBTCERC20",
      tokenAmount: ethers.parseEther("2"),
      nativeAmount: ethers.parseEther("3"),
      transferFunction: async (recipientAddress: string, tokenAmount: any, nativeAmount: any) => 
        mezoTransfers.connect(signers[0]).nativeThenBTCERC20(recipientAddress, tokenAmount, { value: nativeAmount }),
    },
    {
      // WARNING: Sender's balance is wrong.
      description: "btcERC20ThenNative",
      tokenAmount: ethers.parseEther("4"),
      nativeAmount: ethers.parseEther("5"),
      transferFunction: async (recipientAddress: string, tokenAmount: any, nativeAmount: any) => 
        mezoTransfers.connect(signers[0]).btcERC20ThenNative(recipientAddress, tokenAmount, { value: nativeAmount }),
    }
  ];

  for (const testCase of testCases) {
    describe(testCase.description, function () {
      let initialSenderBalance: any;
      let initialRecipientBalance: any;
      let gasCost: any;
      
      beforeEach(async () => {
        initialSenderBalance = await ethers.provider.getBalance(senderAddress);
        initialRecipientBalance = await ethers.provider.getBalance(recipientAddress);

        const mezoTransfersAddress = await mezoTransfers.getAddress();
        const approveTx = await btcErc20Token.connect(signers[0])
            .approve(mezoTransfersAddress, testCase.tokenAmount);
        await approveTx.wait();

        const tx = await testCase.transferFunction(
          recipientAddress, testCase.tokenAmount, testCase.nativeAmount);

        const receipt = await tx.wait();
        gasCost = receipt.gasUsed * tx.gasPrice;
      });

      it("should verify balances are the same using native and btc precompile", async function () {
        const currentSenderNativeBalance = await ethers.provider.getBalance(senderAddress);
        const currentSenderBTCERC20Balance = await btcErc20Token.balanceOf(senderAddress);
        expect(currentSenderNativeBalance).to.equal(currentSenderBTCERC20Balance);

        const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
        const currentRecipientTokenBalance = await btcErc20Token.balanceOf(recipientAddress);
        expect(currentRecipientNativeBalance).to.equal(currentRecipientTokenBalance);
      });

      it("should verify sender balance", async function () {
        const currentSenderNativeBalance = await ethers.provider.getBalance(senderAddress);
        expect(initialSenderBalance - testCase.nativeAmount - testCase.tokenAmount - gasCost).to.equal(currentSenderNativeBalance)
      });

      it("should verify recipient balance", async function () {
        const currentRecipientNativeBalance = await ethers.provider.getBalance(recipientAddress);
        expect(initialRecipientBalance + testCase.nativeAmount + testCase.tokenAmount).to.equal(currentRecipientNativeBalance);
      });
    });
  }
});