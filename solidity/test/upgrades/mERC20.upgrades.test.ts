import { expect } from "chai"
import { ethers, upgrades } from "hardhat"
import hre from "hardhat"

const describeFn =
  process.env.NODE_ENV === "upgrades-test" ? describe : describe.skip

describeFn("mERC20 tokens - upgrade tests", () => {
  const tokens = [
    { contractName: "mSolvBTC", proxyAddress: "0xa10aD2570ea7b93d19fDae6Bd7189fF4929Bc747" },
    { contractName: "mT", proxyAddress: "0xaaC423eDC4E3ee9ef81517e8093d52737165b71F" },
    { contractName: "mxSolvBTC", proxyAddress: "0xdF708431162Ba247dDaE362D2c919e0fbAfcf9DE" },
    { contractName: "mUSDC", proxyAddress: "0x04671C72Aab5AC02A03c1098314b1BB6B560c197" },
    { contractName: "mUSDT", proxyAddress: "0xeB5a5d39dE4Ea42C2Aa6A57EcA2894376683bB8E" },
    { contractName: "mUSDe", proxyAddress: "0xdf6542260a9F768f07030E4895083F804241F4C4" },
    { contractName: "mFBTC", proxyAddress: "0x812fcC0Bb8C207Fd8D6165a7a1173037F43B2dB8" },
    { contractName: "mcbBTC", proxyAddress: "0x6a7CD8E1384d49f502b4A4CE9aC9eb320835c5d7" },
    { contractName: "mDAI", proxyAddress: "0x1531b6e3d51BF80f634957dF81A990B92dA4b154" },
    { contractName: "mswBTC", proxyAddress: "0x29fA8F46CBB9562b87773c8f50a7F9F27178261c" },
  ]

  for (const token of tokens) {
    it(`should be able to upgrade the current mainnet version of ${token.contractName}`, async () => {
      const Token = await ethers.getContractFactory(token.contractName)
      await upgrades.validateUpgrade(token.proxyAddress, Token, { kind: "transparent" })
    }).timeout(120_000)
  }

  it("should be able to upgrade and initialize permit for mUSDC", async () => {
    // Run the full upgrade flow on a single representative token.
    const proxyAddress = "0x04671C72Aab5AC02A03c1098314b1BB6B560c197"
    const Token = await ethers.getContractFactory("mUSDC")
    const newImplementationAddress = await upgrades.prepareUpgrade(proxyAddress, Token, {
      kind: "transparent",
    })

    const proxyAdminAddress = await upgrades.erc1967.getAdminAddress(proxyAddress)
    const proxyAdmin = await ethers.getContractAt(
      [
        "function owner() view returns (address)",
        "function upgradeAndCall(address proxy, address implementation, bytes data)",
      ],
      proxyAdminAddress,
    )
    const proxyAdminOwner = await proxyAdmin.owner()

    await hre.network.provider.send("hardhat_impersonateAccount", [proxyAdminOwner])
    await hre.network.provider.send("hardhat_setBalance", [
      proxyAdminOwner,
      ethers.toBeHex(ethers.parseEther("1")),
    ])

    const proxyAdminOwnerSigner = await ethers.getSigner(proxyAdminOwner)
    const callData = Token.interface.encodeFunctionData("initializeV2")

    await proxyAdmin
      .connect(proxyAdminOwnerSigner)
      .upgradeAndCall(proxyAddress, newImplementationAddress, callData)

    await hre.network.provider.send("hardhat_stopImpersonatingAccount", [
      proxyAdminOwner,
    ])

    const token = await ethers.getContractAt("mUSDC", proxyAddress)
    await expect(token.initializeV2.staticCall()).to.be.revertedWithCustomError(
      token,
      "InvalidInitialization",
    )

    const domainSeparator = await token.DOMAIN_SEPARATOR()
    expect(domainSeparator).to.not.equal(ethers.ZeroHash)

    const [owner, spender] = await ethers.getSigners()
    const amount = ethers.parseEther("100")
    const deadline = ethers.MaxUint256
    const nonce = await token.nonces(owner.address)
    const { chainId } = await ethers.provider.getNetwork()
    const domain = {
      name: await token.name(),
      version: "1",
      chainId: Number(chainId),
      verifyingContract: proxyAddress,
    }

    const types = {
      Permit: [
        { name: "owner", type: "address" },
        { name: "spender", type: "address" },
        { name: "value", type: "uint256" },
        { name: "nonce", type: "uint256" },
        { name: "deadline", type: "uint256" },
      ],
    }
    const values = {
      owner: owner.address,
      spender: spender.address,
      value: amount,
      nonce,
      deadline,
    }

    const signature = await owner.signTypedData(domain, types, values)
    const { v, r, s } = ethers.Signature.from(signature)

    await token
      .connect(spender)
      .permit(owner.address, spender.address, amount, deadline, v, r, s)

    expect(await token.allowance(owner.address, spender.address)).to.equal(amount)
    expect(await token.nonces(owner.address)).to.equal(nonce + 1n)
  }).timeout(120_000)
})
