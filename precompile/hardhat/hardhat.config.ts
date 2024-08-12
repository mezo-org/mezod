import { vars, HardhatUserConfig } from "hardhat/config";
import "@nomicfoundation/hardhat-toolbox";
// import precompile tasks
import "./tasks/validatorPool"
import "./tasks/btctoken"

const getPrivKeys = () : string[] => {
  const strings: string[] = vars.get("MEZO_ACCOUNTS", "").split(",");
  const keys: string[] = [];
  for (let i = 0; i < strings.length; i++) {
    if (strings[i] !== "") {
      keys.push(strings[i]);
    } 
  }
  return keys;
};

const config: HardhatUserConfig = {
  solidity: "0.8.24",
  defaultNetwork: 'localhost',
  networks: {
    localhost: {
      url: "http://localhost:8545",
      chainId: 31611,
      accounts: getPrivKeys(),
      gas: "auto"
    },
    mezo_testnet: {
      url: "http://mezo-node-0.test.mezo.org:8545",
      chainId: 31611,
      accounts: getPrivKeys()
    }
  }
};


export default config;
