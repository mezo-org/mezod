#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e
# Enable aliases to be expanded as commands.
shopt -s expand_aliases

# Check if WRANGLER_ENV is set
if [ -z "$WRANGLER_ENV" ]; then
  printf "WRANGLER_ENV is not set\n"
  exit 1
fi

# Check if the user is authenticated with 1Password CLI. If not, the script will
# fail due to the `set -e` directive.
printf "\nchecking 1Password CLI authentication\n\n"
op whoami

alias wrangler='npx wrangler --env ${WRANGLER_ENV}'

printf "\nchecking Wrangler authentication\n\n"
wrangler whoami && wrangler versions list

# Step 1: Read secrets from 1Password and put them into the Worker environment.
op read "op://Mezo DevOps/faucet_private_key/notes" | wrangler secret put MEZO_FAUCET_PRIVATE_KEY
op read "op://Mezo DevOps/faucet_turnstile_site_key/notes" | wrangler secret put TURNSTILE_SITE_KEY
op read "op://Mezo DevOps/faucet_turnstile_secret_key/notes" | wrangler secret put TURNSTILE_SECRET_KEY
op read "op://Mezo DevOps/faucet_api_key/notes" | wrangler secret put API_KEY

# Step 2: Deploy the Worker.
wrangler deploy src/index.ts