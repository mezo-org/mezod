import { expect } from "chai";
import hre from "hardhat";
import { ethers } from "hardhat";
import { getDeployedContract } from "./helpers/contract";

// EVM-format address of a Cosmos module account: sha256(name)[:20].
function blockedModuleAddress(name: string): string {
  const digest = ethers.sha256(ethers.toUtf8Bytes(name));
  return ethers.getAddress("0x" + digest.slice(2, 42));
}

describe("RecipientGuard", function () {
  const { deployments } = hre;
  let forwarder: any;
  let sender: any;
  // poa is a stable module account in BankKeeper.BlockedAddrs — fee_collector
  // would also be blocked but receives gas fees via SendCoinsFromAccountToModule,
  // so its balance grows during normal block processing.
  const blockedRecipient = blockedModuleAddress("poa");

  const fixture = (async function () {
    await deployments.fixture(["RecipientGuard"]);
    forwarder = await getDeployedContract("Forwarder");
    const signers = await ethers.getSigners();
    sender = signers[0];
  });

  describe("inner CALL via Forwarder", function () {
    let forwarderAddress: string;
    let amount: bigint;

    let initialSenderBalance: bigint;
    let initialForwarderBalance: bigint;
    let initialBlockedBalance: bigint;

    let staticOk: boolean;
    let tx: any;
    let receipt: any;
    let gasCost: any;

    before(async function () {
      await fixture();

      forwarderAddress = await forwarder.getAddress();
      amount = ethers.parseEther("0.0001");

      initialSenderBalance = await ethers.provider.getBalance(sender.address);
      initialForwarderBalance = await ethers.provider.getBalance(forwarderAddress);
      initialBlockedBalance = await ethers.provider.getBalance(blockedRecipient);

      // Probe the inner CALL via staticCall before the live tx so we can
      // pin that the inner frame is the failing site (ok=false) without
      // perturbing on-chain state.
      staticOk = await forwarder.connect(sender).run.staticCall(blockedRecipient, {
        value: amount,
        gasLimit: 200000,
      });

      tx = await forwarder.connect(sender).run(blockedRecipient, {
        value: amount,
        gasLimit: 200000,
      });
      receipt = await tx.wait();
      gasCost = receipt.gasUsed * receipt.gasPrice;
    });

    it("should report ok=false from the inner CALL via staticCall", async function () {
      expect(staticOk).to.equal(false);
    });

    it("should mine the outer tx with a normal receipt", async function () {
      expect(receipt).to.not.be.null;
      expect(receipt.status).to.equal(1);
    });

    it("should make the outer tx visible to eth_getTransactionByHash", async function () {
      const fetched = await ethers.provider.getTransaction(tx.hash);
      expect(fetched).to.not.be.null;
      expect(fetched!.hash).to.equal(tx.hash);
    });

    it("should leave the blocked recipient's balance unchanged", async function () {
      const current = await ethers.provider.getBalance(blockedRecipient);
      expect(current).to.equal(initialBlockedBalance);
    });

    it("should leave the value with the forwarder (inner CALL did not succeed)", async function () {
      const current = await ethers.provider.getBalance(forwarderAddress);
      expect(current).to.equal(initialForwarderBalance + amount);
    });

    it("should debit the sender by value + gas cost", async function () {
      const current = await ethers.provider.getBalance(sender.address);
      expect(current).to.equal(initialSenderBalance - amount - gasCost);
    });
  });

  describe("EOA direct value send to blocked recipient", function () {
    let amount: bigint;
    let initialBlockedBalance: bigint;

    before(async function () {
      await fixture();
      amount = ethers.parseEther("0.0001");
      initialBlockedBalance = await ethers.provider.getBalance(blockedRecipient);
    });

    it("should be rejected at the EVM frame with a queryable failed tx", async function () {
      let receipt: any;
      let txHash: string | undefined;
      try {
        const tx = await sender.sendTransaction({
          to: blockedRecipient,
          value: amount,
          gasLimit: 100000,
        });
        txHash = tx.hash;
        receipt = await tx.wait();
      } catch (e: any) {
        receipt = e.receipt ?? null;
        txHash = txHash ?? e.transactionHash ?? e.transaction?.hash;
      }

      expect(txHash, "tx hash must be present so RPC clients can query the outcome").to.be.a("string");
      const fetched = await ethers.provider.getTransaction(txHash!);
      expect(fetched, "eth_getTransactionByHash must return the tx, not null").to.not.be.null;

      const finalBlockedBalance = await ethers.provider.getBalance(blockedRecipient);
      expect(finalBlockedBalance).to.equal(initialBlockedBalance);
    });
  });

  describe("regression: allowed recipient still receives", function () {
    let recipient: string;
    let amount: bigint;

    let initialRecipientBalance: bigint;
    let initialSenderBalance: bigint;

    let tx: any;
    let receipt: any;
    let gasCost: any;

    before(async function () {
      await fixture();
      recipient = ethers.Wallet.createRandom().address;
      amount = ethers.parseEther("0.0001");

      initialRecipientBalance = await ethers.provider.getBalance(recipient);
      initialSenderBalance = await ethers.provider.getBalance(sender.address);

      tx = await sender.sendTransaction({
        to: recipient,
        value: amount,
        gasLimit: 100000,
      });
      receipt = await tx.wait();
      gasCost = receipt.gasUsed * receipt.gasPrice;
    });

    it("should mine successfully and credit the recipient", async function () {
      expect(receipt.status).to.equal(1);
      const current = await ethers.provider.getBalance(recipient);
      expect(current).to.equal(initialRecipientBalance + amount);
    });

    it("should debit the sender by value + gas cost", async function () {
      const current = await ethers.provider.getBalance(sender.address);
      expect(current).to.equal(initialSenderBalance - amount - gasCost);
    });
  });
});
