// This script lists available hardhat accounts for a given network
// from the hardhat directory run:
// npx hardhat --network NETWORK run scripts/accounts.ts
//
// networks:
// * localhost
// * mezo_testnet
import hre from "hardhat"

async function main() {
    const accounts = await hre.ethers.getSigners();

    for (const account of accounts) {
        console.log(account.address);
    }
}

main().catch((error) => {
    console.error(error);
    process.exitCode = 1;
});