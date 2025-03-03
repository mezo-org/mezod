import { buildModule }  from "@nomicfoundation/hardhat-ignition/modules";

const UpgradeModule = buildModule("Upgrade", (m) => {
  const upgrade = m.contract("UpgradeCaller");

  return { upgrade };
});

module.exports = UpgradeModule;