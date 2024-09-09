import { task } from 'hardhat/config'
import '@nomicfoundation/hardhat-toolbox'

import abi from '../../btctoken/abi.json'
const precompileAddress = '0x7b7c000000000000000000000000000000000000'

task('btcToken:name', 'Returns the name of the token', async (taskArguments, hre) => {
  const btctoken = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
  const name = await btctoken.name()
  console.log(name)
})

task('btcToken:symbol', 'Returns the symbol of the token', async (taskArguments, hre) => {
  const btctoken = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
  const symbol = await btctoken.symbol()
  console.log(symbol)
})

task('btcToken:decimals', 'Returns the decimal places of the token', async (taskArguments, hre) => {
  const btctoken = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
  const decimals = await btctoken.decimals()
  console.log(decimals)
})

task('btcToken:totalSupply', 'Returns the total number of tokens in existence', async (taskArguments, hre) => {
  const btctoken = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
  const totalSupply = await btctoken.totalSupply()
  console.log(totalSupply)
})

task('btcToken:balanceOf', 'Returns the number of tokens owned by `account`')
  .addParam('account', 'Account to get balance of')
  .setAction(async (taskArguments, hre) => {
    const btctoken = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
    const balance = await btctoken.balanceOf(taskArguments.account)
    console.log(balance)
  })

task('btcToken:allowance', 'Returns the remaining number of tokens that `spender` will be allowed to spend on behalf of `owner` through {transferFrom}')
  .addParam('owner', 'Account owning the funds')
  .addParam('spender', 'Account allowed to spend funds on behalf of the owner')
  .setAction(async (taskArguments, hre) => {
    const btctoken = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
    const allowance = await btctoken.allowance(taskArguments.owner, taskArguments.spender)
    console.log(allowance)
  })

task('btcToken:transfer', 'Moves a `value` amount of tokens from the caller\'s account to `to`')
  .addParam('signer', 'The signer address (msg.sender)')
  .addParam('to', 'Recipient account')
  .addParam('value', 'Value to send (abtc)')
  .setAction(async (taskArguments, hre) => {
    const signer = await hre.ethers.getSigner(taskArguments.signer)
    const btctoken = new hre.ethers.Contract(precompileAddress, abi, signer)
    const pending = await btctoken.transfer(taskArguments.to, taskArguments.value)
    const confirmed = await pending.wait()
    console.log(confirmed.hash)
  })

task('btcToken:approve', 'Sets a `value` amount of tokens as the allowance of `spender` over the caller\'s tokens')
  .addParam('signer', 'The signer address (msg.sender)')
  .addParam('spender', 'The account to approve')
  .addParam('value', 'Allowance value (abtc)')
  .setAction(async (taskArguments, hre) => {
    const signer = await hre.ethers.getSigner(taskArguments.signer)
    const btctoken = new hre.ethers.Contract(precompileAddress, abi, signer)
    const pending = await btctoken.approve(taskArguments.spender, taskArguments.value)
    const confirmed = await pending.wait()
    console.log(confirmed.hash)
  })

task('btcToken:transferFrom', 'Moves a `value` amount of tokens from `from` to `to` using the allowance mechanism')
  .addParam('signer', 'The signer address (msg.sender)')
  .addParam('from', 'Origin account')
  .addParam('to', 'Recipient account')
  .addParam('value', 'Value to send (abtc)')
  .setAction(async (taskArguments, hre) => {
    const signer = await hre.ethers.getSigner(taskArguments.signer)
    const btctoken = new hre.ethers.Contract(precompileAddress, abi, signer)
    const pending = await btctoken.transferFrom(taskArguments.from, taskArguments.to, taskArguments.value)
    const confirmed = await pending.wait()
    console.log(confirmed.hash)
  })

task('btcToken:permit', 'Sets an allowance of BTC tokens for a spender with an owner\'s signature')
  .addParam('signer', 'The signer address (msg.sender)')
  .addParam('owner', 'Account owning the funds')
  .addParam('spender', 'Account allowed to spend funds on behalf of the owner')
  .addParam('amount', 'Allowance value (abtc)')
  .addParam('deadline', 'Expiry time for the permit (in Unix format)')
  .addParam('v', 'v-component of the signature')
  .addParam('r', 'r-component of the signature')
  .addParam('s', 's-component of the signature')
  .setAction(async (taskArguments, hre) => {
    const signer = await hre.ethers.getSigner(taskArguments.signer)
    const btctoken = new hre.ethers.Contract(precompileAddress, abi, signer)
    const pending = await btctoken.permit(
      taskArguments.owner,
      taskArguments.spender,
      taskArguments.amount,
      taskArguments.deadline,
      taskArguments.v,
      taskArguments.r,
      taskArguments.s
    )
    const confirmed = await pending.wait()
    console.log(confirmed.hash)
  })

task(
  'btcToken:DOMAIN_SEPARATOR',
  'Returns hash of EIP712 Domain struct with the token name as a signing domain and token contract as a verifying contract',
  async (taskArguments, hre) => {
    const btctoken = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
    const domainSeparator = await btctoken.DOMAIN_SEPARATOR()
    console.log(domainSeparator)
  })

task('btcToken:nonce', 'Returns the current nonce for EIP2612 permission for the provided token owner')
  .addParam('owner', 'Recipient account')
  .setAction(async (taskArguments, hre) => {
    const btctoken = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
    const nonce = await btctoken.nonce(taskArguments.owner)
    console.log(nonce)
  })

task('btcToken:PERMIT_TYPEHASH', 'Returns the EIP2612 Permit message hash', async (taskArguments, hre) => {
  const btctoken = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
  const permitTypehash = await btctoken.PERMIT_TYPEHASH()
  console.log(permitTypehash)
})
