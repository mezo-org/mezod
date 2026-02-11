import { ethers, upgrades } from "hardhat"

const describeFn =
  process.env.NODE_ENV === "upgrades-test" ? describe : describe.skip

describeFn("mERC20 tokens - upgrade tests", () => {
  const tokens = [
    { contractName: "mcbBTC", proxyAddress: "0x6a7CD8E1384d49f502b4A4CE9aC9eb320835c5d7" },
    { contractName: "mDAI", proxyAddress: "0x1531b6e3d51BF80f634957dF81A990B92dA4b154" },
    { contractName: "mFBTC", proxyAddress: "0x812fcC0Bb8C207Fd8D6165a7a1173037F43B2dB8" },
    { contractName: "mSolvBTC", proxyAddress: "0xa10aD2570ea7b93d19fDae6Bd7189fF4929Bc747" },
    { contractName: "mswBTC", proxyAddress: "0x29fA8F46CBB9562b87773c8f50a7F9F27178261c" },
    { contractName: "mT", proxyAddress: "0xaaC423eDC4E3ee9ef81517e8093d52737165b71F" },
    { contractName: "mUSDC", proxyAddress: "0x04671C72Aab5AC02A03c1098314b1BB6B560c197" },
    { contractName: "mUSDe", proxyAddress: "0xdf6542260a9F768f07030E4895083F804241F4C4" },
    { contractName: "mUSDT", proxyAddress: "0xeB5a5d39dE4Ea42C2Aa6A57EcA2894376683bB8E" },
    { contractName: "mxSolvBTC", proxyAddress: "0xdF708431162Ba247dDaE362D2c919e0fbAfcf9DE" },
  ]

  for (const token of tokens) {
    it(`should be able to upgrade the current mainnet version of ${token.contractName}`, async () => {
      const Token = await ethers.getContractFactory(token.contractName)
      await upgrades.validateUpgrade(token.proxyAddress, Token, { kind: "transparent" })
    })
  }
})

