import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

const MAX_VALIDATORS = 150;

const ValidatorPoolModule = buildModule("ValidatorPoolModule", (m) => {
  const initialOwner = m.getParameter("initialOwner", m.getAccount(0));
  const maxValidators = m.getParameter("maxValidators", MAX_VALIDATORS);

  const validatorPool = m.contract("ValidatorPool", [initialOwner, maxValidators]);

  return { validatorPool };
});

export default ValidatorPoolModule;