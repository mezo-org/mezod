# PoA Validator Pool for Mezo

Draft contract/reference implementation for the PoA validator pool used by Mezo for initial launch.

This project uses hardhat. Try running some of the following tasks:

```shell
npx hardhat help
npx hardhat test
REPORT_GAS=true npx hardhat test
npx hardhat node
npx hardhat ignition deploy ./ignition/modules/ValidatorPool.ts
```

### Estimated gas usage

*reported by `REPORT_GAS=true npx hardhat test`*

```
·----------------------------------------|----------------------------|-------------|-----------------------------·
|          Solc version: 0.8.24          ·  Optimizer enabled: false  ·  Runs: 200  ·  Block limit: 30000000 gas  │
·········································|····························|·············|······························
|  Methods                                                                                                        │
··················|······················|··············|·············|·············|···············|··············
|  Contract       ·  Method              ·  Min         ·  Max        ·  Avg        ·  # calls      ·  usd (avg)  │
··················|······················|··············|·············|·············|···············|··············
|  ValidatorPool  ·  approveApplication  ·           -  ·          -  ·      32761  ·           10  ·          -  │
··················|······················|··············|·············|·············|···············|··············
|  ValidatorPool  ·  kick                ·           -  ·          -  ·      48801  ·            3  ·          -  │
··················|······················|··············|·············|·············|···············|··············
|  ValidatorPool  ·  leave               ·           -  ·          -  ·      45876  ·            3  ·          -  │
··················|······················|··············|·············|·············|···············|··············
|  ValidatorPool  ·  submitApplication   ·           -  ·          -  ·      49334  ·           12  ·          -  │
··················|······················|··············|·············|·············|···············|··············
|  Deployments                           ·                                          ·  % of limit   ·             │
·········································|··············|·············|·············|···············|··············
|  ValidatorPool                         ·           -  ·          -  ·    1195328  ·          4 %  ·          -  │
·----------------------------------------|--------------|-------------|-------------|---------------|-------------·
```

### Notes

Imports from Open Zeppelin Ownable and Ownable2Step