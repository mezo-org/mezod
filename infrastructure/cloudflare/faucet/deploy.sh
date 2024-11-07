#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e
# Enable aliases to be expanded as commands.
shopt -s expand_aliases

# Check if the user is authenticated with 1Password CLI. If not, the script will
# fail due to the `set -e` directive.
printf "checking 1Password CLI authentication\n\n"
op whoami

alias wrangler='npx wrangler --env staging'

# Step 1: Read secrets from 1Password and put them into the Worker environment.
op read "op://Mezo DevOps/faucet_private_key/notes" | wrangler secret put MEZO_FAUCET_PRIVATE_KEY
op read "op://Mezo DevOps/faucet_turnstile_site_key/notes" | wrangler secret put TURNSTILE_SITE_KEY
op read "op://Mezo DevOps/faucet_turnstile_secret_key/notes" | wrangler secret put TURNSTILE_SECRET_KEY

# Step 2: Deploy the Worker.
wrangler deploy src/index.ts