#!/bin/bash

# Check if required arguments are provided
if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <rpc_url> <moniker>"
    exit 1
fi

RPC_URL=$1
MONIKER=$2
SECRET_NAME="metrics-scraper-config"
NAMESPACE="monitoring"

# Get the existing secret data
EXISTING_CONFIG=$(kubectl get secret $SECRET_NAME -n $NAMESPACE -o jsonpath='{.data.nodes-config\.json}' | base64 --decode)

# Create a temporary file with the existing config
TMP_FILE=$(mktemp)
echo "$EXISTING_CONFIG" > "$TMP_FILE"

# Add the new node using jq
jq --arg url "$RPC_URL" --arg moniker "$MONIKER" '.nodes += [{"rpc_url": $url, "moniker": $moniker}]' "$TMP_FILE" > "${TMP_FILE}.new"

# Update the secret by patching only the nodes-config.json key
NEW_CONFIG_B64=$(base64 < "${TMP_FILE}.new")
kubectl patch secret $SECRET_NAME -n $NAMESPACE --type='json' -p="[{\"op\": \"replace\", \"path\": \"/data/nodes-config.json\", \"value\": \"$NEW_CONFIG_B64\"}]"

# Clean up temporary files
rm "$TMP_FILE" "${TMP_FILE}.new"

# Restart the metrics-scraper deployment
kubectl rollout restart deployment metrics-scraper -n $NAMESPACE

echo "Successfully added new node with moniker: $MONIKER and restarted metrics-scraper deployment" 