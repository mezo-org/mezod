#!/bin/bash
### This is a deployment script for mezo validator stack
### For now it's created for debian-based systems
### Script handles:
### 1. Installing required packages using apt package manager
### 2. Installing golang in the specified MEZOD_GO_
### 3. Building and installing mezod binary from the source
### 4. Installing skip sidecar
### 5. Deploying mezo validator stack as systemd services

set -euxo pipefail

. testnet.env
. .env

echo "MEZOD_HOME: $MEZOD_HOME"
echo "MEZOD_MONIKER: $MEZOD_MONIKER"
echo "MEZOD_KEYRING_BACKEND: $MEZOD_KEYRING_BACKEND"
echo "MEZOD_P2P_SEEDS: $MEZOD_P2P_SEEDS"
echo "MEZOD_CHAIN_ID: $MEZOD_CHAIN_ID"
echo "MEZOD_ETHEREUM_SIDECAR_CLIENT_SERVER_ADDRESS: $MEZOD_ETHEREUM_SIDECAR_CLIENT_SERVER_ADDRESS"
echo "MEZOD_ETHEREUM_SIDECAR_SERVER_ETHEREUM_NODE_ADDRESS: $MEZOD_ETHEREUM_SIDECAR_SERVER_ETHEREUM_NODE_ADDRESS"
echo "MEZOD_LOG_FORMAT: $MEZOD_LOG_FORMAT"
echo "MEZOD_KEY_NAME: $MEZOD_KEY_NAME"

echo "SETUP_GENESIS_URL: $SETUP_GENESIS_URL"

echo "MEZOD_GO_VERSION: $MEZOD_GO_VERSION"
echo "MEZOD_ARCH: $MEZOD_ARCH"
echo "MEZOD_VERSION: $MEZOD_VERSION"


test_exit() {
    echo "test exit"
    exit 0
}

detect_os() {

    OS_TYPE=$(uname)
    ARCH=$(uname -p)

    if [[ "$OS_TYPE" == "Linux" ]]; then
        echo "This is a Linux operating system."
    elif [[ "$OS_TYPE" == "Darwin" ]]; then
        echo "This is macOS."
    elif [[ "$OS_TYPE" == "FreeBSD" ]]; then
        echo "This is FreeBSD."
        echo "For now, FreeBSD is not supported by mezo, exiting"
        exit 0
    else
        echo "Unknown operating system: $OS_TYPE"
    fi

    if [[ "$ARCH" == "x86_64" ]]; then
        ARCH="amd64"
    fi
    if [[ "$ARCH" == "aarch64" ]]; then
        ARCH="arm64"
        echo "For now, arm64 is not supported by mezo, exiting"
        exit 0
    fi
}

update_system() {
    sudo apt update -y && sudo apt upgrade -y
}

install_tools() {
    sudo apt install wget git ufw make gcc jq bc -y
}

open_ports() {
    sudo ufw --force enable
    sudo ufw allow 26656,26657,1317,9090,8545,8546/tcp
}

install_go() {

    GO_HOMEDIR=$MEZOD_HOME/bin/go-${MEZOD_GO_VERSION}
    MEZOD_DOWNLOADS="$MEZOD_HOME/downloads"

    sudo mkdir -p ${MEZOD_DOWNLOADS}
    sudo mkdir -p ${GO_HOMEDIR}

    if [ ! -f "${MEZOD_DOWNLOADS}/go${MEZOD_GO_VERSION}.linux-amd64.tar.gz" ]; then
        echo "Download Go binary"
        sudo wget https://go.dev/dl/go${MEZOD_GO_VERSION}.linux-amd64.tar.gz -P ${MEZOD_DOWNLOADS}

        echo "Extract Go binary to the mezod directory"
        sudo tar -C ${GO_HOMEDIR} -xzf ${MEZOD_DOWNLOADS}/go${MEZOD_GO_VERSION}.linux-amd64.tar.gz
    fi

    # echo Change ownership of the Go directory
    # sudo chown -R $(whoami) $HOME/go

    echo "Export Go paths temporarily for the script's runtime"

    export GOPATH=$GO_HOMEDIR/go/bin
    export GOBIN=$GO_HOMEDIR/go/bin
    export GOROOT=$GO_HOMEDIR/go
    export GOPROXY=https://proxy.golang.org,direct
    export PATH=$PATH:$GO_HOMEDIR/go/bin

    go env -w GOPROXY=https://proxy.golang.org,direct
    
    ${GOPATH}/go version

    # Optionally, inform the user to restart the terminal
    echo "Go has been installed in ${GOROOT}"
    echo "Go binary path is ${GOBIN}"
    echo "GOPATH is ${GOPATH}"
    echo "GOPROXY is ${GOPROXY}"

}

