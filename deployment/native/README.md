# Mezo Validator Kit - Native binaries

This document describes:

- prerequisites for Mezo Validator Kit,
- how Mezo Validator Kit for native binaries works,
- deployment process for the validator.

## Prerequisites

Native binaries installation is tested on the following operating systems:
- Ubuntu 24 LTS and higher (x86_64 arch)
- Debian 13 Trixie and higher (x86_64 arch)

> [!IMPORTANT]
> If you are planning to install on older system versions or other distributions,
> it's not guaranteed it will work.

Before setup, make sure you have `all-in-one.sh` and `testnet.env` on your machine.

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

Before running `all-in-one.sh`, make sure it can be executed by your shell:

```bash
chmod +x all-in-one.sh
```

### 3. Run the script (setup validator)

#### Before running: acknowledge your options

Deployment script has the following options:

```text
Usage: ./all-in-one.sh

        [-r/--run]
                run the installation

        [-b/--backup]
                backup mezo home dir to /var/mezod-backups

        [-c/--cleanup]
                clean up the installation
                WARNING: this option removes whole Mezo directory (/var/mezod) 
                INCLUDING PRIVATE KEYS

        [--health]
                check health of mezo systemd services

        [-s/--show-variables]
                output variables read from env files

        [-e/--envfile]
                set file with environment variables for setup script

        [-h/--help]
                show this prompt
```

To run full validator setup, run:

with sudo:

```bash
sudo ./all-in-one --run
```

or as root:

```bash
./all-in-one --run
```

> [!IMPORTANT]
> If you are using an environment file other than `testnet.env` make sure to set `--envfile` flag.
>
> ```bash
> ./all-in-one.sh --run --envfile <your_custom_envfile>
> ```

## Other options

### Backup mezo home directory

Backup creates a new folder using the name of mezo home dir prefixed by `-backups`
(example: `/var/mezod-backups` when home dir is `/var/mezod`).
After that it zips the whole home dir to the created folder.

```bash
./all-in-one.sh -b
```

```bash
./all-in-one.sh --backup
```

### Clean up the mezo installation

> [!WARNING]
> This option removes whole Mezo directory (/var/mezod) INCLUDING PRIVATE KEYS. 
> It's highly recommended to backup the home dir before cleanup.

```bash
./all-in-one.sh -c
```

```bash
./all-in-one.sh --cleanup
```

### Check health of mezo systemd services

```bash
./all-in-one.sh --health
```

### Verbose printing with variables

This option views all env variables read by the script and activates shell flag that prints
all executed commands and their results (`set -x`).

```bash
./all-in-one.sh -s
```

```bash
./all-in-one.sh --show-variables
```
