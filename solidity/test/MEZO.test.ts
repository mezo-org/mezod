import { expect } from "chai"
import { ethers } from "hardhat"
import { loadFixture } from "@nomicfoundation/hardhat-network-helpers"

describe("MEZO", function () {
  let mezo: any
  let owner: any
  let minter1: any
  let minter2: any
  let burner1: any
  let burner2: any
  let user1: any
  let user2: any

  async function fixture() {
    const signers = await ethers.getSigners()
    owner = signers[0]
    minter1 = signers[1]
    minter2 = signers[2]
    burner1 = signers[3]
    burner2 = signers[4]
    user1 = signers[5]
    user2 = signers[6]

    const MEZO = await ethers.getContractFactory("MEZO")
    mezo = await MEZO.deploy()

    return { mezo, owner, minter1, minter2, burner1, burner2, user1, user2 }
  }

  describe("setMinter", function () {
    context("when called by owner", function () {
      it("should set minter to true", async function () {
        const { mezo, owner, minter1 } = await loadFixture(fixture)

        await mezo.connect(owner).setMinter(minter1.address, true)

        expect(await mezo.minters(minter1.address)).to.equal(true)
      })

      it("should set minter to false", async function () {
        const { mezo, owner, minter1 } = await loadFixture(fixture)

        await mezo.connect(owner).setMinter(minter1.address, true)
        await mezo.connect(owner).setMinter(minter1.address, false)

        expect(await mezo.minters(minter1.address)).to.equal(false)
      })

      it("should emit MinterSet event", async function () {
        const { mezo, owner, minter1 } = await loadFixture(fixture)

        await expect(mezo.connect(owner).setMinter(minter1.address, true))
          .to.emit(mezo, "MinterSet")
          .withArgs(minter1.address, true)
      })

      it("should allow setting multiple minters", async function () {
        const { mezo, owner, minter1, minter2 } = await loadFixture(fixture)

        await mezo.connect(owner).setMinter(minter1.address, true)
        await mezo.connect(owner).setMinter(minter2.address, true)

        expect(await mezo.minters(minter1.address)).to.equal(true)
        expect(await mezo.minters(minter2.address)).to.equal(true)
      })
    })

    context("when called by non-owner", function () {
      it("should revert", async function () {
        const { mezo, user1, minter1 } = await loadFixture(fixture)

        await expect(
          mezo.connect(user1).setMinter(minter1.address, true)
        ).to.be.reverted
      })
    })
  })

  describe("setBurner", function () {
    context("when called by owner", function () {
      it("should set burner to true", async function () {
        const { mezo, owner, burner1 } = await loadFixture(fixture)

        await mezo.connect(owner).setBurner(burner1.address, true)

        expect(await mezo.burners(burner1.address)).to.equal(true)
      })

      it("should set burner to false", async function () {
        const { mezo, owner, burner1 } = await loadFixture(fixture)

        await mezo.connect(owner).setBurner(burner1.address, true)
        await mezo.connect(owner).setBurner(burner1.address, false)

        expect(await mezo.burners(burner1.address)).to.equal(false)
      })

      it("should emit BurnerSet event", async function () {
        const { mezo, owner, burner1 } = await loadFixture(fixture)

        await expect(mezo.connect(owner).setBurner(burner1.address, true))
          .to.emit(mezo, "BurnerSet")
          .withArgs(burner1.address, true)
      })

      it("should allow setting multiple burners", async function () {
        const { mezo, owner, burner1, burner2 } = await loadFixture(fixture)

        await mezo.connect(owner).setBurner(burner1.address, true)
        await mezo.connect(owner).setBurner(burner2.address, true)

        expect(await mezo.burners(burner1.address)).to.equal(true)
        expect(await mezo.burners(burner2.address)).to.equal(true)
      })
    })

    context("when called by non-owner", function () {
      it("should revert", async function () {
        const { mezo, user1, burner1 } = await loadFixture(fixture)

        await expect(
          mezo.connect(user1).setBurner(burner1.address, true)
        ).to.be.reverted
      })
    })
  })

  describe("mint", function () {
    context("when called by authorized minter", function () {
      let tx: any

      before(async function () {
        const { mezo, owner, minter1, user1 } = await loadFixture(fixture)

        await mezo.connect(owner).setMinter(minter1.address, true)

        tx = await mezo.connect(minter1).mint(user1.address, ethers.parseEther("100"))
      })

      it("should mint tokens to the specified account", async function () {
        const { mezo, user1 } = await loadFixture(fixture)
        const { mezo: mezoWithMint, owner, minter1 } = await loadFixture(fixture)

        await mezoWithMint.connect(owner).setMinter(minter1.address, true)
        await mezoWithMint.connect(minter1).mint(user1.address, ethers.parseEther("100"))

        expect(await mezoWithMint.balanceOf(user1.address)).to.equal(ethers.parseEther("100"))
      })

      it("should increase total supply", async function () {
        const { mezo, owner, minter1, user1 } = await loadFixture(fixture)

        await mezo.connect(owner).setMinter(minter1.address, true)

        const initialSupply = await mezo.totalSupply()
        await mezo.connect(minter1).mint(user1.address, ethers.parseEther("100"))

        expect(await mezo.totalSupply()).to.equal(initialSupply + ethers.parseEther("100"))
      })

      it("should allow minting to multiple accounts", async function () {
        const { mezo, owner, minter1, user1, user2 } = await loadFixture(fixture)

        await mezo.connect(owner).setMinter(minter1.address, true)
        await mezo.connect(minter1).mint(user1.address, ethers.parseEther("100"))
        await mezo.connect(minter1).mint(user2.address, ethers.parseEther("200"))

        expect(await mezo.balanceOf(user1.address)).to.equal(ethers.parseEther("100"))
        expect(await mezo.balanceOf(user2.address)).to.equal(ethers.parseEther("200"))
      })

      it("should allow minting zero tokens", async function () {
        const { mezo, owner, minter1, user1 } = await loadFixture(fixture)

        await mezo.connect(owner).setMinter(minter1.address, true)
        await mezo.connect(minter1).mint(user1.address, 0)

        expect(await mezo.balanceOf(user1.address)).to.equal(0)
      })
    })

    context("when called by unauthorized address", function () {
      it("should revert with NotMinter error", async function () {
        const { mezo, user1, user2 } = await loadFixture(fixture)

        await expect(
          mezo.connect(user1).mint(user2.address, ethers.parseEther("100"))
        ).to.be.revertedWithCustomError(mezo, "NotMinter")
      })
    })

    context("when minter is revoked", function () {
      it("should revert with NotMinter error", async function () {
        const { mezo, owner, minter1, user1 } = await loadFixture(fixture)

        await mezo.connect(owner).setMinter(minter1.address, true)
        await mezo.connect(owner).setMinter(minter1.address, false)

        await expect(
          mezo.connect(minter1).mint(user1.address, ethers.parseEther("100"))
        ).to.be.revertedWithCustomError(mezo, "NotMinter")
      })
    })
  })

  describe("burn", function () {
    context("when called by authorized burner with sufficient balance", function () {
      it("should burn tokens from the caller", async function () {
        const { mezo, owner, minter1, burner1 } = await loadFixture(fixture)

        await mezo.connect(owner).setMinter(minter1.address, true)
        await mezo.connect(owner).setBurner(burner1.address, true)
        await mezo.connect(minter1).mint(burner1.address, ethers.parseEther("100"))

        await mezo.connect(burner1).burn(ethers.parseEther("100"))

        expect(await mezo.balanceOf(burner1.address)).to.equal(0)
      })

      it("should decrease total supply", async function () {
        const { mezo, owner, minter1, burner1 } = await loadFixture(fixture)

        await mezo.connect(owner).setMinter(minter1.address, true)
        await mezo.connect(owner).setBurner(burner1.address, true)
        await mezo.connect(minter1).mint(burner1.address, ethers.parseEther("100"))

        const initialSupply = await mezo.totalSupply()
        await mezo.connect(burner1).burn(ethers.parseEther("50"))

        expect(await mezo.totalSupply()).to.equal(initialSupply - ethers.parseEther("50"))
      })

      it("should allow burning partial balance", async function () {
        const { mezo, owner, minter1, burner1 } = await loadFixture(fixture)

        await mezo.connect(owner).setMinter(minter1.address, true)
        await mezo.connect(owner).setBurner(burner1.address, true)
        await mezo.connect(minter1).mint(burner1.address, ethers.parseEther("100"))

        await mezo.connect(burner1).burn(ethers.parseEther("30"))

        expect(await mezo.balanceOf(burner1.address)).to.equal(ethers.parseEther("70"))
      })

      it("should allow burning zero tokens", async function () {
        const { mezo, owner, minter1, burner1 } = await loadFixture(fixture)

        await mezo.connect(owner).setMinter(minter1.address, true)
        await mezo.connect(owner).setBurner(burner1.address, true)
        await mezo.connect(minter1).mint(burner1.address, ethers.parseEther("100"))

        await mezo.connect(burner1).burn(0)

        expect(await mezo.balanceOf(burner1.address)).to.equal(ethers.parseEther("100"))
      })
    })

    context("when called by unauthorized address", function () {
      it("should revert with NotBurner error", async function () {
        const { mezo, owner, minter1, user1 } = await loadFixture(fixture)

        await mezo.connect(owner).setMinter(minter1.address, true)
        await mezo.connect(minter1).mint(user1.address, ethers.parseEther("100"))

        await expect(
          mezo.connect(user1).burn(ethers.parseEther("100"))
        ).to.be.revertedWithCustomError(mezo, "NotBurner")
      })
    })

    context("when burner is revoked", function () {
      it("should revert with NotBurner error", async function () {
        const { mezo, owner, minter1, burner1 } = await loadFixture(fixture)

        await mezo.connect(owner).setMinter(minter1.address, true)
        await mezo.connect(owner).setBurner(burner1.address, true)
        await mezo.connect(minter1).mint(burner1.address, ethers.parseEther("100"))
        await mezo.connect(owner).setBurner(burner1.address, false)

        await expect(
          mezo.connect(burner1).burn(ethers.parseEther("100"))
        ).to.be.revertedWithCustomError(mezo, "NotBurner")
      })
    })

    context("when burning more than balance", function () {
      it("should revert", async function () {
        const { mezo, owner, minter1, burner1 } = await loadFixture(fixture)

        await mezo.connect(owner).setMinter(minter1.address, true)
        await mezo.connect(owner).setBurner(burner1.address, true)
        await mezo.connect(minter1).mint(burner1.address, ethers.parseEther("100"))

        await expect(
          mezo.connect(burner1).burn(ethers.parseEther("101"))
        ).to.be.reverted
      })
    })
  })
})
