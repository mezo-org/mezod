import { expect } from "chai";
import hre from "hardhat";
import { ethers } from "hardhat";
import assetsbridgeabi from "../../../precompile/assetsbridge/abi.json";
import btcabi from "../../../precompile/btctoken/abi.json";
import maintenanceabi from "../../../precompile/maintenance/abi.json";
import mezoabi from "../../../precompile/mezotoken/abi.json";
import validatorpoolabi from "../../../precompile/validatorpool/abi.json";
import { BridgeOutDelegate } from "../typechain-types/BridgeOutDelegate";
import { SimpleToken } from "../typechain-types/SimpleToken";
import { getDeployedContract } from "./helpers/contract";
import { extractMessage } from "./helpers/rpc-error";

const validatorPoolPrecompileAddress = "0x7b7c000000000000000000000000000000000011";
const assetsBridgePrecompileAddress = "0x7b7c000000000000000000000000000000000012";
const maintenancePrecompileAddress = "0x7b7c000000000000000000000000000000000013";
const btcTokenPrecompileAddress = "0x7b7c000000000000000000000000000000000000";
const mezoTokenPrecompileAddress = "0x7b7c000000000000000000000000000000000001";

// NOTE: IN MOST OF THE FOLLOWING TESTS WE EXECUTE TWICE EVERY TRANSACTION
// A FIRST TIME TO SIMULATE THE TRANSACTION, WHICH GIVE US PRECISE ERROR
// MESSAGE FROM THE EXECUTION LAYER. A SECOND TIME WHICH ACTUALLY UPDATE
// THE STATE, WHICH WE CAN THE ASSERT, BUT ALSO RETURN LESS PRECISE ERRORS.

