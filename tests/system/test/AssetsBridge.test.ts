import { expect } from "chai";
import hre from "hardhat";
import { ethers } from "hardhat";
import assetsbridgeabi from "../../../precompile/assetsbridge/abi.json";
import btcabi from "../../../precompile/btctoken/abi.json";
import validatorpoolabi from "../../../precompile/validatorpool/abi.json";
import { BridgeOut } from "../typechain-types/BridgeOut";
import { SimpleToken } from "../typechain-types/SimpleToken";
import { getDeployedContract } from "./helpers/contract";

const validatorPoolPrecompileAddress = "0x7b7c000000000000000000000000000000000011";
const assetsBridgePrecompileAddress = "0x7b7c000000000000000000000000000000000012";
const btcTokenPrecompileAddress = "0x7b7c000000000000000000000000000000000000";

describe("AssetsBridge", function() {
  const { deployments } = hre;
  let assetsBridge: any;
  let btcToken: any;
  let validatorPool: any;
  let bridgeOut: BridgeOut;
  let simpleToken: SimpleToken;
  let signers: any;
  let senderSigner: any;
  let poolOwner: any;
  let senderAddress: string;
  let contractAddress: string;

  const fixture = async function() {
    await deployments.fixture();
    validatorPool = new hre.ethers.Contract(validatorPoolPrecompileAddress, validatorpoolabi, ethers.provider);
    assetsBridge = new hre.ethers.Contract(assetsBridgePrecompileAddress, assetsbridgeabi, ethers.provider);
    btcToken = new hre.ethers.Contract(btcTokenPrecompileAddress, btcabi, ethers.provider);
    bridgeOut = await getDeployedContract("BridgeOut");
    simpleToken = await getDeployedContract("SimpleToken");
    signers = await ethers.getSigners();
    contractAddress = await bridgeOut.getAddress();
    poolOwner = await ethers.getSigner(await validatorPool.owner());

    // define the address used in th test
    senderSigner = ethers.Wallet.createRandom().connect(ethers.provider);
    senderAddress = senderSigner.address;

    // now send funds to the random address used to do the tests
    // send 10 eth to the address
    var fundingSigner = signers[0];
    const transferTx = await btcToken.connect(fundingSigner).transfer(senderSigner, ethers.parseEther("10"));
    await transferTx.wait();
  };

  describe("bridgeOutBtcFailureApprovalTooSmall", function() {
    let receipt1: any;
    let receipt2: any;
    let tx: any;
    let tokenAmount: any;
    let initialSenderBalance: any;
    let recipient = "150bCF49Ee8E2Bd9f59e991821DE5B74C6D876aA";
    let gasCost = 0;
    let errorMessage: string;

    before(async function() {
      await fixture();
      tokenAmount = ethers.parseEther("8");

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);

      // approve for token amount
      tx = await btcToken.connect(senderSigner)
        .approve(assetsBridgePrecompileAddress, tokenAmount / 2n);
      await tx.wait();
      receipt1 = await ethers.provider.getTransactionReceipt(tx.hash);
      gasCost = receipt1.gasUsed * receipt1.gasPrice;

      try {
        await assetsBridge.connect(senderSigner).bridgeOut(
          btcTokenPrecompileAddress,
          tokenAmount,
          0,
          Buffer.from(recipient, "hex"),
        );
      } catch (error: any) {
        errorMessage = error.message;
      }

      try {
        tx = await assetsBridge.connect(senderSigner).bridgeOut(
          btcTokenPrecompileAddress,
          tokenAmount,
          0,
          Buffer.from(recipient, "hex"),
          { gasLimit: 100000 },
        );
        await tx.wait();
      } catch (error: any) {
        expect(error.message).to.include(
          "reverted",
        );
      }
      receipt2 = await ethers.provider.getTransactionReceipt(tx.hash);
      gasCost += receipt2.gasUsed * receipt2.gasPrice;
    });

    // clean up for following tests
    after(async function() {
      tx = await btcToken.connect(senderSigner)
        .approve(assetsBridgePrecompileAddress, 0);
      await tx.wait();
    });

    it("should verify the transaction failed", async function() {
      expect(receipt1!.status).to.equal(1);
    });

    it("should verify the error message", async function() {
      expect(errorMessage).to.include(
        "couldn't accept authorization: requested amount is more than spend limit: insufficient funds",
      );
    });

    it("should verify the balance hasn't changed", async function() {
      var updatedSenderBalance = await ethers.provider.getBalance(senderAddress);
      expect(updatedSenderBalance).to.equal(initialSenderBalance - gasCost);
    });

    it("should verify that BTC and BTC ERC20 balance are equal", async function() {
      expect(await ethers.provider.getBalance(senderAddress)).to.equal(await btcToken.balanceOf(senderAddress));
    });
  });

  describe("bridgeOutBtcFailureInvalidChain", function() {
    let receipt1: any;
    let receipt2: any;
    let tx: any;
    let tokenAmount: any;
    let initialSenderBalance: any;
    let recipient = "150bCF49Ee8E2Bd9f59e991821DE5B74C6D876aA";
    let gasCost = 0;
    let errorMessage: string;

    before(async function() {
      await fixture();
      tokenAmount = ethers.parseEther("8");

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);

      // approve for token amount
      tx = await btcToken.connect(senderSigner)
        .approve(assetsBridgePrecompileAddress, tokenAmount);
      await tx.wait();
      receipt1 = await ethers.provider.getTransactionReceipt(tx.hash);
      gasCost = receipt1.gasUsed * receipt1.gasPrice;

      try {
        await assetsBridge.connect(senderSigner).bridgeOut(
          btcTokenPrecompileAddress,
          tokenAmount,
          10,
          Buffer.from(recipient, "hex"),
        );
      } catch (error: any) {
        errorMessage = error.message;
      }

      try {
        tx = await assetsBridge.connect(senderSigner).bridgeOut(
          btcTokenPrecompileAddress,
          tokenAmount,
          10,
          Buffer.from(recipient, "hex"),
          { gasLimit: 100000 },
        );
        await tx.wait();
      } catch (error: any) {
        expect(error.message).to.include("reverted");
      }
      receipt2 = await ethers.provider.getTransactionReceipt(tx.hash);
      gasCost = gasCost + receipt2.gasUsed * receipt2.gasPrice;
    });

    // clean up for following tests
    after(async function() {
      tx = await btcToken.connect(senderSigner)
        .approve(assetsBridgePrecompileAddress, 0);
      await tx.wait();
    });

    it("should verify the transaction failed", async function() {
      expect(receipt1!.status).to.equal(1);
      expect(receipt2!.status).to.equal(0);
    });

    it("should verify the error message", async function() {
      expect(errorMessage).to.include("unsupported chain: 10");
    });

    it("should verify the balance hasn't changed", async function() {
      var updatedSenderBalance = await ethers.provider.getBalance(senderAddress);
      expect(updatedSenderBalance).to.equal(initialSenderBalance - gasCost);
    });

    it("should verify that BTC and BTC ERC20 balance are equal", async function() {
      expect(await ethers.provider.getBalance(senderAddress)).to.equal(await btcToken.balanceOf(senderAddress));
    });
  });

  describe("bridgeOutBtcFailureInvalidEthereumRecipient", function() {
    let receipt1: any;
    let receipt2: any;
    let tx: any;
    let tokenAmount: any;
    let initialSenderBalance: any;
    let recipient = "150bCF49Ee8E2Bd9f59e99182";
    let gasCost = 0;
    let errorMessage: string;

    before(async function() {
      await fixture();
      tokenAmount = ethers.parseEther("8");

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);

      // approve for token amount
      tx = await btcToken.connect(senderSigner)
        .approve(assetsBridgePrecompileAddress, tokenAmount);
      await tx.wait();
      receipt1 = await ethers.provider.getTransactionReceipt(tx.hash);
      gasCost = receipt1.gasUsed * receipt1.gasPrice;

      try {
        await assetsBridge.connect(senderSigner).bridgeOut(
          btcTokenPrecompileAddress,
          tokenAmount,
          0,
          Buffer.from(recipient, "hex"),
        );
      } catch (error: any) {
        errorMessage = error.message;
      }

      try {
        tx = await assetsBridge.connect(senderSigner).bridgeOut(
          btcTokenPrecompileAddress,
          tokenAmount,
          0,
          Buffer.from(recipient, "hex"),
          { gasLimit: 100000 },
        );
        await tx.wait();
      } catch (error: any) {
        expect(error.message).to.include("reverted");
      }

      receipt2 = await ethers.provider.getTransactionReceipt(tx.hash);
      gasCost = gasCost + receipt2.gasUsed * receipt2.gasPrice;
    });

    // clean up for following tests
    after(async function() {
      tx = await btcToken.connect(senderSigner)
        .approve(assetsBridgePrecompileAddress, 0);
      await tx.wait();
    });

    it("should verify the transaction failed", async function() {
      expect(receipt1!.status).to.equal(1);
      expect(receipt2!.status).to.equal(0);
    });

    it("should verify the error message", async function() {
      expect(errorMessage).to.include("invalid recipient address for Ethereum chain: 150bcf49ee8e2bd9f59e9918");
    });

    it("should verify the balance hasn't changed", async function() {
      var updatedSenderBalance = await ethers.provider.getBalance(senderAddress);
      expect(updatedSenderBalance).to.equal(initialSenderBalance - gasCost);
    });

    it("should verify that BTC and BTC ERC20 balance are equal", async function() {
      expect(await ethers.provider.getBalance(senderAddress)).to.equal(await btcToken.balanceOf(senderAddress));
    });
  });

  describe("bridgeOutBtcFailureNotEnoughFunds", function() {
    let receipt1: any;
    let receipt2: any;
    let tx: any;
    let tokenAmount: any;
    let initialSenderBalance: any;
    let recipient = "150bCF49Ee8E2Bd9f59e991821DE5B74C6D876aA";
    let gasCost = 0;
    let errorMessage: string;

    before(async function() {
      await fixture();
      tokenAmount = ethers.parseEther("12");

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);

      // approve for token amount
      tx = await btcToken.connect(senderSigner)
        .approve(assetsBridgePrecompileAddress, tokenAmount);
      await tx.wait();
      receipt1 = await ethers.provider.getTransactionReceipt(tx.hash);
      gasCost = receipt1.gasUsed * receipt1.gasPrice;

      try {
        await assetsBridge.connect(senderSigner).bridgeOut(
          btcTokenPrecompileAddress,
          tokenAmount,
          0,
          Buffer.from(recipient, "hex"),
        );
      } catch (error: any) {
        errorMessage = error.message;
      }

      try {
        tx = await assetsBridge.connect(senderSigner).bridgeOut(
          btcTokenPrecompileAddress,
          tokenAmount,
          0,
          Buffer.from(recipient, "hex"),
          { gasLimit: 100000 },
        );
        await tx.wait();
      } catch (error: any) {
        expect(error.message).to.include(
          "reverted",
        );
      }
      receipt2 = await ethers.provider.getTransactionReceipt(tx.hash);
      gasCost = gasCost + receipt2.gasUsed * receipt2.gasPrice;
    });

    // clean up for following tests
    after(async function() {
      tx = await btcToken.connect(senderSigner)
        .approve(assetsBridgePrecompileAddress, 0);
      await tx.wait();
    });

    it("should verify the transaction failed", async function() {
      expect(receipt1!.status).to.equal(1);
      expect(receipt2!.status).to.equal(0);
    });

    it("should verify the error message", async function() {
      expect(errorMessage).to.include(
        "insufficient funds",
      );
    });

    it("should verify the balance hasn't changed", async function() {
      var updatedSenderBalance = await ethers.provider.getBalance(senderAddress);
      expect(updatedSenderBalance).to.equal(initialSenderBalance - gasCost);
    });

    it("should verify that BTC and BTC ERC20 balance are equal", async function() {
      expect(await ethers.provider.getBalance(senderAddress)).to.equal(await btcToken.balanceOf(senderAddress));
    });
  });

  describe("bridgeOutERC20ToEthereumFailureNotMapped", function() {
    let tx: any;
    let tokenAmount: any;
    let recipient = "150bCF49Ee8E2Bd9f59e991821DE5B74C6D876aA";
    let errorMessage: string;

    before(async function() {
      await fixture();

      // mint some token to ourselves
      tokenAmount = 1000;
      // approve for token amount
      tx = await simpleToken.connect(senderSigner)
        .mint(senderAddress, tokenAmount);
      await tx.wait();

      // approve for token amount
      tx = await simpleToken.connect(senderSigner)
        .approve(assetsBridgePrecompileAddress, tokenAmount);
      await tx.wait();

      try {
        await assetsBridge.connect(senderSigner).bridgeOut(
          await simpleToken.getAddress(),
          tokenAmount,
          0,
          Buffer.from(recipient, "hex"),
        );
      } catch (error: any) {
        errorMessage = error.message;
      }

      try {
        tx = await assetsBridge.connect(senderSigner).bridgeOut(
          await simpleToken.getAddress(),
          tokenAmount,
          0,
          Buffer.from(recipient, "hex"),
          { gasLimit: 100000 },
        );
        await tx.wait();
      } catch (error: any) {
        expect(error.message).to.include(
          "reverted",
        );
      }
    });

    it("should verify the error message", async function() {
      expect(errorMessage).to.include(
        "unsupported token",
      );
      expect(errorMessage).to.include(
        "for ethereum target chain",
      );
    });

    it("should verify the new balances", async function() {
      var updatedSenderBalance = await simpleToken.balanceOf(senderAddress);
      expect(updatedSenderBalance).to.equal(1000);
    });
  });

  describe("bridgeOutERC20ToBitcoinInvalid", function() {
    let tx: any;
    let tokenAmount: any;
    let recipient = "1976a91462e907b15cbf27d5425399ebf6f0fb50ebb88f1888ac";
    let errorMessage: string;

    before(async function() {
      await fixture();

      // mint some token to ourselves
      tokenAmount = 1000;
      // approve for token amount
      tx = await simpleToken.connect(senderSigner)
        .mint(senderAddress, tokenAmount);
      await tx.wait();

      // approve for token amount
      tx = await simpleToken.connect(senderSigner)
        .approve(assetsBridgePrecompileAddress, tokenAmount);
      await tx.wait();

      try {
        await assetsBridge.connect(senderSigner).bridgeOut(
          await simpleToken.getAddress(),
          tokenAmount,
          1,
          Buffer.from(recipient, "hex"),
        );
      } catch (error: any) {
        errorMessage = error.message;
      }

      try {
        tx = await assetsBridge.connect(senderSigner).bridgeOut(
          await simpleToken.getAddress(),
          tokenAmount,
          1,
          Buffer.from(recipient, "hex"),
          { gasLimit: 1000000 },
        );
        await tx.wait();
      } catch (error: any) {
        expect(error.message).to.include("revert");
      }
    });

    it("should verify the error message", async function() {
      expect(errorMessage).to.include(
        "unsupported token",
      );
      expect(errorMessage).to.include(
        "for bitcoin target chain",
      );
    });

    it("should verify the new balances", async function() {
      var updatedSenderBalance = await simpleToken.balanceOf(senderAddress);
      expect(updatedSenderBalance).to.equal(1000);
    });
  });

  describe("bridgeOutERC20ToEthereumFailureNotApproved", function() {
    let tokenAmount: any;
    let recipient = "150bCF49Ee8E2Bd9f59e991821DE5B74C6D876aA";
    let sourceTokenAddress = ethers.Wallet.createRandom().address;
    let errorMessage: string;

    before(async function() {
      await fixture();

      // do the erc20 token mapping
      let tx = await assetsBridge.connect(poolOwner).createERC20TokenMapping(
        sourceTokenAddress,
        await simpleToken.getAddress(),
      );
      await tx.wait();

      // mint some token to ourselves
      tokenAmount = 1000;
      // approve for token amount
      tx = await simpleToken.connect(senderSigner)
        .mint(senderAddress, tokenAmount);
      await tx.wait();

      try {
        await assetsBridge.connect(senderSigner).bridgeOut(
          await simpleToken.getAddress(),
          tokenAmount,
          0,
          Buffer.from(recipient, "hex"),
        );
      } catch (error: any) {
        errorMessage = error.message;
      }

      try {
        tx = await assetsBridge.connect(senderSigner).bridgeOut(
          await simpleToken.getAddress(),
          tokenAmount,
          0,
          Buffer.from(recipient, "hex"),
          { gasLimit: 1000000 },
        );
        await tx.wait();
      } catch (error: any) {
        expect(error.message).to.include("reverted");
      }
    });

    after(async function() {
      let tx = await assetsBridge.connect(poolOwner).deleteERC20TokenMapping(
        sourceTokenAddress,
      );
      await tx.wait();
    });
    it("should verify error message", async function() {
      expect(errorMessage).to.include(
        "failed to execute ERC20 burnFrom call: execution reverted: evm transaction execution failed",
      );
    });

    it("should verify the new balances", async function() {
      var updatedSenderBalance = await simpleToken.balanceOf(senderAddress);
      expect(updatedSenderBalance).to.equal(1000);
    });
  });

  describe("bridgeOutBtcFailureInvalidBitcoinRecipient", function() {
    let tx: any;
    let receipt1: any;
    let receipt2: any;
    let tokenAmount: any;
    let initialSenderBalance: any;
    let recipient = "150bCF49Ee8E2";
    let gasCost = 0;
    let errorMessage: string;

    before(async function() {
      await fixture();
      tokenAmount = ethers.parseEther("8");

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);

      // approve for token amount
      let tx = await btcToken.connect(senderSigner)
        .approve(assetsBridgePrecompileAddress, tokenAmount);
      await tx.wait();
      receipt1 = await ethers.provider.getTransactionReceipt(tx.hash);
      gasCost = receipt1.gasUsed * receipt1.gasPrice;

      try {
        await assetsBridge.connect(senderSigner).bridgeOut(
          btcTokenPrecompileAddress,
          tokenAmount,
          1,
          Buffer.from(recipient, "hex"),
        );
      } catch (error: any) {
        errorMessage = error.message;
      }

      try {
        tx = await assetsBridge.connect(senderSigner).bridgeOut(
          btcTokenPrecompileAddress,
          tokenAmount,
          1,
          Buffer.from(recipient, "hex"),
          { gasLimit: 1000000 },
        );
        await tx.wait();
      } catch (error: any) {
        expect(error.message).to.include("reverted");
      }

      receipt2 = await ethers.provider.getTransactionReceipt(tx.hash);
      gasCost = gasCost + receipt2.gasUsed * receipt2.gasPrice;
    });

    // clean up for following tests
    after(async function() {
      let tx = await btcToken.connect(senderSigner)
        .approve(assetsBridgePrecompileAddress, 0);
      await tx.wait();
    });

    it("should verify the error message", async function() {
      expect(errorMessage).to.include("couldn't get script from var-len data: malformed var len data");
    });

    it("should verify the transaction failed", async function() {
      expect(receipt1!.status).to.equal(1);
      expect(receipt2!.status).to.equal(0);
    });

    it("should verify the balance hasn't changed", async function() {
      var updatedSenderBalance = await ethers.provider.getBalance(senderAddress);
      expect(updatedSenderBalance).to.equal(initialSenderBalance - gasCost);
    });

    it("should verify that BTC and BTC ERC20 balance are equal", async function() {
      expect(await ethers.provider.getBalance(senderAddress)).to.equal(await btcToken.balanceOf(senderAddress));
    });
  });

  describe("bridgeOutBtcToEthereumFailureNotApproved", function() {
    let tx: any;
    let tokenAmount: any;
    let initialSenderBalance: any;
    let recipient = "150bCF49Ee8E2Bd9f59e991821DE5B74C6D876aA";
    let errorMessage: string;
    let receipt: any;
    let gasCost = 0;

    before(async function() {
      await fixture();
      tokenAmount = ethers.parseEther("8");
      initialSenderBalance = await ethers.provider.getBalance(senderAddress);

      try {
        await assetsBridge.connect(senderSigner).bridgeOut(
          btcTokenPrecompileAddress,
          tokenAmount,
          0,
          Buffer.from(recipient, "hex"),
        );
      } catch (error: any) {
        errorMessage = error.message;
      }

      try {
        tx = await assetsBridge.connect(senderSigner).bridgeOut(
          btcTokenPrecompileAddress,
          tokenAmount,
          0,
          Buffer.from(recipient, "hex"),
          { gasLimit: 1000000 },
        );
        await tx.wait();
      } catch (error: any) {
        expect(error.message).to.include(
          "reverted",
        );
      }

      receipt = await ethers.provider.getTransactionReceipt(tx.hash);
      gasCost = receipt.gasUsed * receipt.gasPrice;
    });
    it("should verify the transaction failed", async function() {
      expect(receipt!.status).to.equal(0);
    });

    it("should have had the specific error when estimating gas", async function() {
      expect(errorMessage).to.include(
        "/cosmos.bank.v1beta1.MsgSend authorization type does not exist or is expired for address",
      );
    });

    it("should verify the balance hasn't changed", async function() {
      var updatedSenderBalance = await ethers.provider.getBalance(senderAddress);
      expect(updatedSenderBalance).to.equal(initialSenderBalance - gasCost);
    });

    it("should verify that BTC and BTC ERC20 balance has same balances", async function() {
      expect(await ethers.provider.getBalance(senderAddress)).to.equal(await btcToken.balanceOf(senderAddress));
    });
  });

  describe("contractBridgeOutBtcToBitcoinReverts", function() {
    let receipt: any;
    let tx: any;
    let tokenAmount: any;
    let initialSenderBalance: any;
    let initialContractBalance: any;
    let recipient = "1976a91462e907b15cbf27d5425399ebf6f0fb50ebb88f1888ac";
    let gasCost = 0;

    before(async function() {
      await fixture();
      tokenAmount = ethers.parseEther("8");

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);

      // send the funds to the contract first
      const transferTx = await btcToken.connect(senderSigner).transfer(contractAddress, tokenAmount);
      await transferTx.wait();
      receipt = await ethers.provider.getTransactionReceipt(transferTx.hash);
      gasCost = receipt.gasUsed * receipt.gasPrice;

      initialContractBalance = await ethers.provider.getBalance(contractAddress);

      try {
        tx = await bridgeOut.connect(senderSigner).bridgeOutBTCToBitcoinReverts(
          Buffer.from(recipient, "hex"),
          tokenAmount,
          { gasLimit: 1000000 },
        );
        await tx.wait();
      } catch (error: any) {
        expect(error.shortMessage).to.include(
          "execution reverted",
        );
      }

      receipt = await ethers.provider.getTransactionReceipt(tx.hash);
      gasCost = gasCost + receipt.gasUsed * receipt.gasPrice;
    });

    it("should verify the transaction did revert", async function() {
      expect(receipt!.status).to.equal(0);
    });

    it("should verify the new balances", async function() {
      var updatedSenderBalance = await ethers.provider.getBalance(senderAddress);
      expect(updatedSenderBalance).to.equal(initialSenderBalance - gasCost - tokenAmount);
    });

    it("should verify that BTC and BTC ERC20 balance are equal", async function() {
      expect(await ethers.provider.getBalance(senderAddress))
        .to.equal(await btcToken.balanceOf(senderAddress));
    });

    it("should verify that contract address balances haven't changed", async function() {
      expect(await ethers.provider.getBalance(contractAddress))
        .to.equal(await btcToken.balanceOf(contractAddress))
        .to.equal(ethers.parseEther("8"));
    });
  });

  describe("contractBridgeOutBtcSuccessToBitcoin", function() {
    let receipt: any;
    let tx: any;
    let tokenAmount: any;
    let initialSenderBalance: any;
    let recipient = "1976a91462e907b15cbf27d5425399ebf6f0fb50ebb88f1888ac";
    let gasCost = 0;

    before(async function() {
      await fixture();
      tokenAmount = ethers.parseEther("8");

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);

      // send the funds to the contract first
      const transferTx = await btcToken.connect(senderSigner).transfer(contractAddress, tokenAmount);
      await transferTx.wait();
      receipt = await ethers.provider.getTransactionReceipt(transferTx.hash);
      gasCost = receipt.gasUsed * receipt.gasPrice;

      tx = await bridgeOut.connect(senderSigner).bridgeOutBTCToBitcoinSuccess(
        Buffer.from(recipient, "hex"),
        tokenAmount,
      );
      await tx.wait();

      receipt = await ethers.provider.getTransactionReceipt(tx.hash);
      gasCost = gasCost + receipt.gasUsed * receipt.gasPrice;
    });

    it("should verify the transaction didn't revert", async function() {
      expect(receipt!.status).to.equal(1);
    });

    it("should verify the new balances", async function() {
      var updatedSenderBalance = await ethers.provider.getBalance(senderAddress);
      expect(updatedSenderBalance).to.equal(initialSenderBalance - gasCost - tokenAmount);
    });

    it("should verify that BTC and BTC ERC20 balance are equal", async function() {
      expect(await ethers.provider.getBalance(senderAddress))
        .to.equal(await btcToken.balanceOf(senderAddress));
    });

    it("should verify that contract address have 0 balance", async function() {
      expect(await ethers.provider.getBalance(contractAddress))
        .to.equal(await btcToken.balanceOf(contractAddress))
        .to.equal(0n);
    });
  });

  describe("bridgeOutBtcSuccessToEthereum", function() {
    let receipt1: any;
    let receipt2: any;
    let tx: any;
    let tokenAmount: any;
    let initialSenderBalance: any;
    let recipient = "150bCF49Ee8E2Bd9f59e991821DE5B74C6D876aA";
    let gasCost = 0;

    before(async function() {
      await fixture();
      tokenAmount = ethers.parseEther("8");

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);

      // approve for token amount
      tx = await btcToken.connect(senderSigner)
        .approve(assetsBridgePrecompileAddress, tokenAmount);
      await tx.wait();
      receipt1 = await ethers.provider.getTransactionReceipt(tx.hash);
      gasCost = receipt1.gasUsed * receipt1.gasPrice;

      tx = await assetsBridge.connect(senderSigner).bridgeOut(
        btcTokenPrecompileAddress,
        tokenAmount,
        0,
        Buffer.from(recipient, "hex"),
      );
      await tx.wait();
      receipt2 = await ethers.provider.getTransactionReceipt(tx.hash);
      gasCost = gasCost + receipt2.gasUsed * receipt2.gasPrice;
    });

    it("should verify the transaction didn't revert", async function() {
      expect(receipt1!.status).to.equal(1);
      expect(receipt2!.status).to.equal(1);
    });

    it("should verify the new balances", async function() {
      var updatedSenderBalance = await ethers.provider.getBalance(senderAddress);
      expect(updatedSenderBalance).to.equal(initialSenderBalance - gasCost - tokenAmount);
    });

    it("should verify that BTC and BTC ERC20 balance are equal", async function() {
      expect(await ethers.provider.getBalance(senderAddress)).to.equal(await btcToken.balanceOf(senderAddress));
    });
  });

  describe("bridgeOutBtcSuccessToBitcoin", function() {
    let receipt1: any;
    let receipt2: any;
    let tx: any;
    let tokenAmount: any;
    let initialSenderBalance: any;
    let recipient = "1976a91462e907b15cbf27d5425399ebf6f0fb50ebb88f1888ac";
    let gasCost = 0;

    before(async function() {
      await fixture();
      tokenAmount = ethers.parseEther("8");

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);

      // approve for token amount
      tx = await btcToken.connect(senderSigner)
        .approve(assetsBridgePrecompileAddress, tokenAmount);
      await tx.wait();
      receipt1 = await ethers.provider.getTransactionReceipt(tx.hash);
      gasCost = receipt1.gasUsed * receipt1.gasPrice;

      tx = await assetsBridge.connect(senderSigner).bridgeOut(
        btcTokenPrecompileAddress,
        tokenAmount,
        1,
        Buffer.from(recipient, "hex"),
      );
      await tx.wait();
      receipt2 = await ethers.provider.getTransactionReceipt(tx.hash);
      gasCost = gasCost + receipt2.gasUsed * receipt2.gasPrice;
    });

    it("should verify the transaction didn't revert", async function() {
      expect(receipt1!.status).to.equal(1);
      expect(receipt2!.status).to.equal(1);
    });

    it("should verify the new balances", async function() {
      var updatedSenderBalance = await ethers.provider.getBalance(senderAddress);
      expect(updatedSenderBalance).to.equal(initialSenderBalance - gasCost - tokenAmount);
    });

    it("should verify that BTC and BTC ERC20 balance are equal", async function() {
      expect(await ethers.provider.getBalance(senderAddress)).to.equal(await btcToken.balanceOf(senderAddress));
    });
  });

  describe("bridgeOutERC20SuccessToEthereum", function() {
    let receipt: any;
    let tx: any;
    let tokenAmount: any;
    let recipient = "150bCF49Ee8E2Bd9f59e991821DE5B74C6D876aA";
    let sourceTokenAddress = ethers.Wallet.createRandom().address;

    before(async function() {
      await fixture();

      // do the erc20 token mapping
      let tx = await assetsBridge.connect(poolOwner).createERC20TokenMapping(
        sourceTokenAddress,
        await simpleToken.getAddress(),
      );
      await tx.wait();

      // mint some token to ourselves
      tokenAmount = ethers.parseEther("10");
      // approve for token amount
      tx = await simpleToken.connect(senderSigner)
        .mint(senderAddress, tokenAmount);
      await tx.wait();

      // approve for token amount
      tx = await simpleToken.connect(senderSigner)
        .approve(assetsBridgePrecompileAddress, tokenAmount);
      await tx.wait();

      tx = await assetsBridge.connect(senderSigner).bridgeOut(
        await simpleToken.getAddress(),
        tokenAmount,
        0,
        Buffer.from(recipient, "hex"),
      );
      await tx.wait();
      receipt = await ethers.provider.getTransactionReceipt(tx.hash);
    });

    after(async function() {
      let tx = await assetsBridge.connect(poolOwner).deleteERC20TokenMapping(
        sourceTokenAddress,
      );
      await tx.wait();
    });

    it("should verify the transaction didn't revert", async function() {
      expect(receipt!.status).to.equal(1);
    });

    it("should verify the new balances", async function() {
      var updatedSenderBalance = await simpleToken.balanceOf(senderAddress);
      expect(updatedSenderBalance).to.equal(0n);
    });
  });
});
