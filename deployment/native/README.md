# Native binaries

This document describes:

- prerequisites for Mezo Validator Kit,
- how Mezo Validator Kit for native binaries works,
- deployment process for the validator.

> [!NOTE]
> For now, transactions for entering the testnet/mainnet (aka submitting validator application)
> are out of scope for this script. Feature will be supported as soon as Mezo team implements
> it into the Mezo binary.

## Prerequisites

Native binaries installation is tested on the following operating systems:

- Ubuntu 24 LTS and higher (x86_64 arch)
- Debian 13 Trixie and higher (x86_64 arch)

> [!IMPORTANT]
> If you are planning to install on older system versions or other distributions,
> it's not guaranteed it will work.

Before setup, make sure you have `v-kit.sh` and `testnet.env` on your machine.

Make sure to you can run the setup script as `root` or using `sudo`.

## Setup

### 1. Prepare environment file

For the validator to be successfully deployed, it's necessary to
fill the environment file (in case of testnet it's `testnet.env`).

1. Edit the following variables in `testnet.env`:

- `MEZOD_MONIKER` - a human-readable name for the validator
(Example: `my-lovely-validator`)
- `MEZOD_KEYRING_NAME` - a human-readable name for the mezod keyring
(Example: `my-lovely-keyring`)
- `MEZOD_KEYRING_PASSWORD` - password for the keyring
(to generate best possible password, you can use `openssl rand -hex 32` command)
- `MEZOD_ETHEREUM_SIDECAR_SERVER_ETHEREUM_NODE_ADDRESS` - address for the Ethereum node
(required for the sidecar to run)
- `MEZOD_PUBLIC_IP` - public IP address of the validator
- `GITHUB_TOKEN` - github token with `repo` scope
(required to download mezo binary from github release)

### 2. Prepare installation script to run

Before running `v-kit.sh`, make sure it can be executed by your shell:

```bash
chmod +x v-kit.sh
```

### 3. Run the script (setup validator)

#### Before running: acknowledge your options

Deployment script has the following options:

```text
Usage: ./v-kit.sh

        [stop <opt>]
                stop chosen mezo service (opts: mezo|ethereum-sidecar|connect-sidecar)

        [start <opt>]
                start chosen mezo service (opts: mezo|ethereum-sidecar|connect-sidecar)

        [restart <opt>]
                restart chosen mezo service (opts: mezo|ethereum-sidecar|connect-sidecar)

        [logs <opt>]
                show logs for  chosen mezo service (opts: mezo|ethereum-sidecar|connect-sidecar)

        [-b/--backup]
                backup mezo home dir to /var/mezod-backups

        [-c/--cleanup]
                clean up the installation
                WARNING: this option removes whole Mezo directory (/var/mezod) INCLUDING PRIVATE KEYS

        [--health]
                check health of mezo systemd services

        [-s/--show-variables]
                output variables read from env files

        [-v/--validator-info]
                show validator info

        [-e/--envfile <arg>]
                set file with environment variables for setup script

        [-h/--help]
                show this prompt
```

To run full validator setup, run:

with sudo:

```bash
sudo ./v-kit.sh --run
```

or as root:

```bash
./all-in-one --run
```

> [!IMPORTANT]
> If you are using an environment file other than `testnet.env` make sure to set `--envfile` flag.
>
> ```bash
> ./v-kit.sh --run --envfile <your_custom_envfile>
> ```

## Other options

### Backup mezo home directory

Backup creates a new folder using the name of mezo home dir suffixed by `-backups`
(example: `/var/mezod-backups` when home dir is `/var/mezod`).
After that it zips the whole home dir to the created folder.

```bash
./v-kit.sh -b
```

```bash
./v-kit.sh --backup
```

### Clean up the mezo installation

> [!WARNING]
> This option removes whole Mezo directory (/var/mezod) INCLUDING PRIVATE KEYS.
> It's highly recommended to backup the home dir before cleanup.

```bash
./v-kit.sh -c
```

```bash
./v-kit.sh --cleanup
```

### Check health of mezo systemd services

```bash
./v-kit.sh --health
```

### Verbose printing with variables

This option views all env variables read by the script and activates shell flag that prints
all executed commands and their results (`set -x`).

```bash
./v-kit.sh -s
```

```bash
./v-kit.sh --show-variables
```

### Start/Stop/Status/Logs

You can perform basic administrative actions using the script:

```bash
./v-kit.sh <start/stop/status/logs> <service>
```

Example:

```bash
./v-kit.sh logs mezo
```
