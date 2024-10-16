#!/bin/bash
### This is a deployment script for mezo validator stack
### For now it's created for debian-based systems and amd64 architecture (x86_64)
### Script handles:
### 1. Installing required packages using apt package manager
### 2. Building and installing mezod binary from the source
### 3. Installing connect sidecar
### 4. Deploying mezo validator stack as systemd services

set -euo pipefail

update_system() {
    sudo apt update -y && sudo apt upgrade -y
}

install_tools() {
    sudo apt install wget git ufw make gcc jq bc yq -y
}

open_ports() {
    sudo ufw --force enable
    sudo ufw allow 26660,26656,26657,1317,9090,8545,8546,6065/tcp
    # allow ssh connections:
    sudo ufw allow 22/tcp
}

install_mezo() {

    MEZOD_DESTINATION=$MEZOD_HOME/bin/mezod-${MEZOD_VERSION}
    MEZO_EXEC=$MEZOD_DESTINATION/mezod

    sudo mkdir -p ${MEZOD_DESTINATION}
    
    echo "Downloading mezod package to temporary dir"
    # gh release download ${MEZOD_VERSION} -D ${MEZOD_TMP} --skip-existing
    # sudo wget -P ${MEZOD_DOWNLOADS} https://github.com/mezo-org/mezod/releases/download/v0.1.0/mezod.tar.gz
    url=$(curl --silent "https://api.github.com/repos/mezo-org/mezod/releases" \
        --header "Authorization: token ${GITHUB_TOKEN}" \
        | jq --arg MEZOD_VERSION "$MEZOD_VERSION" --arg MEZOD_ARCH "$MEZOD_ARCH" '.[] | select(.name == $MEZOD_VERSION) | .assets[] | select(.name == ("mezod-" + $MEZOD_ARCH + ".tar.gz")) | .url' | tr -d '"')

    echo DOWNLOAD URL: $url

    curl \
        --silent \
        --location \
        --header "Authorization: token ${GITHUB_TOKEN}" \
        --header "Accept: application/octet-stream" \
        --output /tmp/mezod-${MEZOD_ARCH}.tar.gz $url

    # echo "Unpacking the build ${MEZOD_VERSION}"
    sudo tar -xvf /tmp/mezod-${MEZOD_ARCH}.tar.gz -C ${MEZOD_DESTINATION}

    # sudo mv /tmp/mezod ${MEZOD_DESTINATION}

    sudo chown root:root ${MEZO_EXEC}
    sudo chmod +x ${MEZO_EXEC}

    echo "Mezo binary installed with path: ${MEZO_EXEC}"

    sudo $MEZO_EXEC --help

}

install_skip() {

    curl -sSL https://raw.githubusercontent.com/skip-mev/connect/main/scripts/install.sh | sudo bash

    CONNECT_TMP=$(which connect)
    CONNECT_VERSION=$(${CONNECT_TMP} version)
    
    CONNECT_EXEC_PATH=$MEZOD_HOME/bin/skip-${CONNECT_VERSION}
    CONNECT_EXEC=$CONNECT_EXEC_PATH/connect

    sudo mkdir -p $CONNECT_EXEC_PATH

    sudo mv $CONNECT_TMP $CONNECT_EXEC_PATH
    sudo rm -rf $CONNECT_TMP

    echo "Skip binary installed with path: ${CONNECT_EXEC_PATH}"
}


gen_mnemonic() {
  mnemonic_file="$1"
  # Check if the environment variable is set
  if [ -n "$MEZOD_KEYRING_MNEMONIC" ]; then
    echo "Using KEYRING_MNEMONIC from the environment"
    echo "$MEZOD_KEYRING_MNEMONIC" > "$mnemonic_file"
  else
    # Ask the user to generate a new mnemonic
    printf "Do you want to generate a new mnemonic? [y/N]: "; read -r response
    case "$response" in
      [yY])
        echo "Generating a new mnemonic..."
        m=$(${MEZO_EXEC} keys mnemonic)
        echo "$m" > "$mnemonic_file"
        printf "\n%s\n%s\n\n" "Generated mnemonic (make backup!):" "$m"
        echo "Press any key to continue..."; read -r _
        ;;
      *)
        # Ask the user to enter the mnemonic
        printf "Enter the mnemonic: "; read -r mnemonic
        echo "$mnemonic" > "$mnemonic_file"
        ;;
    esac
  fi
}

