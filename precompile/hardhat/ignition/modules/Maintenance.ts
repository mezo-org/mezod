import { buildModule }  from "@nomicfoundation/hardhat-ignition/modules";

const MaintenanceModule = buildModule("Maintenance", (m) => {
  const maintenance = m.contract("MaintenanceCaller");

  return { maintenance };
});

module.exports = MaintenanceModule;