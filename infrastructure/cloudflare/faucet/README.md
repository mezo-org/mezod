# Faucet

This module contains the implementation of the Mezo Testnet Faucet.
The faucet is written and deployed as a Cloudflare Worker.

### Development

To work on the faucet locally, you need to set up a Mezo localnet first.
Follow the instructions in the [docs/development.md](../../../docs/development.md) 
file. Binary-based localnet setup is recommended.

Make sure the EVM JSON-RPC of one of the localnet nodes is available under
`http://localhost:8545`. The faucet uses this default endpoint to send transactions. 
This can be changed by setting the `vars.MEZO_API_URL` variable in the 
`wrangler.toml` file.

Once the localnet is running, configure development secrets by running
```shell
cp .dev.vars.sample .dev.vars
```
and fill in the required values in the fresh `.dev.vars` file. 

The `MEZO_FAUCET_PRIVATE_KEY` secret can be obtained by running the 
`mezod keys unsafe-export-eth-key` command against the localnet node's 
homedir supposed to be the purse for the faucet. Assuming binary-based localnet 
is used, the call should look like this:
```shell
# Invoke from `mezod` project root.
./build/mezod keys unsafe-export-eth-key --home=./.localnet/node0/mezod --keyring-backend=test node0   
```

Once `.dev.vars` is ready, you can start the development server by running
```shell
npm run dev
```

The faucet will be available at `http://localhost:8000`. Code changes
are hot-reloaded by Wrangler.

### Deployment

To deploy the faucet to the target environment, run
```shell
npm run deploy
```

This script deploys the faucet as a Cloudflare Worker using Wrangler.

There are several steps in the deployment process:
1. The script reads the faucet's secrets from the dedicated 1Password vault
   and puts them into the Worker environment. Make sure you are authenticated
   with 1Password CLI before running this script and have access to the vault.
   Consult the 1password CLI help to see how to authenticate. At the time of 
   writing, the command is `eval $(op signin)`.
2. The script deploys the Worker using Wrangler. It uses the staging environment
   by default. If you are not authenticated with Wrangler, you will be prompted
   to do so. In case of troubles, you can use `wrangler login` to authenticate,
   manually.