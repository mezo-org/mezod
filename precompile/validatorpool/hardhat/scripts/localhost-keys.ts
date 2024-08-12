import hre from "hardhat"

const BUILD_DIR = "../../../../.localnet/";
const COUNT = 4;

async function main() {
    let keys: string = ""
    for (let i = 0; i < COUNT; i++) {
        const path = BUILD_DIR + "node" + i + "/mezod/key_seed.json";
        const seed = await import(path);
        const pk = hre.ethers.Wallet.fromPhrase(seed.secret).privateKey;
        if (i > 0) {
            keys += ","
        }
        keys += pk
    }
    console.log(keys);
}

main().catch((error) => {
    console.error(error);
    process.exitCode = 1;
});