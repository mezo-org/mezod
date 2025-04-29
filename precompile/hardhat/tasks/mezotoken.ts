import { task, vars } from 'hardhat/config'
import '@nomicfoundation/hardhat-toolbox'

import abi from '../../mezotoken/abi.json'
const precompileAddress = '0x7b7c000000000000000000000000000000000001'

task('mezoToken:name', 'Returns the name of the token', async (taskArguments, hre) => {
  const mezotoken = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
  const name = await mezotoken.name()
  console.log(name)
})

task('mezoToken:symbol', 'Returns the symbol of the token', async (taskArguments, hre) => {
  const mezotoken = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
  const symbol = await mezotoken.symbol()
  console.log(symbol)
})

task('mezoToken:decimals', 'Returns the decimal places of the token', async (taskArguments, hre) => {
  const mezotoken = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
  const decimals = await mezotoken.decimals()
  console.log(decimals)
})

task('mezoToken:totalSupply', 'Returns the total number of tokens in existence', async (taskArguments, hre) => {
  const mezotoken = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
  const totalSupply = await mezotoken.totalSupply()
  console.log(totalSupply)
})

task('mezoToken:balanceOf', 'Returns the number of tokens owned by `account`')
  .addParam('account', 'Account to get balance of')
  .setAction(async (taskArguments, hre) => {
    const mezotoken = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
    const balance = await mezotoken.balanceOf(taskArguments.account)
    console.log(balance)
  })

task('mezoToken:allowance', 'Returns the remaining number of tokens that `spender` will be allowed to spend on behalf of `owner` through {transferFrom}')
  .addParam('owner', 'Account owning the funds')
  .addParam('spender', 'Account allowed to spend funds on behalf of the owner')
  .setAction(async (taskArguments, hre) => {
    const mezotoken = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
    const allowance = await mezotoken.allowance(taskArguments.owner, taskArguments.spender)
    console.log(allowance)
  })

task('mezoToken:transfer', 'Moves a `value` amount of tokens from the caller\'s account to `to`')
  .addParam('signer', 'The signer address (msg.sender)')
  .addParam('to', 'Recipient account')
  .addParam('value', 'Value to send (amezo)')
  .setAction(async (taskArguments, hre) => {
    const signer = await hre.ethers.getSigner(taskArguments.signer)
    const mezotoken = new hre.ethers.Contract(precompileAddress, abi, signer)
    const pending = await mezotoken.transfer(taskArguments.to, taskArguments.value)
    const confirmed = await pending.wait()
    console.log(confirmed.hash)
  })

task('mezoToken:approve', 'Sets a `value` amount of tokens as the allowance of `spender` over the caller\'s tokens')
  .addParam('signer', 'The signer address (msg.sender)')
  .addParam('spender', 'The account to approve')
  .addParam('value', 'Allowance value (amezo)')
  .setAction(async (taskArguments, hre) => {
    const signer = await hre.ethers.getSigner(taskArguments.signer)
    const mezotoken = new hre.ethers.Contract(precompileAddress, abi, signer)
    const pending = await mezotoken.approve(taskArguments.spender, taskArguments.value)
    const confirmed = await pending.wait()
    console.log(confirmed.hash)
  })

task('mezoToken:transferFrom', 'Moves a `value` amount of tokens from `from` to `to` using the allowance mechanism')
  .addParam('signer', 'The signer address (msg.sender)')
  .addParam('from', 'Origin account')
  .addParam('to', 'Recipient account')
  .addParam('value', 'Value to send (amezo)')
  .setAction(async (taskArguments, hre) => {
    const signer = await hre.ethers.getSigner(taskArguments.signer)
    const mezotoken = new hre.ethers.Contract(precompileAddress, abi, signer)
    const pending = await mezotoken.transferFrom(taskArguments.from, taskArguments.to, taskArguments.value)
    const confirmed = await pending.wait()
    console.log(confirmed.hash)
  })

