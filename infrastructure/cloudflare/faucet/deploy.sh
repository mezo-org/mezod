#!/bin/bash

# This script deploys the faucet as a Cloudflare Worker using Wrangler.
# There are several steps in the deployment process:
# 1. The script reads the faucet's secrets from the dedicated 1Password vault
#    and puts them into the Worker environment. Make sure you are authenticated
#    with 1Password CLI before running this script and have access to the vault.
#    Consult the 1password CLI help to see how to authenticate. At the time of
#    writing, the command is `eval $(op signin)`.
# 2. The script deploys the Worker using Wrangler. It uses the staging environment
#    by default. You can change it by modifying the `wrangler` alias.

shopt -s expand_aliases

alias wrangler='npx wrangler --env staging'

# Step 1: Read secrets from 1Password and put them into the Worker environment.
op read "op://Mezo DevOps/faucet_private_key/notes" | wrangler secret put MEZO_FAUCET_PRIVATE_KEY
op read "op://Mezo DevOps/faucet_turnstile_site_key/notes" | wrangler secret put TURNSTILE_SITE_KEY
op read "op://Mezo DevOps/faucet_turnstile_secret_key/notes" | wrangler secret put TURNSTILE_SECRET_KEY

# Step 2: Deploy the Worker.
wrangler deploy src/index.ts