# Testnet

## Overview

The Mezo testnet is a network of nodes running the Mezo client software, 
assisted by several auxiliary services. This network is used for testing and 
experimentation. Specific components that form the core of the testnet are:


| Component               | Public address                 |
|-------------------------|--------------------------------|
| Validator 0             | mezo-node-0.test.mezo.org      |
| Validator 1             | mezo-node-1.test.mezo.org      |
| Validator 2             | mezo-node-2.test.mezo.org      |
| Validator 3             | mezo-node-3.test.mezo.org      |
| Validator 4             | mezo-node-4.test.mezo.org      |
| Block explorer app      | https://explorer.test.mezo.org |
| Block explorer server   | N/A                            |
| Block explorer database | N/A                            |

Validator nodes expose the following services:

| Service       | Port  |
|---------------|-------|
| P2P           | 26656 |
| RPC           | 26657 |
| REST          | 1317  |
| GRPC          | 9090  |
| JSON-RPC HTTP | 8545  |
| JSON-RPC WS   | 8546  |

External nodes can join the network by connecting to any of the validator nodes.
A standard onboarding procedure for validator and non-validator nodes is
in progress and will be documented in the future.

## Setup

From the operational perspective, all aforementioned core components of the 
testnet are Kubernetes deployments, living on a Google Kubernetes Engine (GKE) 
cluster. The cluster is managed by the Mezo engineering team and is not
accessible to the public. Below, we provide an overview of the setup from
the ground up.

### Testnet artifacts

The main artifacts necessary to boostrap any Mezo chain are:
- A global genesis file defining the initial state of the chain 
  (most importantly, the initial validator set)
- Configuration packages for initial validators (private keys, configuration 
  files, etc.)
- A configuration package for the faucet
  (used to distribute the gas token)
- A seed file containing the public addresses of the initial validators
  (used by new nodes to join the network)

We provide an automated way to generate these artifacts using the 
[`scripts/public_testnet.sh`](../scripts/public-testnet.sh) script. 
This script generates the genesis file, 5 configuration packages for 
initial validators, one configuration package for the faucet, and a seed file. 
The resulting artifacts are stored in the 
[`.public-testnet`](../.public-testnet) directory.

---
**Note**: In the future, the `scripts/public-testnet.sh` script will be
integrated into the Mezo client software and exposed as a command-line
interface (CLI) command.
---