task('mezoToken:permit', 'Sets an allowance of MEZO tokens for a spender with an owner\'s signature')
  .addParam('signer', 'The signer address (msg.sender)')
  .addParam('owner', 'Account owning the funds')
  .addParam('spender', 'Account allowed to spend funds on behalf of the owner')
  .addParam('amount', 'Allowance value (amezo)')
  .addParam('deadline', 'Expiry time for the permit (in Unix format)')
  .addParam('v', 'v-component of the signature')
  .addParam('r', 'r-component of the signature')
  .addParam('s', 's-component of the signature')
  .setAction(async (taskArguments, hre) => {
    const signer = await hre.ethers.getSigner(taskArguments.signer)
    const mezotoken = new hre.ethers.Contract(precompileAddress, abi, signer)
    const pending = await mezotoken.permit(
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
  'mezoToken:DOMAIN_SEPARATOR',
  'Returns hash of EIP712 Domain struct with the token name as a signing domain and token contract as a verifying contract',
  async (taskArguments, hre) => {
    const mezotoken = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
    const domainSeparator = await mezotoken.DOMAIN_SEPARATOR()
    console.log(domainSeparator)
  })

task('mezoToken:nonce', 'Returns the current nonce for EIP2612 permission for the provided token owner')
  .addParam('owner', 'Recipient account')
  .setAction(async (taskArguments, hre) => {
    const mezotoken = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
    const nonce = await mezotoken.nonce(taskArguments.owner)
    console.log(nonce)
  })

task('mezoToken:PERMIT_TYPEHASH', 'Returns the EIP2612 Permit message hash', async (taskArguments, hre) => {
  const mezotoken = new hre.ethers.Contract(precompileAddress, abi, hre.ethers.provider)
  const permitTypehash = await mezotoken.PERMIT_TYPEHASH()
  console.log(permitTypehash)
})

task('mezoToken:PERMIT_SIGNATURE', 'Returns a signature that can be used with `permit`')
  .addParam('signer', 'The signer address (msg.sender)')
  .addParam('owner', 'Account owning the funds')
  .addParam('spender', 'Account allowed to spend funds on behalf of the owner')
  .addParam('amount', 'Allowance value (amezo)')
  .addParam('deadline', 'Expiry time for the permit (in Unix format)')
  .setAction(async (taskArguments, hre) => {
    // We use ethers.SigningKey for a Wallet instead of
    // Signer.signMessage to do not add '\x19Ethereum Signed Message:\n'
    // prefix to the signed message. The '\x19` protection (see EIP191 for
    // more details on '\x19' rationale and format) is already included in
    // EIP2612 permit signed message and '\x19Ethereum Signed Message:\n'
    // should not be used there.
    const signer = await hre.ethers.getSigner(taskArguments.signer)
    const privkeys: string[] = vars.get('MEZO_ACCOUNTS', '').split(',')
    const prefix: string = '0x1901'
    // Loop through available keys, create SigningKey based wallet
    // to compare addresses with the given signer address
    for (let i = 0; i < privkeys.length; i++) {
      const signingKey = new hre.ethers.SigningKey(privkeys[i])
      const wallet = new hre.ethers.BaseWallet(signingKey)
      if (wallet.address === signer.address) {
        // This is the correct key/wallet. Get info from contract/precompile
        // to produce data to sign
        const mezotoken = new hre.ethers.Contract(precompileAddress, abi, signer)
        const domainSeparator = await mezotoken.DOMAIN_SEPARATOR()
        const typehash = await mezotoken.PERMIT_TYPEHASH()
        const nonce = await mezotoken.nonce(signer.address)
        // Encode data
        const abiCoder = new hre.ethers.AbiCoder()
        const message = abiCoder.encode(
          ['bytes32', 'address', 'address', 'uint256', 'uint256', 'uint256'],
          [typehash, taskArguments.owner, taskArguments.spender, taskArguments.amount, nonce, taskArguments.deadline]
        )

        // Create digest and sign
        const digest = hre.ethers.keccak256(
          prefix.concat(String(domainSeparator).substring(2), hre.ethers.keccak256(message).substring(2))
        )
        const signature = wallet.signingKey.sign(digest)

        // Output signature values
        console.log('v: %d', signature.v)
        console.log('r: %s', signature.r)
        console.log('s: %s', signature.s)
        break
      }
    }
  }) 
  