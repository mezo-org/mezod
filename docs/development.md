# Local development

## Running the client locally

### Single instance

To start a single client instance locally please use the `local_node.sh` script.
The script will compile the client, perform genesis, and start a local client
instance.

### Multiple client instances

There is a set of `make` targets allowing to set up multiple client instances
connected to each other locally. The formulas use Docker Compose to set up the
cluster so on macOS, Docker Desktop must be running locally.

```bash
# Build and install the `meso/node` image locally.
$ make localnet-build
# Start four client instances.
$ make localnet-start
# Stop client instances started with localnet-start.
$ make localnet-stop
# Show logs of all four client instances together.
$ make localnet-show-logstream
```

The client build can take some time even if there have been no changes, so it is
not executed as part of the `localnet-start` target to save time. To rebuild the
client after code changes, please explicitly re-run the `localnet-build` target.

The configuration for the four nodes is available in the `build` directory.

When `make localnet-start` is used for the first time, a configuration for the
four nodes is generated. To clear the chain state and reset the configuration,
use `make localnet-clean` followed by `make localnet-start`.
