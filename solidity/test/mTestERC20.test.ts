import { expect } from "chai"
import { ethers } from "hardhat"
import hre from "hardhat"
import { loadFixture } from "@nomicfoundation/hardhat-network-helpers"
import { getDeployedContract } from "./helpers/contract";

describe("mTestERC20", function () {
  const { deployments } = hre;
  let mTestERC20: any;
  let owner: any;
  let governance: any;
  let minter: any;
  let user1: any;
  let user2: any;

  async function fixture() {
    await deployments.fixture();
    const signers = await ethers.getSigners();
    owner = signers[0];
    governance = signers[1];
    minter = signers[2];
    user1 = signers[3];
    user2 = signers[4];

    mTestERC20 = await getDeployedContract("mTestERC20");

    return { mTestERC20, owner, governance, minter, user1, user2 }
  }

  describe("Initialization", function () {
    it("should set the correct name and symbol", async function () {
      const { mTestERC20 } = await loadFixture(fixture)
      expect(await mTestERC20.name()).to.equal("Test Token")
      expect(await mTestERC20.symbol()).to.equal("TEST")
    })

    it("should set the correct minter", async function () {
      const { mTestERC20, minter } = await loadFixture(fixture)
      expect(await mTestERC20.minter()).to.equal(minter)
    })

    it("should set the correct owner", async function () {
      const { mTestERC20, owner } = await loadFixture(fixture)
      expect(await mTestERC20.owner()).to.equal(owner.address)
    })
  })

  describe("Minting", function () {
    it("should allow minter to mint tokens", async function () {
      const { mTestERC20, minter, user1 } = await loadFixture(fixture)
      const amount = ethers.parseEther("100")
      
      await mTestERC20.connect(minter).mint(user1.address, amount)
      expect(await mTestERC20.balanceOf(user1.address)).to.equal(amount)
    })

    it("should not allow non-minter to mint tokens", async function () {
      const { mTestERC20, user1, user2 } = await loadFixture(fixture)
      const amount = ethers.parseEther("100")
      
      await expect(
        mTestERC20.connect(user1).mint(user2.address, amount)
      ).to.be.revertedWithCustomError(mTestERC20, "NotMinter")
    })
  })

  describe("Burning", function () {
    it("should allow token holder to burn their own tokens", async function () {
      const { mTestERC20, minter, user1 } = await loadFixture(fixture)
      const amount = ethers.parseEther("100")
      
      await mTestERC20.connect(minter).mint(user1.address, amount)
      await mTestERC20.connect(user1).burn(amount)
      
      expect(await mTestERC20.balanceOf(user1.address)).to.equal(0)
    })

    it("should allow token holder to burn tokens from another address with allowance", async function () {
      const { mTestERC20, minter, user1, user2 } = await loadFixture(fixture)
      const amount = ethers.parseEther("100")
      
      await mTestERC20.connect(minter).mint(user1.address, amount)
      await mTestERC20.connect(user1).approve(user2.address, amount)
      await mTestERC20.connect(user2).burnFrom(user1.address, amount)
      
      expect(await mTestERC20.balanceOf(user1.address)).to.equal(0)
    })

    it("should not allow burning more tokens than balance", async function () {
      const { mTestERC20, minter, user1 } = await loadFixture(fixture)
      const amount = ethers.parseEther("100")
      
      await mTestERC20.connect(minter).mint(user1.address, amount)
      
      await expect(
        mTestERC20.connect(user1).burn(amount + 1n)
      ).to.be.reverted
    })
  })

  describe("Transfer", function () {
    it("should allow token holder to transfer tokens", async function () {
      const { mTestERC20, minter, user1, user2 } = await loadFixture(fixture)
      const amount = ethers.parseEther("100")
      
      await mTestERC20.connect(minter).mint(user1.address, amount)
      await mTestERC20.connect(user1).transfer(user2.address, amount)
      
      expect(await mTestERC20.balanceOf(user1.address)).to.equal(0)
      expect(await mTestERC20.balanceOf(user2.address)).to.equal(amount)
    })

    it("should not allow transferring more tokens than balance", async function () {
      const { mTestERC20, minter, user1, user2 } = await loadFixture(fixture)
      const amount = ethers.parseEther("100")
      
      await mTestERC20.connect(minter).mint(user1.address, amount)
      
      await expect(
        mTestERC20.connect(user1).transfer(user2.address, amount + 1n)
      ).to.be.reverted
    })
  })

  describe("Minter Management", function () {
    it("should allow owner to change minter", async function () {
      const { mTestERC20, owner, user1 } = await loadFixture(fixture)
      
      await mTestERC20.connect(owner).setMinter(user1.address)
      expect(await mTestERC20.minter()).to.equal(user1.address)
    })

    it("should not allow non-owner to change minter", async function () {
      const { mTestERC20, user1 } = await loadFixture(fixture)
      
      await expect(
        mTestERC20.connect(user1).setMinter(user1.address)
      ).to.be.reverted
    })

    it("should not allow setting minter to zero address", async function () {
      const { mTestERC20, owner } = await loadFixture(fixture)
      
      await expect(
        mTestERC20.connect(owner).setMinter(ethers.ZeroAddress)
      ).to.be.revertedWithCustomError(mTestERC20, "ZeroAddressMinter")
    })
  })

  describe("Ownership Transfer", function () {
    it("should allow owner to transfer contract ownership", async function () {
      const { mTestERC20, owner, governance } = await loadFixture(fixture)
      
      await mTestERC20.connect(owner).transferOwnership(governance.address)
      await mTestERC20.connect(governance).acceptOwnership()
      expect(await mTestERC20.owner()).to.equal(governance.address)
    })

    it("should not allow non-owner to transfer contract ownership", async function () {
      const { mTestERC20, user1 } = await loadFixture(fixture)
      
      await expect(
        mTestERC20.connect(user1).transferOwnership(user1.address)
      ).to.be.reverted
    })
  })

  describe("Decimals", function () {
    it("should return the correct number of decimals", async function () {
      const { mTestERC20 } = await loadFixture(fixture)
      expect(await mTestERC20.decimals()).to.equal(8) // 8 is the decimals for the mTestERC20 token
    })
  })
}) 