install_mezo() {

    MEZOD_DESTINATION=$MEZOD_HOME/bin/mezod-${MEZOD_VERSION}
    MEZO_EXEC=$MEZOD_DESTINATION/mezod

    sudo mkdir -p ${MEZOD_DESTINATION}
    
    echo "Downloading mezod package to temporary dir"
    # gh release download ${MEZOD_VERSION} -D ${MEZOD_TMP} --skip-existing
    # sudo wget -P ${MEZOD_DOWNLOADS} https://github.com/mezo-org/mezod/releases/download/v0.1.0/mezod.tar.gz
    url=$(curl --silent \
        --header "Authorization: token ${GITHUB_TOKEN}" \
        "https://api.github.com/repos/mezo-org/mezod/releases" | jq --arg MEZOD_VERSION "$MEZOD_VERSION" '.[] | select(.name == $MEZOD_VERSION) | .assets[] | select(.name == "mezod-amd64.tar.gz") | .url' | tr -d '"')

    curl \
        --silent \
        --location \
        --header "Authorization: token ${GITHUB_TOKEN}" \
        --header "Accept: application/octet-stream" \
        --output /tmp/mezod $url

    # echo "Unpacking the build ${MEZOD_VERSION}"
    # sudo tar -xvf /tmp/mezod-amd64.tar.gz -C ${MEZOD_DESTINATION}

    sudo mv /tmp/mezod ${MEZOD_DESTINATION}

    sudo chown root:root ${MEZO_EXEC}
    sudo chmod +x ${MEZO_EXEC}

    echo "Mezo binary installed with path: ${MEZO_EXEC}"

    sudo $MEZO_EXEC --help

    # echo "Removing temporary repo"
    # sudo rm -rf ${MEZOD_TMP}
}

install_skip() {
    SKIP_EXEC_PATH=$MEZOD_HOME/bin/skip/connect

    sudo mkdir -p $SKIP_EXEC_PATH
    curl -sSL https://raw.githubusercontent.com/skip-mev/connect/main/scripts/install.sh | sudo bash

    SKIP_TMP=$(which connect)

    sudo mv $SKIP_TMP $SKIP_EXEC_PATH
    echo "Skip binary installed with path: ${SKIP_EXEC_PATH}"
}

