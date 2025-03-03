import { buildModule }  from "@nomicfoundation/hardhat-ignition/modules";

const PriceOracleModule = buildModule("PriceOracle", (m) => {
  const priceOracle = m.contract("PriceOracleCaller");

  return { priceOracle };
});

module.exports = PriceOracleModule;