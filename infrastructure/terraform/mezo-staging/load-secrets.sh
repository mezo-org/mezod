#!/bin/bash

# Load secrets from 1Password.

## Load SSL certificates.

CERT_DIR=./ssl-certificates
CERT_NAMES=("mezo-staging-explorer")

for CERT_NAME in "${CERT_NAMES[@]}"; do
  op inject -i $CERT_DIR/$CERT_NAME.crt.tpl -o $CERT_DIR/$CERT_NAME.crt
  op inject -i $CERT_DIR/$CERT_NAME.key.tpl -o $CERT_DIR/$CERT_NAME.key
done

## Load configuration secrets.

CONFIG_DIR=./configs
CONFIG_NAMES=("faucet-config.yaml")

for CONFIG_NAME in "${CONFIG_NAMES[@]}"; do
  op inject -i $CONFIG_DIR/$CONFIG_NAME.tpl -o $CONFIG_DIR/$CONFIG_NAME
done