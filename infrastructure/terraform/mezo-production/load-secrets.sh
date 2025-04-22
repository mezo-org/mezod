#!/bin/bash

# Load secrets from 1Password.

## Load SSL certificates.

CERT_DIR=./ssl-certificates
CERT_NAMES=("mezo-production-explorer" "mezo-production-monitoring-hub")

for CERT_NAME in "${CERT_NAMES[@]}"; do
  op inject -i $CERT_DIR/$CERT_NAME.crt.tpl -o $CERT_DIR/$CERT_NAME.crt
  op inject -i $CERT_DIR/$CERT_NAME.key.tpl -o $CERT_DIR/$CERT_NAME.key
done
