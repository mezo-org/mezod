import { ethers } from "ethers"

/**
 * The address of the SingletonFactory defined in [[EIP-2470]].
 *
 * [[EIP-2470]]: https://eips.ethereum.org/EIPS/eip-2470.
 */
export const singletonFactoryAddress =
  "0xce0042B868300000d44A59004Da54A005ffdcf9f"

/**
 * The address of the SingletonFactory defined in [[EIP-2470]].
 *
 * [[EIP-2470]]: https://eips.ethereum.org/EIPS/eip-2470.
 */
export const singletonFactoryABI = [
  {
    constant: false,
    inputs: [
      {
        internalType: "bytes",
        name: "_initCode",
        type: "bytes",
      },
      {
        internalType: "bytes32",
        name: "_salt",
        type: "bytes32",
      },
    ],
    name: "deploy",
    outputs: [
      {
        internalType: "address payable",
        name: "createdContract",
        type: "address",
      },
    ],
    payable: false,
    stateMutability: "nonpayable",
    type: "function",
  },
]

/**
 * The address that should deploy the SingletonFactory defined in [[EIP-2470]].
 */
const singletonFactoryDeployerAddress =
  "0xBb6e024b9cFFACB947A71991E386681B1Cd1477D"

/**
 * The cost of the deployment of the SingletonFactory defined in [[EIP-2470]].
 */
const singletonFactoryDeploymentCost = 24700000000000000n

/**
 * The transaction data that will deploy the SingletonFactory defined in
 * [[EIP-2470]].
 */
const singletonFactoryDeploymentTx =
  "0xf9016c8085174876e8008303c4d88080b90154608060405234801561001057600080fd5b50610134806100206000396000f3fe6080604052348015600f57600080fd5b506004361060285760003560e01c80634af63f0214602d575b600080fd5b60cf60048036036040811015604157600080fd5b810190602081018135640100000000811115605b57600080fd5b820183602082011115606c57600080fd5b80359060200191846001830284011164010000000083111715608d57600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600092019190915250929550509135925060eb915050565b604080516001600160a01b039092168252519081900360200190f35b6000818351602085016000f5939250505056fea26469706673582212206b44f8a82cb6b156bfcc3dc6aadd6df4eefd204bc928a4397fd15dacf6d5320564736f6c634300060200331b83247000822470"

/**
 * A contract instance for the SingletonFactory defined in [[EIP-2470]].
 *
 * [[EIP-2470]]: https://eips.ethereum.org/EIPS/eip-2470.
 */
export const singletonFactory = new ethers.Contract(
  singletonFactoryAddress,
  [
    {
      constant: false,
      inputs: [
        {
          internalType: "bytes",
          name: "_initCode",
          type: "bytes",
        },
        {
          internalType: "bytes32",
          name: "_salt",
          type: "bytes32",
        },
      ],
      name: "deploy",
      outputs: [
        {
          internalType: "address payable",
          name: "createdContract",
          type: "address",
        },
      ],
      payable: false,
      stateMutability: "nonpayable",
      type: "function",
    },
  ],
)

/**
 * Ensures there is a singleton factory on the network the given Signer is
 * connected to, and returns a contract attached to the given signer for that
 * instance.
 *
 * Deploys the factory if needed.
 */
export async function ensureSingletonFactory(
  signer: ethers.Signer,
): Promise<ethers.Contract> {
  if (!signer.provider) throw new Error("Signer not connected to a provider")

  const connectedFactory = singletonFactory.connect(signer)

  const existingCode = await signer.provider.getCode(singletonFactoryAddress)

  if (
    typeof existingCode !== "undefined" &&
    existingCode !== null &&
    existingCode !== "0x" &&
    existingCode !== "0x0"
  ) {
    // Already deployed, just return the connected factory.
    return connectedFactory as ethers.Contract
  }

  await signer.sendTransaction({
    to: singletonFactoryDeployerAddress,
    value: singletonFactoryDeploymentCost,
  })

  // Wait for confirmation.
  await (
    await signer.provider.broadcastTransaction(singletonFactoryDeploymentTx)
  ).wait()

  return connectedFactory as ethers.Contract
}