prepare_keyring() {
  sudo test -f "${MEZOD_HOME}/keyring-file/keyhash" && {
    echo "Keyring already prepared!"
    return
  }

  mnemonic_file="/tmp/mnemonic.txt"
  gen_mnemonic "${mnemonic_file}"
  read -r keyring_mnemonic < "${mnemonic_file}"

  echo "Prepare keyring..."
  (echo "${keyring_mnemonic}"; echo "${MEZOD_KEYRING_PASSWORD}"; echo "${MEZOD_KEYRING_PASSWORD}") \
    | sudo ${MEZO_EXEC} keys add \
      "${MEZOD_KEYRING_NAME}" \
      --home="${MEZOD_HOME}" \
      --keyring-backend="file" \
      --recover
  echo "Keyring prepared!"
}

init_mezo_config() {

    prepare_keyring
    
    echo "Initialize configuration..."
    sudo ${MEZO_EXEC} \
        init \
        "${MEZOD_MONIKER}" \
        --chain-id="${MEZOD_CHAIN_ID}" \
        --home="${MEZOD_HOME}" \
        --keyring-backend="file" \
        --overwrite
    echo "Configuration initialized!"
}

configure_mezo() {
    
    client_config_file="${MEZOD_HOME}/config/client.toml"
    app_config_file="${MEZOD_HOME}/config/app.toml"
    config_file="${MEZOD_HOME}/config/config.toml"

    echo "Backup original configuration..."
    echo "Backup ${client_config_file} to ${client_config_file}.bak"
    test -f "${client_config_file}.bak" || sudo cat "$client_config_file" | sudo tee "${client_config_file}.bak" > /dev/null
    echo "Backup ${app_config_file} to ${app_config_file}.bak"
    test -f "${app_config_file}.bak" || sudo cat "$app_config_file" | sudo tee "${app_config_file}.bak" > /dev/null
    echo "Backup ${config_file} to ${config_file}.bak"
    test -f "${config_file}.bak" || sudo cat "$config_file" | sudo tee "${config_file}.bak" > /dev/null

    echo "Customize configuration..."

    sudo ${MEZO_EXEC} toml set \
        ${client_config_file} \
        -v chain-id="${MEZOD_CHAIN_ID}" \
        -v keyring-backend="file" \
        -v node="tcp://0.0.0.0:26657"

    sudo ${MEZO_EXEC} toml set \
        ${config_file} \
        -v moniker="${MEZOD_MONIKER}" \
        -v p2p.laddr="tcp://0.0.0.0:26656" \
        -v rpc.laddr="tcp://0.0.0.0:26657" \
        -v instrumentation.prometheus=true \
        -v instrumentation.prometheus_listen_addr="0.0.0.0:26660" \
        -v p2p.external_address="${MEZOD_PUBLIC_IP}:26656"

    sudo ${MEZO_EXEC} toml set \
        ${app_config_file} \
        -v ethereum-sidecar.client.server-address="0.0.0.0:7500" \
        -v api.enable=true \
        -v api.address="tcp://0.0.0.0:1317" \
        -v grpc.enable=true \
        -v grpc.address="0.0.0.0:9090" \
        -v grpc-web.enable=true \
        -v json-rpc.enable=true \
        -v json-rpc.address="0.0.0.0:8545" \
        -v json-rpc.api="eth,txpool,personal,net,debug,web3" \
        -v json-rpc.ws-address="0.0.0.0:8546" \
        -v json-rpc.metrics-address="0.0.0.0:6065"

}

setup_systemd_skip(){

    echo "
[Unit]
Description=Connect Sidecar Service
After=network.target

[Service]
Restart=no
ExecStartPre=/bin/echo "Starting connect-sidecar systemd initialization..."
ExecStart=${CONNECT_EXEC} --log-disable-file-rotation --port=${CONNECT_SIDECAR_PORT} --market-map-endpoint=\"127.0.0.1:9090\"
StandardOutput=journal
StandardError=journal
User=root

[Install]
WantedBy=multi-user.target" | sudo tee /etc/systemd/system/connect-sidecar.service

}

setup_systemd_sidecar(){
    echo "
[Unit]
Description=Ethereum Sidecar Service
After=network.target

[Service]
Restart=no
ExecStartPre=/bin/echo "Starting ethereum-sidecar systemd initialization..."
ExecStart=${MEZO_EXEC} ethereum-sidecar --log_format=${MEZOD_LOG_FORMAT} --ethereum-sidecar.server.ethereum-node-address=${MEZOD_ETHEREUM_SIDECAR_SERVER_ETHEREUM_NODE_ADDRESS}
StandardOutput=journal
StandardError=journal
User=root

[Install]
WantedBy=multi-user.target" | sudo tee /etc/systemd/system/ethereum-sidecar.service

}

setup_systemd_mezo(){

    echo "
[Unit]
Description=Mezo Service
After=network.target

[Service]
Restart=no
ExecStartPre=/bin/echo "Starting mezod systemd initialization..."
ExecStart=${MEZO_EXEC} start --home=${MEZOD_HOME} --metrics
StandardOutput=journal
StandardError=journal
User=root

[Install]
WantedBy=multi-user.target" | sudo tee /etc/systemd/system/mezo.service

}