describe("AssetsBridge", function() {
  const { deployments } = hre;
  let assetsBridge: any;
  let btcToken: any;
  let mezoToken: any;
  let validatorPool: any;
  let bridgeOutDelegate: BridgeOutDelegate;
  let simpleToken: SimpleToken;
  let signers: any;
  let senderSigner: any;
  let poolOwner: any;
  let senderAddress: string;
  let contractAddress: string;

  const fixture = async function() {
    await deployments.fixture(["BridgeOutDelegate"]);
    validatorPool = new hre.ethers.Contract(validatorPoolPrecompileAddress, validatorpoolabi, ethers.provider);
    assetsBridge = new hre.ethers.Contract(assetsBridgePrecompileAddress, assetsbridgeabi, ethers.provider);
    btcToken = new hre.ethers.Contract(btcTokenPrecompileAddress, btcabi, ethers.provider);
    mezoToken = new hre.ethers.Contract(mezoTokenPrecompileAddress, mezoabi, ethers.provider);
    bridgeOutDelegate = await getDeployedContract("BridgeOutDelegate");
    simpleToken = await getDeployedContract("SimpleToken");
    signers = await ethers.getSigners();
    contractAddress = await bridgeOutDelegate.getAddress();
    poolOwner = await ethers.getSigner(await validatorPool.owner());

    // define the address used in th test
    senderSigner = ethers.Wallet.createRandom().connect(ethers.provider);
    senderAddress = senderSigner.address;

    // now send funds to the random address used to do the tests
    // send 10 eth to the address
    var fundingSigner = signers[0];
    const transferTx = await btcToken.connect(fundingSigner).transfer(senderSigner, ethers.parseEther("10"));
    await transferTx.wait();

    // set the outflow limits to the maximum value of uint256
    await assetsBridge.connect(poolOwner).setOutflowLimit(btcTokenPrecompileAddress, ethers.MaxUint256);
    await assetsBridge.connect(poolOwner).setOutflowLimit(await simpleToken.getAddress(), ethers.MaxUint256);
  };

  describe("bridgeOutBTCFailureNoAllowance", function() {
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

  describe("bridgeOutBTCFailureNoBalance", function() {
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

  describe("bridgeOutBTCSuccess", function() {
    let receipt1: any;
    let receipt2: any;
    let receipt3: any;
    let tx: any;
    let tokenAmount: any;
    let mezoTokenAmount: any;
    let initialSenderBalance: any;
    let recipient = "150bCF49Ee8E2Bd9f59e991821DE5B74C6D876aA";
    let gasCost = 0;
    let totalSupply = 0;

    before(async function() {
      await fixture();
      tokenAmount = ethers.parseEther("8");
      mezoTokenAmount = ethers.parseEther("42");

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);
      totalSupply = await btcToken.totalSupply();

      // approve for btc token amount
      tx = await btcToken.connect(senderSigner)
        .approve(assetsBridgePrecompileAddress, tokenAmount);
      await tx.wait();
      receipt1 = await ethers.provider.getTransactionReceipt(tx.hash);
      gasCost = receipt1.gasUsed * receipt1.gasPrice;

      // approve for mezo token amount
      tx = await mezoToken.connect(senderSigner)
        .approve(assetsBridgePrecompileAddress, mezoTokenAmount);
      await tx.wait();
      receipt2 = await ethers.provider.getTransactionReceipt(tx.hash);
      gasCost = gasCost + receipt2.gasUsed * receipt2.gasPrice;

      tx = await assetsBridge.connect(senderSigner).bridgeOut(
        btcTokenPrecompileAddress,
        tokenAmount,
        0,
        Buffer.from(recipient, "hex"),
      );
      await tx.wait();
      receipt3 = await ethers.provider.getTransactionReceipt(tx.hash);
      gasCost = gasCost + receipt3.gasUsed * receipt3.gasPrice;
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

    it("should verify the remaining BTC and MEZO approvals", async function() {
      expect(await btcToken.allowance(senderAddress, assetsBridgePrecompileAddress)).to.equal(0);
      expect(await mezoToken.allowance(senderAddress, assetsBridgePrecompileAddress)).to.equal(ethers.parseEther("42"));
    });
    it("should verify the totalSupply", async function() {
      expect(totalSupply).to.equal(await btcToken.totalSupply() + tokenAmount);
    });
  });

  describe("contractBridgeOutBTCFailureNoAllowance", function() {
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
        tx = await bridgeOutDelegate.connect(senderSigner).bridgeOutBTCFailureNoAllowance(
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

  describe("contractBridgeOutBTCFailureNoBalance", function() {
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
        tx = await bridgeOutDelegate.connect(senderSigner).bridgeOutBTCFailureNoBalance(
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

  describe("contractBridgeOutBTCSuccess", function() {
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

      tx = await bridgeOutDelegate.connect(senderSigner).bridgeOutBTCSuccess(
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

  describe("bridgeOutERC20FailureNoAllowance", function() {
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

  describe("bridgeOutERC20FailureNoBalance", function() {
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

      // approve twice, the amount that we are going to try to deposit
      tx = await simpleToken.connect(senderSigner)
        .approve(assetsBridgePrecompileAddress, tokenAmount * 2);
      await tx.wait();

      try {
        await assetsBridge.connect(senderSigner).bridgeOut(
          await simpleToken.getAddress(),
          tokenAmount * 2,
          0,
          Buffer.from(recipient, "hex"),
        );
      } catch (error: any) {
        errorMessage = error.message;
      }

      try {
        tx = await assetsBridge.connect(senderSigner).bridgeOut(
          await simpleToken.getAddress(),
          tokenAmount * 2,
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

  describe("bridgeOutERC20Success", function() {
    let receipt1: any;
    let receipt2: any;
    let receipt3: any;
    let receipt4: any;
    let tx: any;
    let tokenAmount: any;
    let mezoTokenAmount: any;
    let recipient = "150bCF49Ee8E2Bd9f59e991821DE5B74C6D876aA";
    let sourceTokenAddress = ethers.Wallet.createRandom().address;
    let initialSenderBalance: any;
    let totalSupply = 0;
    let gasCost = 0;

    before(async function() {
      await fixture();
      tokenAmount = ethers.parseEther("10");
      mezoTokenAmount = ethers.parseEther("42");
      initialSenderBalance = await ethers.provider.getBalance(senderAddress);

      // do the erc20 token mapping
      let tx = await assetsBridge.connect(poolOwner).createERC20TokenMapping(
        sourceTokenAddress,
        await simpleToken.getAddress(),
      );
      await tx.wait();

      // mint some token to ourselves
      // approve for token amount
      tx = await simpleToken.connect(senderSigner)
        .mint(senderAddress, tokenAmount * 2n); // mint extra supply to compare before and after
      await tx.wait();
      receipt1 = await ethers.provider.getTransactionReceipt(tx.hash);
      gasCost = receipt1.gasUsed * receipt1.gasPrice;

      totalSupply = await simpleToken.totalSupply();

      // approve for token amount
      tx = await simpleToken.connect(senderSigner)
        .approve(assetsBridgePrecompileAddress, tokenAmount);
      await tx.wait();
      receipt2 = await ethers.provider.getTransactionReceipt(tx.hash);
      gasCost = gasCost + receipt2.gasUsed * receipt2.gasPrice;

      // approve for some mezo token amount
      tx = await mezoToken.connect(senderSigner)
        .approve(assetsBridgePrecompileAddress, mezoTokenAmount);
      await tx.wait();
      receipt3 = await ethers.provider.getTransactionReceipt(tx.hash);
      gasCost = gasCost + receipt3.gasUsed * receipt3.gasPrice;

      tx = await assetsBridge.connect(senderSigner).bridgeOut(
        await simpleToken.getAddress(),
        tokenAmount,
        0,
        Buffer.from(recipient, "hex"),
      );
      await tx.wait();

      receipt4 = await ethers.provider.getTransactionReceipt(tx.hash);
      gasCost = gasCost + receipt4.gasUsed * receipt4.gasPrice;
    });

    after(async function() {
      let tx = await assetsBridge.connect(poolOwner).deleteERC20TokenMapping(
        sourceTokenAddress,
      );
      await tx.wait();
    });

    it("should verify the transaction didn't revert", async function() {
      expect(receipt1!.status).to.equal(1);
      expect(receipt2!.status).to.equal(1);
      expect(receipt3!.status).to.equal(1);
      expect(receipt4!.status).to.equal(1);
    });

    it("should verify the new erc20 balances", async function() {
      var updatedSenderBalance = await simpleToken.balanceOf(senderAddress);
      expect(updatedSenderBalance).to.equal(tokenAmount); // half left
    });

    it("should verify the new btc balances", async function() {
      var updatedSenderBalance = await ethers.provider.getBalance(senderAddress);
      expect(updatedSenderBalance).to.equal(initialSenderBalance - gasCost);
    });

    it("should verify the totalSupply", async function() {
      var updatedSupply = await simpleToken.totalSupply();
      expect(updatedSupply).to.equal(totalSupply / 2n); // half left too
    });
    it("should verify the remaining BTC and MEZO approvals", async function() {
      expect(await mezoToken.allowance(senderAddress, assetsBridgePrecompileAddress)).to.equal(ethers.parseEther("42"));
      expect(await simpleToken.allowance(senderAddress, assetsBridgePrecompileAddress)).to.equal(0);
    });
  });

  describe("contractBridgeOutERC20FailureNoAllowance", function() {
    let receipt: any;
    let tx: any;
    let tokenAmount: any;
    let initialSenderBalance: any;
    let initialContractBalance: any;
    let recipient = "1976a91462e907b15cbf27d5425399ebf6f0fb50ebb88f1888ac";
    let gasCost = 0;
    let sourceTokenAddress = ethers.Wallet.createRandom().address;

    before(async function() {
      await fixture();
      tokenAmount = ethers.parseEther("8");

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);

      // do the erc20 token mapping
      let tx = await assetsBridge.connect(poolOwner).createERC20TokenMapping(
        sourceTokenAddress,
        await simpleToken.getAddress(),
      );
      await tx.wait();

      // mint some token to ourselves
      // approve for token amount
      tx = await simpleToken.connect(senderSigner)
        .mint(senderAddress, tokenAmount);
      await tx.wait();
      receipt = await ethers.provider.getTransactionReceipt(tx.hash);
      gasCost = receipt.gasUsed * receipt.gasPrice;

      // send the funds to the contract first
      const transferTx = await simpleToken.connect(senderSigner).transfer(contractAddress, tokenAmount);
      await transferTx.wait();
      receipt = await ethers.provider.getTransactionReceipt(transferTx.hash);

      gasCost = gasCost + receipt.gasUsed * receipt.gasPrice;

      initialContractBalance = await ethers.provider.getBalance(contractAddress);

      try {
        tx = await bridgeOutDelegate.connect(senderSigner).bridgeOutERC20FailureNoBalance(
          Buffer.from(recipient, "hex"),
          tokenAmount,
          simpleToken.getAddress(),
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

    after(async function() {
      let tx = await assetsBridge.connect(poolOwner).deleteERC20TokenMapping(
        sourceTokenAddress,
      );
      await tx.wait();
    });

    it("should verify the transaction did revert", async function() {
      expect(receipt!.status).to.equal(0);
    });

    it("should verify the new balances", async function() {
      var updatedSenderBalance = await ethers.provider.getBalance(senderAddress);
      expect(updatedSenderBalance).to.equal(initialSenderBalance - gasCost);
    });

    it("should verify that contract address balances haven't changed", async function() {
      expect(await simpleToken.balanceOf(contractAddress))
        .to.equal(ethers.parseEther("8"));
    });
  });

  describe("contractBridgeOutERC20FailureNoBalance", function() {
    let receipt: any;
    let tx: any;
    let tokenAmount: any;
    let initialSenderBalance: any;
    let initialContractBalance: any;
    let recipient = "1976a91462e907b15cbf27d5425399ebf6f0fb50ebb88f1888ac";
    let gasCost = 0;
    let sourceTokenAddress = ethers.Wallet.createRandom().address;

    before(async function() {
      await fixture();
      tokenAmount = ethers.parseEther("8");

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);

      // do the erc20 token mapping
      let tx = await assetsBridge.connect(poolOwner).createERC20TokenMapping(
        sourceTokenAddress,
        await simpleToken.getAddress(),
      );
      await tx.wait();

      // mint some token to ourselves
      // approve for token amount
      tx = await simpleToken.connect(senderSigner)
        .mint(senderAddress, tokenAmount);
      await tx.wait();
      receipt = await ethers.provider.getTransactionReceipt(tx.hash);
      gasCost = receipt.gasUsed * receipt.gasPrice;

      // send the funds to the contract first
      const transferTx = await simpleToken.connect(senderSigner).transfer(contractAddress, tokenAmount);
      await transferTx.wait();
      receipt = await ethers.provider.getTransactionReceipt(transferTx.hash);

      gasCost = gasCost + receipt.gasUsed * receipt.gasPrice;

      initialContractBalance = await ethers.provider.getBalance(contractAddress);

      try {
        tx = await bridgeOutDelegate.connect(senderSigner).bridgeOutERC20FailureNoBalance(
          Buffer.from(recipient, "hex"),
          tokenAmount,
          simpleToken.getAddress(),
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

    after(async function() {
      let tx = await assetsBridge.connect(poolOwner).deleteERC20TokenMapping(
        sourceTokenAddress,
      );
      await tx.wait();
    });

    it("should verify the transaction did revert", async function() {
      expect(receipt!.status).to.equal(0);
    });

    it("should verify the new balances", async function() {
      var updatedSenderBalance = await ethers.provider.getBalance(senderAddress);
      expect(updatedSenderBalance).to.equal(initialSenderBalance - gasCost);
    });

    it("should verify that contract address balances haven't changed", async function() {
      expect(await simpleToken.balanceOf(contractAddress))
        .to.equal(ethers.parseEther("8"));
    });
  });

  describe("contractBridgeOutERC20Success", function() {
    let receipt: any;
    let tx: any;
    let tokenAmount: any;
    let initialSenderBalance: any;
    let initialContractBalance: any;
    let recipient = "150bCF49Ee8E2Bd9f59e991821DE5B74C6D876aA";
    let gasCost = 0;
    let sourceTokenAddress = ethers.Wallet.createRandom().address;

    before(async function() {
      await fixture();
      tokenAmount = ethers.parseEther("8");

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);

      // do the erc20 token mapping
      let tx = await assetsBridge.connect(poolOwner).createERC20TokenMapping(
        sourceTokenAddress,
        await simpleToken.getAddress(),
      );
      await tx.wait();

      // mint some token to ourselves
      tx = await simpleToken.connect(senderSigner)
        .mint(senderAddress, tokenAmount);
      await tx.wait();
      receipt = await ethers.provider.getTransactionReceipt(tx.hash);
      gasCost = receipt.gasUsed * receipt.gasPrice;

      // send the funds to the contract first
      const transferTx = await simpleToken.connect(senderSigner).transfer(contractAddress, tokenAmount);
      await transferTx.wait();
      receipt = await ethers.provider.getTransactionReceipt(transferTx.hash);

      gasCost = gasCost + receipt.gasUsed * receipt.gasPrice;

      initialContractBalance = await ethers.provider.getBalance(contractAddress);

      try {
        tx = await bridgeOutDelegate.connect(senderSigner).bridgeOutERC20Success(
          Buffer.from(recipient, "hex"),
          tokenAmount,
          simpleToken.getAddress(),
          { gasLimit: 10000000 },
        );
        await tx.wait();
      } catch (error: any) {
        console.log(error);
        expect(error.shortMessage).to.include(
          "execution reverted",
        );
      }

      receipt = await ethers.provider.getTransactionReceipt(tx.hash);
      gasCost = gasCost + receipt.gasUsed * receipt.gasPrice;
    });

    after(async function() {
      let tx = await assetsBridge.connect(poolOwner).deleteERC20TokenMapping(
        sourceTokenAddress,
      );
      await tx.wait();
    });

    it("should verify the transaction did not revert", async function() {
      expect(receipt!.status).to.equal(1);
    });

    it("should verify the new balances", async function() {
      var updatedSenderBalance = await ethers.provider.getBalance(senderAddress);
      expect(updatedSenderBalance).to.equal(initialSenderBalance - gasCost);
    });

    it("should verify that contract address balances have been updated", async function() {
      expect(await simpleToken.balanceOf(contractAddress))
        .to.equal(ethers.parseEther("0"));
    });
  });

  describe("contractBridgeOutBTCReverts", function() {
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
        tx = await bridgeOutDelegate.connect(senderSigner).bridgeOutBTCReverts(
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

  describe("contractBridgeOutERC20Reverts", function() {
    let receipt: any;
    let tx: any;
    let tokenAmount: any;
    let initialSenderBalance: any;
    let initialContractBalance: any;
    let recipient = "1976a91462e907b15cbf27d5425399ebf6f0fb50ebb88f1888ac";
    let gasCost = 0;
    let sourceTokenAddress = ethers.Wallet.createRandom().address;

    before(async function() {
      await fixture();
      tokenAmount = ethers.parseEther("8");

      initialSenderBalance = await ethers.provider.getBalance(senderAddress);

      // do the erc20 token mapping
      let tx = await assetsBridge.connect(poolOwner).createERC20TokenMapping(
        sourceTokenAddress,
        await simpleToken.getAddress(),
      );
      await tx.wait();

      // mint some token to ourselves
      tx = await simpleToken.connect(senderSigner)
        .mint(senderAddress, tokenAmount);
      await tx.wait();
      receipt = await ethers.provider.getTransactionReceipt(tx.hash);
      gasCost = receipt.gasUsed * receipt.gasPrice;

      // send the funds to the contract first
      const transferTx = await simpleToken.connect(senderSigner).transfer(contractAddress, tokenAmount);
      await transferTx.wait();
      receipt = await ethers.provider.getTransactionReceipt(transferTx.hash);

      gasCost = gasCost + receipt.gasUsed * receipt.gasPrice;

      initialContractBalance = await ethers.provider.getBalance(contractAddress);

      try {
        tx = await bridgeOutDelegate.connect(senderSigner).bridgeOutERC20Reverts(
          Buffer.from(recipient, "hex"),
          tokenAmount,
          simpleToken.getAddress(),
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

    after(async function() {
      let tx = await assetsBridge.connect(poolOwner).deleteERC20TokenMapping(
        sourceTokenAddress,
      );
      await tx.wait();
    });

    it("should verify the transaction did revert", async function() {
      expect(receipt!.status).to.equal(0);
    });

    it("should verify the new balances", async function() {
      var updatedSenderBalance = await ethers.provider.getBalance(senderAddress);
      expect(updatedSenderBalance).to.equal(initialSenderBalance - gasCost);
    });

    it("should verify that contract address balances haven't changed", async function() {
      expect(await simpleToken.balanceOf(contractAddress))
        .to.equal(ethers.parseEther("8"));
    });
  });

  // The bridge-out chain set holds the target chains enabled for bridge-outs.
  // The assets bridge precompile exposes the owner-only methods that remove a
  // chain from the set and add it back. A bridge-out to a chain outside the
  // set reverts in the bridge keeper, so the gate covers direct and indirect
  // bridge-outs alike.
  describe("bridgeOutChains", function() {
    let maintenance: any;
    let emergencyTeam: any;
    let thirdParty: any;
    let simpleTokenAddress: string;

    // Target chains of the bridgeOut method.
    const ethereumChain = 0;
    const bitcoinChain = 1;

    // A chain outside the target chain enum.
    const unsupportedChain = 5;

    // A bridge-out to Ethereum takes a 20-byte recipient address. A bridge-out
    // to Bitcoin takes a var-len recipient script.
    const ethereumRecipient = "150bCF49Ee8E2Bd9f59e991821DE5B74C6D876aA";
    const bitcoinRecipient = "1976a91462e907b15cbf27d5425399ebf6f0fb50ebb88f1888ac";

    // The Ethereum-side token of the ERC20 mapping the bridge-out to Ethereum
    // needs. A fresh address per run keeps the mapping free of leftovers.
    const sourceTokenAddress = ethers.Wallet.createRandom().address;

    const btcAmount = ethers.parseEther("0.5");

    // The ERC20 amount stays above the minimum bridge-out amount that the
    // AssetsBridgeIndirectBridgeOut suite sets for the same token and never
    // clears, so the suites can run in any order.
    const erc20Amount = ethers.parseEther("10");

    // eventArguments returns the arguments of the named event emitted by the
    // transaction. It throws when the transaction does not emit that event.
    const eventArguments = (receipt: any, name: string) => {
      const iface = new ethers.Interface(assetsbridgeabi);

      for (const log of receipt.logs) {
        try {
          const parsed = iface.parseLog({
            topics: log.topics as string[],
            data: log.data,
          });
          if (parsed && parsed.name === name) {
            return parsed.args;
          }
        } catch {
          // Not a decodable event from our ABI, skip.
        }
      }

      throw new Error(`the transaction did not emit the ${name} event`);
    };

    // bridgeOutChains returns the chain set as plain numbers.
    const bridgeOutChains = async () =>
      Array.from(await assetsBridge.getBridgeOutChains(), (chain: any) =>
        Number(chain),
      );

    // callFailure returns the error message of a failed call. It throws when
    // the call succeeds.
    const callFailure = async (call: () => Promise<any>) => {
      try {
        await call();
      } catch (error: any) {
        return extractMessage(error);
      }

      throw new Error("the call did not fail");
    };

    const btcBridgeOutArgs = (chain: number) => [
      btcTokenPrecompileAddress,
      btcAmount,
      chain,
      Buffer.from(
        chain === bitcoinChain ? bitcoinRecipient : ethereumRecipient,
        "hex",
      ),
    ];

    // sendBTCBridgeOut approves the assets bridge and sends a native BTC
    // bridge-out to the given target chain.
    const sendBTCBridgeOut = async (chain: number) => {
      await (
        await btcToken
          .connect(senderSigner)
          .approve(assetsBridgePrecompileAddress, btcAmount)
      ).wait();

      const tx = await assetsBridge
        .connect(senderSigner)
        .bridgeOut(...btcBridgeOutArgs(chain));
      return tx.wait();
    };

    // simulateBTCBridgeOut runs the same bridge-out as a call. The call
    // surfaces the revert reason of the precompile, which a sent transaction
    // does not.
    const simulateBTCBridgeOut = (chain: number) =>
      assetsBridge
        .connect(senderSigner)
        .bridgeOut.staticCall(...btcBridgeOutArgs(chain));

    before(async function() {
      await fixture();

      maintenance = new hre.ethers.Contract(
        maintenancePrecompileAddress,
        maintenanceabi,
        ethers.provider,
      );

      emergencyTeam = signers[1];
      thirdParty = signers[2];
      simpleTokenAddress = await simpleToken.getAddress();

      // A bridge-out to Ethereum needs the ERC20 mapping of the bridged token.
      let tx = await assetsBridge.connect(poolOwner).createERC20TokenMapping(
        sourceTokenAddress,
        simpleTokenAddress,
      );
      await tx.wait();

      tx = await simpleToken.connect(senderSigner)
        .mint(senderAddress, erc20Amount);
      await tx.wait();
    });

    // Leave both chains enabled, the mapping deleted and the chain without an
    // Emergency Team for the other tests.
    after(async function() {
      if (!(await bridgeOutChains()).includes(bitcoinChain)) {
        await (
          await assetsBridge.connect(poolOwner).addBridgeOutChain(bitcoinChain)
        ).wait();
      }

      await (
        await assetsBridge.connect(poolOwner).deleteERC20TokenMapping(
          sourceTokenAddress,
        )
      ).wait();
      await (
        await maintenance.connect(poolOwner).setEmergencyTeam(
          ethers.ZeroAddress,
        )
      ).wait();
    });

    it("reports both chains by default and bridges out to both", async function() {
      expect(await bridgeOutChains()).to.deep.equal([
        ethereumChain,
        bitcoinChain,
      ]);

      expect((await sendBTCBridgeOut(bitcoinChain)).status).to.equal(1);
      expect((await sendBTCBridgeOut(ethereumChain)).status).to.equal(1);
    });

    it("rejects the chain set methods from a third party", async function() {
      expect(
        await callFailure(() =>
          assetsBridge
            .connect(thirdParty)
            .removeBridgeOutChain.staticCall(bitcoinChain),
        ),
      ).to.include("not the owner");

      expect(
        await callFailure(() =>
          assetsBridge
            .connect(thirdParty)
            .addBridgeOutChain.staticCall(bitcoinChain),
        ),
      ).to.include("not the owner");

      expect(await bridgeOutChains()).to.deep.equal([
        ethereumChain,
        bitcoinChain,
      ]);
    });

    it("rejects the chain set methods from the emergency team", async function() {
      // The chain set retires a bridging venue, so it stays owner-only. The
      // Emergency Team drives the bridge lockdown instead.
      await (
        await maintenance.connect(poolOwner).setEmergencyTeam(
          emergencyTeam.address,
        )
      ).wait();

      expect(
        await callFailure(() =>
          assetsBridge
            .connect(emergencyTeam)
            .removeBridgeOutChain.staticCall(bitcoinChain),
        ),
      ).to.include("not the owner");

      expect(
        await callFailure(() =>
          assetsBridge
            .connect(emergencyTeam)
            .addBridgeOutChain.staticCall(bitcoinChain),
        ),
      ).to.include("not the owner");

      expect(await bridgeOutChains()).to.deep.equal([
        ethereumChain,
        bitcoinChain,
      ]);
    });

    it("rejects a chain outside the target chain enum", async function() {
      expect(
        await callFailure(() =>
          assetsBridge
            .connect(poolOwner)
            .removeBridgeOutChain.staticCall(unsupportedChain),
        ),
      ).to.include("unsupported chain");

      expect(
        await callFailure(() =>
          assetsBridge
            .connect(poolOwner)
            .addBridgeOutChain.staticCall(unsupportedChain),
        ),
      ).to.include("unsupported chain");
    });

    it("rejects adding a chain that is already in the set", async function() {
      expect(
        await callFailure(() =>
          assetsBridge
            .connect(poolOwner)
            .addBridgeOutChain.staticCall(bitcoinChain),
        ),
      ).to.include("chain is already enabled for bridge-outs");
    });

    it("lets the owner remove the Bitcoin chain", async function() {
      const receipt = await (
        await assetsBridge.connect(poolOwner).removeBridgeOutChain(bitcoinChain)
      ).wait();

      expect(receipt.status).to.equal(1);

      const args = eventArguments(receipt, "BridgeOutChainRemoved");
      expect(Number(args.chain)).to.equal(bitcoinChain);

      expect(await bridgeOutChains()).to.deep.equal([ethereumChain]);
    });

    it("blocks only the bridge-out to Bitcoin", async function() {
      expect(
        await callFailure(() => simulateBTCBridgeOut(bitcoinChain)),
      ).to.include("target chain is not enabled for bridge-outs");

      expect((await sendBTCBridgeOut(ethereumChain)).status).to.equal(1);

      await (
        await simpleToken
          .connect(senderSigner)
          .approve(assetsBridgePrecompileAddress, erc20Amount)
      ).wait();

      const receipt = await (
        await assetsBridge.connect(senderSigner).bridgeOut(
          simpleTokenAddress,
          erc20Amount,
          ethereumChain,
          Buffer.from(ethereumRecipient, "hex"),
        )
      ).wait();

      expect(receipt.status).to.equal(1);
    });

    it("rejects removing the Bitcoin chain again", async function() {
      expect(
        await callFailure(() =>
          assetsBridge
            .connect(poolOwner)
            .removeBridgeOutChain.staticCall(bitcoinChain),
        ),
      ).to.include("chain is not enabled for bridge-outs");

      expect(await bridgeOutChains()).to.deep.equal([ethereumChain]);
    });

    it("blocks the indirect bridge-out to Bitcoin", async function() {
      // A contract-mediated bridge-out fails without a revert reason, so the
      // receipt status and the untouched balance carry the assertion.
      await (
        await btcToken
          .connect(senderSigner)
          .transfer(contractAddress, btcAmount)
      ).wait();

      const balanceBefore = await btcToken.balanceOf(contractAddress);

      // The explicit gas limit skips the gas estimation, so the transaction
      // reaches the chain and leaves a receipt to assert on.
      let tx: any;
      try {
        tx = await bridgeOutDelegate
          .connect(senderSigner)
          .bridgeOutBTCSuccess(Buffer.from(bitcoinRecipient, "hex"), btcAmount, {
            gasLimit: 1000000,
          });
        await tx.wait();
      } catch (error: any) {
        expect(extractMessage(error)).to.include("reverted");
      }

      const receipt = await ethers.provider.getTransactionReceipt(tx.hash);
      expect(receipt!.status).to.equal(0);
      expect(await btcToken.balanceOf(contractAddress)).to.equal(
        balanceBefore,
      );
    });

    it("lets the owner add the Bitcoin chain back", async function() {
      const receipt = await (
        await assetsBridge.connect(poolOwner).addBridgeOutChain(bitcoinChain)
      ).wait();

      expect(receipt.status).to.equal(1);

      const args = eventArguments(receipt, "BridgeOutChainAdded");
      expect(Number(args.chain)).to.equal(bitcoinChain);

      expect(await bridgeOutChains()).to.deep.equal([
        ethereumChain,
        bitcoinChain,
      ]);

      expect((await sendBTCBridgeOut(bitcoinChain)).status).to.equal(1);
    });
  });
});