install_validator() {

    # Path variables
    CONFIG=$MEZOD_HOME/config/config.toml
    APP_TOML=$MEZOD_HOME/config/app.toml
    GENESIS=$MEZOD_HOME/config/genesis.json
    TMP_GENESIS=$MEZOD_HOME/config/tmp_genesis.json


    echo "Prepare keyring..."
    (echo $MEZOD_KEYRING_SEED; echo $MEZOD_KEYRING_PASSWORD; echo $MEZOD_KEYRING_PASSWORD) \
        | sudo ${MEZO_EXEC} keys add ${MEZOD_KEYRING_KEY_NAME} \
        --home=${MEZOD_HOME} \
        --keyring-backend=${MEZOD_KEYRING_BACKEND} \
        --keyring-dir=${MEZOD_KEYRING_DIR} \
        --recover
    
    sudo ${MEZO_EXEC} init $MEZOD_MONIKER \
        --chain-id=${MEZOD_CHAIN_ID} \
        --home=${MEZOD_HOME} \
        --keyring-backend=${MEZOD_KEYRING_BACKEND} \
        # --keyring-dir=${MEZOD_KEYRING_DIR}


    echo "" | sudo tee ${MEZOD_HOME}/config/genesis.json
    wget --output-document=/tmp/genesis.yaml ${SETUP_GENESIS_URL} || { echo "Genesis file not found!"; exit 1; }
    echo $(yq '.data["genesis.json"]' /tmp/genesis.yaml | sed -e 's/\\n/\n/g' -e 's/\\"/"/g' -e '1s/^"//' -e '$s/"$//' | jq) | sudo tee ${MEZOD_HOME}/config/genesis.json
    echo "Genesis file downloaded!"
    
    # Start the node (remove the --pruning=nothing flag if historical queries are not needed) TODO: move this to systemd
    # mezod start --metrics "$TRACE" --log_level $LOGLEVEL --minimum-gas-prices=0.0001abtc --json-rpc.api eth,txpool,personal,net,debug,web3 --api.enable --home "$HOMEDIR"
}

setup_systemd_skip(){
    echo "
[Unit]
Description=Skip Sidecar Service
After=network.target

[Service]
ExecStart=${SKIP_EXEC_PATH} --market-map-endpoint=\"127.0.0.1:8545\"
StandardOutput=journal
StandardError=journal
User=root

[Install]
WantedBy=multi-user.target" | sudo tee /etc/systemd/system/skip-sidecar.service

}

setup_systemd_sidecar(){
        echo "
[Unit]
Description=Ethereum Sidecar Service
After=network.target

[Service]
ExecStart=${MEZO_EXEC} ethereum-sidecar
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
ExecStart=${MEZO_EXEC} start --log_format=${MEZOD_LOG_FORMAT} --chain-id=${MEZOD_CHAIN_ID} --home=${MEZOD_HOME} --keyring-backend=${MEZOD_KEYRING_BACKEND} --moniker=${MEZOD_MONIKER} --p2p.seeds=${MEZOD_P2P_SEEDS} --ethereum-sidecar.client.server-address=${MEZOD_ETHEREUM_SIDECAR_CLIENT_SERVER_ADDRESS}
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
    sudo systemctl start skip-sidecar
}

cleanup() {
    sudo systemctl stop mezo.service || echo 'mezo stopped'
    sudo systemctl stop ethereum-sidecar.service || echo 'ethereum sidecar stopped'
    sudo systemctl stop skip-sidecar.service || echo 'skip sidecar stopped'

    sudo systemctl disable mezo.service || echo 'mezo sidecar already disabled'
    sudo systemctl disable ethereum-sidecar.service || echo 'ethereum already disabled'
    sudo systemctl disable skip-sidecar.service || echo 'skip sidecar already disabled'
    
    sudo rm -f /etc/systemd/system/mezo.service
    sudo rm -f /etc/systemd/system/ethereum-sidecar.service
    sudo rm -f /etc/systemd/system/skip-sidecar.service

    sudo systemctl daemon-reload

    echo -n "Do you want to remove go? (yY/nN)"
    read -r delete_go

    WHICHGO=${MEZOD_HOME}/go-${MEZOD_GO_VERSION}

    if [[ "$delete_go" == "y" || "$delete_go" == "Y" ]]; then
        echo "removing go..."
        sudo rm -rf $WHICHGO
    fi

    sudo rm -rf ${MEZOD_HOME}
}

if [[ "$CLEANUP" == "true" ]]; then
    cleanup
    exit 0
fi

# not using detect_os, setting architecture to amd64
# detect_os

update_system
install_tools
open_ports
install_go
install_mezo
install_skip
install_validator
setup_systemd_skip
setup_systemd_sidecar
setup_systemd_mezo
systemd_restart
