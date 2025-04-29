import { expect } from "chai"
import { ethers } from "hardhat"
import hre from "hardhat"
import { loadFixture } from "@nomicfoundation/hardhat-network-helpers"
import { getDeployedContract } from "./helpers/contract";

describe("TestERC20", function () {
  const { deployments } = hre;
  let testERC20: any;
  let owner: any;
  let minter: any;
  let user1: any;
  let user2: any;

  async function fixture() {
    await deployments.fixture();
    const signers = await ethers.getSigners();
    owner = signers[0];
    minter = signers[1];
    user1 = signers[2];
    user2 = signers[3];

    testERC20 = await getDeployedContract("TestERC20");

    return { testERC20, owner, minter, user1, user2 }
  }

  describe("Initialization", function () {
    it("should set the correct name and symbol", async function () {
      const { testERC20 } = await loadFixture(fixture)
      expect(await testERC20.name()).to.equal("TestERC20")
      expect(await testERC20.symbol()).to.equal("TEST42")
    })

    it("should set the correct minter", async function () {
      const { testERC20, minter } = await loadFixture(fixture)
      expect(await testERC20.minter()).to.equal(minter)
    })

    it("should set the correct owner", async function () {
      const { testERC20, owner } = await loadFixture(fixture)
      expect(await testERC20.owner()).to.equal(owner.address)
    })
  })

  describe("Minting", function () {
    it("should allow minter to mint tokens", async function () {
      const { testERC20, minter, user1 } = await loadFixture(fixture)
      const amount = ethers.parseEther("100")
      
      await testERC20.connect(minter).mint(user1.address, amount)
      expect(await testERC20.balanceOf(user1.address)).to.equal(amount)
    })

    it("should not allow non-minter to mint tokens", async function () {
      const { testERC20, user1, user2 } = await loadFixture(fixture)
      const amount = ethers.parseEther("100")
      
      await expect(
        testERC20.connect(user1).mint(user2.address, amount)
      ).to.be.revertedWithCustomError(testERC20, "NotMinter")
    })
  })

  describe("Burning", function () {
    it("should allow token holder to burn their own tokens", async function () {
      const { testERC20, minter, user1 } = await loadFixture(fixture)
      const amount = ethers.parseEther("100")
      
      await testERC20.connect(minter).mint(user1.address, amount)
      await testERC20.connect(user1).burn(amount)
      
      expect(await testERC20.balanceOf(user1.address)).to.equal(0)
    })

    it("should allow token holder to burn tokens from another address with allowance", async function () {
      const { testERC20, minter, user1, user2 } = await loadFixture(fixture)
      const amount = ethers.parseEther("100")
      
      await testERC20.connect(minter).mint(user1.address, amount)
      await testERC20.connect(user1).approve(user2.address, amount)
      await testERC20.connect(user2).burnFrom(user1.address, amount)
      
      expect(await testERC20.balanceOf(user1.address)).to.equal(0)
    })

    it("should not allow burning more tokens than balance", async function () {
      const { testERC20, minter, user1 } = await loadFixture(fixture)
      const amount = ethers.parseEther("100")
      
      await testERC20.connect(minter).mint(user1.address, amount)
      
      await expect(
        testERC20.connect(user1).burn(amount + 1n)
      ).to.be.reverted
    })
  })

  describe("Transfer", function () {
    it("should allow token holder to transfer tokens", async function () {
      const { testERC20, minter, user1, user2 } = await loadFixture(fixture)
      const amount = ethers.parseEther("100")
      
      await testERC20.connect(minter).mint(user1.address, amount)
      await testERC20.connect(user1).transfer(user2.address, amount)
      
      expect(await testERC20.balanceOf(user1.address)).to.equal(0)
      expect(await testERC20.balanceOf(user2.address)).to.equal(amount)
    })

    it("should not allow transferring more tokens than balance", async function () {
      const { testERC20, minter, user1, user2 } = await loadFixture(fixture)
      const amount = ethers.parseEther("100")
      
      await testERC20.connect(minter).mint(user1.address, amount)
      
      await expect(
        testERC20.connect(user1).transfer(user2.address, amount + 1n)
      ).to.be.reverted
    })
  })

  describe("Minter Management", function () {
    it("should allow owner to change minter", async function () {
      const { testERC20, owner, user1 } = await loadFixture(fixture)
      
      await testERC20.connect(owner).setMinter(user1.address)
      expect(await testERC20.minter()).to.equal(user1.address)
    })

    it("should not allow non-owner to change minter", async function () {
      const { testERC20, user1 } = await loadFixture(fixture)
      
      await expect(
        testERC20.connect(user1).setMinter(user1.address)
      ).to.be.reverted
    })

    it("should not allow setting minter to zero address", async function () {
      const { testERC20, owner } = await loadFixture(fixture)
      
      await expect(
        testERC20.connect(owner).setMinter(ethers.ZeroAddress)
      ).to.be.revertedWithCustomError(testERC20, "ZeroAddressMinter")
    })
  })

  describe("Ownership Transfer", function () {
    it("should allow owner to transfer contract ownership", async function () {
      const { testERC20, owner, user1 } = await loadFixture(fixture)
      
      await testERC20.connect(owner).transferOwnership(user1.address)
      await testERC20.connect(user1).acceptOwnership()
      expect(await testERC20.owner()).to.equal(user1.address)
    })

    it("should not allow non-owner to transfer contract ownership", async function () {
      const { testERC20, user1 } = await loadFixture(fixture)
      
      await expect(
        testERC20.connect(user1).transferOwnership(user1.address)
      ).to.be.reverted
    })
  })
}) 