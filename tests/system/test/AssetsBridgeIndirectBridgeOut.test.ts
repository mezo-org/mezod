import { expect } from "chai";
import hre from "hardhat";
import { ethers } from "hardhat";
import assetsbridgeabi from "../../../precompile/assetsbridge/abi.json";
import btcabi from "../../../precompile/btctoken/abi.json";
import validatorpoolabi from "../../../precompile/validatorpool/abi.json";
import { BridgeOutCaller, SimpleToken } from "../typechain-types";
import { getDeployedContract } from "./helpers/contract";

const validatorPoolPrecompileAddress = "0x7b7c000000000000000000000000000000000011";
const assetsBridgePrecompileAddress = "0x7b7c000000000000000000000000000000000012";
const btcTokenPrecompileAddress = "0x7b7c000000000000000000000000000000000000";

describe("AssetsBridgeIndirectBridgeOut", function () {
  const { deployments } = hre;
  let assetsBridge: any;
  let btcToken: any;
  let validatorPool: any;
  let bridgeOutCaller: BridgeOutCaller;
  let simpleToken: SimpleToken;
  let signers: any;
  let senderSigner: any;
  let poolOwner: any;
  let senderAddress: string;
  let bridgeOutCallerAddr: string;
  let bridgeOutCallerOwner: string;

  const fixture = async function () {
    await deployments.fixture(["BridgeOutDelegate", "IndirectBridgeOut"]);
    validatorPool = new hre.ethers.Contract(
      validatorPoolPrecompileAddress,
      validatorpoolabi,
      ethers.provider
    );
    assetsBridge = new hre.ethers.Contract(
      assetsBridgePrecompileAddress,
      assetsbridgeabi,
      ethers.provider
    );
    btcToken = new hre.ethers.Contract(
      btcTokenPrecompileAddress,
      btcabi,
      ethers.provider
    );
    bridgeOutCaller = await getDeployedContract("BridgeOutCaller");
    simpleToken = await getDeployedContract("SimpleToken");
    signers = await ethers.getSigners();
    bridgeOutCallerAddr = await bridgeOutCaller.getAddress();
    bridgeOutCallerOwner = await bridgeOutCaller.owner();
    poolOwner = await ethers.getSigner(await validatorPool.owner());

    senderSigner = ethers.Wallet.createRandom().connect(ethers.provider);
    senderAddress = senderSigner.address;

    const fundingSigner = signers[0];
    const transferTx = await btcToken
      .connect(fundingSigner)
      .transfer(senderSigner, ethers.parseEther("10"));
    await transferTx.wait();
  };

  describe("contractIndirectBridgeOutERC20", function () {
    let receipt: any;
    let tx: any;
    let tokenAmount: any;
    let initialSupply: any;
    const sourceTokenAddress = "0x4a8f8d8e4d2b7c6a9f3e1d5c0b2a7e6f1c9d4b8a";
    const recipient = "150bCF49Ee8E2Bd9f59e991821DE5B74C6D876aA";

    before(async function () {
      await fixture();
      tokenAmount = ethers.parseEther("8");

      tx = await assetsBridge
        .connect(poolOwner)
        .createERC20TokenMapping(
          sourceTokenAddress,
          await simpleToken.getAddress()
        );
      await tx.wait();

      tx = await assetsBridge
        .connect(poolOwner)
        .setOutflowLimit(await simpleToken.getAddress(), ethers.MaxUint256);
      await tx.wait();

      tx = await assetsBridge
        .connect(poolOwner)
        .setMinBridgeOutAmount(await simpleToken.getAddress(), tokenAmount);
      await tx.wait();

      tx = await simpleToken
        .connect(senderSigner)
        .mint(senderAddress, tokenAmount);
      await tx.wait();

      initialSupply = await simpleToken.totalSupply();

      tx = await simpleToken
        .connect(senderSigner)
        .approve(bridgeOutCallerAddr, tokenAmount);
      await tx.wait();

      tx = await bridgeOutCaller
        .connect(senderSigner)
        .execute(
          await simpleToken.getAddress(),
          tokenAmount,
          Buffer.from(recipient, "hex")
        );
      receipt = await tx.wait();
    });

    after(async function () {
      let tx = await assetsBridge
        .connect(poolOwner)
        .deleteERC20TokenMapping(sourceTokenAddress);
      await tx.wait();
    });

    it("should succeed", async function () {
      expect(receipt!.status).to.equal(1);
    });

    it("should set the indirect bridge-out caller balance to zero", async function () {
      expect(await simpleToken.balanceOf(bridgeOutCallerAddr)).to.equal(0);
    });

    it("should not transfer tokens from the indirect bridge-out caller to its owner", async function () {
      expect(await simpleToken.balanceOf(bridgeOutCallerOwner)).to.equal(0);
    });

    it("should set the indirect bridge-out caller allowance to zero", async function () {
      expect(
        await simpleToken.allowance(
          bridgeOutCallerAddr,
          assetsBridgePrecompileAddress
        )
      ).to.equal(0);
    });

    it("should reduce totalSupply by the bridged amount", async function () {
      expect(await simpleToken.totalSupply()).to.equal(
        initialSupply - tokenAmount
      );
    });
  });
});
