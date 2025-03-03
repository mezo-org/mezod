import { buildModule }  from "@nomicfoundation/hardhat-ignition/modules";

const BTCModule = buildModule("BTC", (m) => {
  const btc = m.contract("BTCCaller");

  return { btc };
});

module.exports = BTCModule;