import { buildModule }  from "@nomicfoundation/hardhat-ignition/modules";

const AssetsBridgeModule = buildModule("AssetsBridge", (m) => {
  const assetsBridge = m.contract("AssetsBridgeCaller");

  return { assetsBridge };
});

module.exports = AssetsBridgeModule;