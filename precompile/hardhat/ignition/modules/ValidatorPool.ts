import { buildModule }  from "@nomicfoundation/hardhat-ignition/modules";

const ValidatorPoolModule = buildModule("ValidatorPool", (m) => {
  const validatorPool = m.contract("ValidatorPoolCaller");

  return { validatorPool };
});

module.exports = ValidatorPoolModule;