systemd_restart() {
    sudo systemctl daemon-reload
    sudo systemctl start mezo
    sudo systemctl start ethereum-sidecar
    sudo systemctl start connect-sidecar
}

cleanup() {
    sudo systemctl stop mezo.service || echo 'mezo stopped'
    sudo systemctl stop ethereum-sidecar.service || echo 'ethereum sidecar stopped'
    sudo systemctl stop connect-sidecar.service || echo 'skip sidecar stopped'

    sudo systemctl disable mezo.service || echo 'mezo sidecar already disabled'
    sudo systemctl disable ethereum-sidecar.service || echo 'ethereum already disabled'
    sudo systemctl disable connect-sidecar.service || echo 'skip sidecar already disabled'
    
    sudo rm -f /etc/systemd/system/mezo.service
    sudo rm -f /etc/systemd/system/ethereum-sidecar.service
    sudo rm -f /etc/systemd/system/connect-sidecar.service

    sudo systemctl daemon-reload

    sudo rm -rf ${MEZOD_HOME}
}

usage() {
  echo -e "\nUsage: $0\n\n" \
    "[-c/--cleanup] - clean up the installation\n\n" \
    "[--health] - check health of mezo systemd services\n\n" \
    "[-s/--show-variables] - output variables read from env files\n"
#   echo -e "\nRequired command line arguments:\n"
}

healthcheck() {
    sudo systemctl status --no-pager mezo || echo "issues with mezo"
    sudo systemctl status --no-pager ethereum-sidecar || echo "issues with ethereum sidecar"
    sudo systemctl status --no-pager connect-sidecar || echo "issues with connect sidecar"
}

show_variables() {
    echo "MEZOD_HOME: $MEZOD_HOME"
    echo "MEZOD_MONIKER: $MEZOD_MONIKER"
    echo "MEZOD_P2P_SEEDS: $MEZOD_P2P_SEEDS"
    echo "MEZOD_CHAIN_ID: $MEZOD_CHAIN_ID"
    echo "MEZOD_ETHEREUM_SIDECAR_CLIENT_SERVER_ADDRESS: $MEZOD_ETHEREUM_SIDECAR_CLIENT_SERVER_ADDRESS"
    echo "MEZOD_ETHEREUM_SIDECAR_SERVER_ETHEREUM_NODE_ADDRESS: $MEZOD_ETHEREUM_SIDECAR_SERVER_ETHEREUM_NODE_ADDRESS"
    echo "MEZOD_LOG_FORMAT: $MEZOD_LOG_FORMAT"
    echo "MEZOD_KEY_NAME: $MEZOD_KEY_NAME"

    echo "SETUP_GENESIS_URL: $SETUP_GENESIS_URL"

    echo "MEZOD_ARCH: $MEZOD_ARCH"
    echo "MEZOD_VERSION: $MEZOD_VERSION"
    echo "MEZOD_PUBLIC_IP: $MEZOD_PUBLIC_IP"

    set -x
}

main() {
    update_system
    install_tools
    open_ports
    install_mezo
    install_skip
    init_mezo_config
    configure_mezo
    setup_systemd_skip
    setup_systemd_sidecar
    setup_systemd_mezo
    systemd_restart
    # optionally encrypt temporary env variable
    # encrypt_env
}

setenvs() {
    echo "Reading configuration from environment files"
    . ${ENVIRONMENT_FILE}
    . .env
}

encrypt_env() {
    sudo openssl enc -aes-256-cbc -salt -in ${MEZOD_HOME}/startup/env -out ${MEZOD_HOME}/startup/env.enc -k ${MEZOD_KEYRING_PASSWORD}
    sudo rm -f ${MEZOD_HOME}/startup/env
}

decrypt_env() {
    sudo openssl enc -aes-256-cbc -d -in ${MEZOD_HOME}/startup/env.enc -out ${MEZOD_HOME}/startup/env.dec -k ${MEZOD_KEYRING_PASSWORD}
}

# default env file name - can be changed through -e/--envfile option
ENVIRONMENT_FILE="testnet.env"

while [[ $# -gt 0 ]]; do
    case $1 in
        --health)
            healthcheck
            exit 0
            shift
            ;;
        -s|--show-variables)
            setenvs
            show_variables
            shift
            ;;
        -e|--envfile)
            ENVIRONMENT_FILE="$2"
            shift
            shift
            ;;      
        -c|--cleanup)
            setenvs
            cleanup
            exit 0
            ;;
        -d|--decrypt-env)
            setenvs
            decrypt_env
            exit 0
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

setenvs
main
