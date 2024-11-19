#!/bin/bash
### This is a deployment script for mezo validator stack
### For now it's created for debian-based systems and amd64 architecture (x86_64)
### Script handles:
### 1. Installing required packages using apt package manager
### 2. Building and installing mezod binary from the source
### 3. Installing connect sidecar
### 4. Deploying mezo validator stack as systemd services

set -eo pipefail

update_system() {
    apt update -y &&  apt upgrade -y
}

install_tools() {
    apt install ufw jq curl -y
}

open_ports() {
    ufw --force enable
    ufw allow 26660,26656,26657,1317,9090,8545,8546,6065/tcp
    # allow ssh connections:
    ufw allow 22/tcp
}
install_mezo() {
    MEZOD_DESTINATION=$MEZOD_HOME/bin/mezod-${MEZOD_VERSION}
    MEZO_EXEC=$MEZOD_DESTINATION/mezod

    mkdir -p ${MEZOD_DESTINATION}
    mkdir -p ./tmp

    echo "Downloading mezod package to temporary dir"
    
    url=$(curl --silent "https://api.github.com/repos/mezo-org/mezod/releases" \
        --header "Authorization: token ${GITHUB_TOKEN}" \
        | jq --arg MEZOD_VERSION "$MEZOD_VERSION" --arg MEZOD_ARCH "$MEZOD_ARCH" \
              '.[] | select(.name == $MEZOD_VERSION) | .assets[] | select(.name == ("mezod-" + $MEZOD_ARCH + ".tar.gz")) | .url' \
        | tr -d '"')

    if [ -z "$url" ]; then
        echo "Error: URL is empty. Exiting."
        exit 1
    fi

    echo "Download URL: $url"

    curl --verbose --silent --location --header "Authorization: token ${GITHUB_TOKEN}" \
        --header "Accept: application/octet-stream" \
        --output ./tmp/mezod-${MEZOD_ARCH}.tar.gz $url

    if [ $? -ne 0 ]; then
        echo "Error: curl command failed during download."
        exit 1
    fi

    if [ ! -f ./tmp/mezod-${MEZOD_ARCH}.tar.gz ]; then
        echo "Error: Downloaded file does not exist."
        exit 1
    fi

    echo "Unpacking the binary build ${MEZOD_VERSION}"
    tar -xvf ./tmp/mezod-${MEZOD_ARCH}.tar.gz -C ${MEZOD_DESTINATION}

    chown root:root ${MEZO_EXEC}
    chmod +x ${MEZO_EXEC}

    echo "Mezo binary installed with path: ${MEZO_EXEC}"
}

install_skip() {

    curl -ksSL https://raw.githubusercontent.com/skip-mev/connect/main/scripts/install.sh | bash

    CONNECT_TMP=$(which connect)
    CONNECT_VERSION=$(${CONNECT_TMP} version)
    
    CONNECT_EXEC_PATH=$MEZOD_HOME/bin/skip-${CONNECT_VERSION}
    CONNECT_EXEC=$CONNECT_EXEC_PATH/connect

    mkdir -p $CONNECT_EXEC_PATH

    mv $CONNECT_TMP $CONNECT_EXEC_PATH
    rm -rf $CONNECT_TMP

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
   test -f "${MEZOD_HOME}/keyring-file/keyhash" && {
    echo "Keyring already prepared!"
    return
  }

  mnemonic_file="./tmp/mnemonic.txt"
  gen_mnemonic "${mnemonic_file}"
  read -r keyring_mnemonic < "${mnemonic_file}"

  echo "Prepare keyring..."
  (echo "${keyring_mnemonic}"; echo "${MEZOD_KEYRING_PASSWORD}"; echo "${MEZOD_KEYRING_PASSWORD}") \
    | ${MEZO_EXEC} keys add \
      "${MEZOD_KEYRING_NAME}" \
      --home="${MEZOD_HOME}" \
      --keyring-backend="file" \
      --recover
  echo "Keyring prepared!"
}

init_mezo_config() {

    prepare_keyring
    
    echo "Initialize configuration..."
    ${MEZO_EXEC} \
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
    test -f "${client_config_file}.bak" ||  cat "$client_config_file" | tee "${client_config_file}.bak" > /dev/null
    echo "Backup ${app_config_file} to ${app_config_file}.bak"
    test -f "${app_config_file}.bak" ||  cat "$app_config_file" | tee "${app_config_file}.bak" > /dev/null
    echo "Backup ${config_file} to ${config_file}.bak"
    test -f "${config_file}.bak" ||  cat "$config_file" | tee "${config_file}.bak" > /dev/null

    echo "Customize configuration..."

    ${MEZO_EXEC} toml set \
        ${client_config_file} \
        -v chain-id="${MEZOD_CHAIN_ID}" \
        -v keyring-backend="file" \
        -v node="tcp://0.0.0.0:26657"

    ${MEZO_EXEC} toml set \
        ${config_file} \
        -v moniker="${MEZOD_MONIKER}" \
        -v p2p.laddr="tcp://0.0.0.0:26656" \
        -v rpc.laddr="tcp://0.0.0.0:26657" \
        -v instrumentation.prometheus=true \
        -v instrumentation.prometheus_listen_addr="0.0.0.0:26660" \
        -v p2p.external_address="${MEZOD_PUBLIC_IP}:26656" \
        -v consensus.timeout_propose="30s" \
        -v consensus.timeout_propose_delta="5s" \
        -v consensus.timeout_prevote="10s" \
        -v consensus.timeout_prevote_delta="5s" \
        -v consensus.timeout_precommit="5s" \
        -v consensus.timeout_precommit_delta="5s" \
        -v consensus.timeout_commit="150s" \
        -v rpc.timeout_broadcast_tx_commit="150s"

    ${MEZO_EXEC} toml set \
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

init_genval() {
  test -f "${MEZOD_HOME}"/config/genval/genval-*.json && {
    echo "Genval already exists!"
    return
  }

  echo "Prepare genval..."
  echo "${MEZOD_KEYRING_PASSWORD}" \
    | ${MEZO_EXEC} genesis genval \
      "${MEZOD_KEYRING_NAME}" \
      --keyring-backend="file" \
      --chain-id="${MEZOD_CHAIN_ID}" \
      --home="${MEZOD_HOME}" \
      --ip="${MEZOD_PUBLIC_IP}"

  echo "Genval prepared!"
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
WantedBy=multi-user.target" | tee /etc/systemd/system/connect-sidecar.service

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
WantedBy=multi-user.target" | tee /etc/systemd/system/ethereum-sidecar.service

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
WantedBy=multi-user.target" | tee /etc/systemd/system/mezo.service

}

systemd_restart() {
    echo "Reloading systemd daemon"
    systemctl daemon-reload
    echo "Starting systemd services"
    systemctl start mezo
    systemctl start ethereum-sidecar
    systemctl start connect-sidecar
}

cleanup() {
    echo "Stopping mezo.service"
    systemctl stop mezo.service || echo 'mezo stopped'
    echo "Stopping ethereum-sidecar.service"
    systemctl stop ethereum-sidecar.service || echo 'ethereum sidecar stopped'
    echo "Stopping connect-sidecar.service"
    systemctl stop connect-sidecar.service || echo 'skip sidecar stopped'

    echo "Disabling systemd unit mezo.service"
    systemctl disable mezo.service || echo 'mezo sidecar already disabled'
    echo "Disabling systemd unit ethereum-sidecar.service"
    systemctl disable ethereum-sidecar.service || echo 'ethereum already disabled'
    echo "Disabling systemd unit connect-sidecar.service"
    systemctl disable connect-sidecar.service || echo 'skip sidecar already disabled'

    echo "Removing systemd file /etc/systemd/system/mezo.service"
    rm -f /etc/systemd/system/mezo.service
    echo "Removing systemd file /etc/systemd/system/ethereum-sidecar.service"
    rm -f /etc/systemd/system/ethereum-sidecar.service
    echo "Removing systemd file /etc/systemd/system/connect-sidecar.service"
    rm -f /etc/systemd/system/connect-sidecar.service

    echo "Reloading systemd daemon"
    systemctl daemon-reload

    rm -rf ${MEZOD_HOME}
}

backup() {
    if [ ! -d "$MEZOD_HOME" ]; then
        echo "Error: Directory $MEZOD_HOME does not exist."
        exit 1
    fi
    
    BACKUP_DIRNAME=$(dirname "$MEZOD_HOME")
    BACKUP_FOLDER=$(basename "$MEZOD_HOME")

    echo "Trying to create directory for backups"
    mkdir -p "$MEZOD_HOME-backups"

    echo "Creating a backup of $MEZOD_HOME to $MEZOD_HOME-backups/mezo_backup_$(date +%Y%m%d).tar.gz"
    tar -czvf "$MEZOD_HOME-backups/mezo_backup_$(date +%Y%m%d).tar.gz" -C "$BACKUP_DIRNAME" "$BACKUP_FOLDER"
}

usage() {
    echo -e "This is a Mezo Validator Kit Native installation script."
    echo -e "Script handles installation of the validator software as native binaries managed by systemd services.\n"
    echo -e "Steps executed during installation:"
    echo -e "1. Update system (apt package manager)"
    echo -e "2. Install tools required for script (ufw jq curl)"
    echo -e "3. Open firewall ports (ufw)"
    echo -e "4. Install mezo binary"
    echo -e "5. Install connect-sidecar binary"
    echo -e "6. Configure Mezo - keyring, configuration files"
    echo -e "7. Setup systemd services for Mezo\n"

    echo -e "Usage: $0\n\n" \
    "\t[-r/--run]\n\t\trun the installation\n\n" \
    "\t[-b/--backup]\n\t\tbackup mezo home dir to ${MEZOD_HOME}-backups\n\n" \
    "\t[-c/--cleanup]\n\t\tclean up the installation\n\t\tWARNING: this option removes whole Mezo directory (${MEZOD_HOME}) INCLUDING PRIVATE KEYS\n\n" \
    "\t[--health]\n\t\tcheck health of mezo systemd services\n\n" \
    "\t[-s/--show-variables]\n\t\toutput variables read from env files\n\n" \
    "\t[-e/--envfile]\n\t\tset file with environment variables for setup script\n\n" \
    "\t[-h/--help]\n\t\tshow this prompt\n" 
}

healthcheck() {
    systemctl status --no-pager mezo || echo "issues with mezo"
    systemctl status --no-pager ethereum-sidecar || echo "issues with ethereum sidecar"
    systemctl status --no-pager connect-sidecar || echo "issues with connect sidecar"
}

show_variables() {
    ### Application ###
    echo "MEZOD_CHAIN_ID $MEZOD_CHAIN_ID"
    echo "MEZOD_HOME $MEZOD_HOME"
    echo "MEZOD_MONIKER $MEZOD_MONIKER"

    ### Keyring ###
    echo "MEZOD_KEYRING_NAME $MEZOD_KEYRING_NAME"
    echo "MEZOD_KEYRING_DIR $MEZOD_KEYRING_DIR"
    echo "MEZOD_KEYRING_PASSWORD $MEZOD_KEYRING_PASSWORD"

    ### Other ###
    echo "MEZOD_ETHEREUM_SIDECAR_CLIENT_SERVER_ADDRESS $MEZOD_ETHEREUM_SIDECAR_CLIENT_SERVER_ADDRESS"
    echo "MEZOD_ETHEREUM_SIDECAR_SERVER_ETHEREUM_NODE_ADDRESS $MEZOD_ETHEREUM_SIDECAR_SERVER_ETHEREUM_NODE_ADDRESS"
    echo "MEZOD_LOG_FORMAT $MEZOD_LOG_FORMAT"
    echo "MEZOD_KEY_NAME $MEZOD_KEY_NAME"
    echo "MEZOD_LOGLEVEL $MEZOD_LOGLEVEL"

    echo "CONNECT_SIDECAR_PORT $CONNECT_SIDECAR_PORT"

    ### Setup ###
    echo "MEZOD_VERSION $MEZOD_VERSION"
    echo "MEZOD_ARCH $MEZOD_ARCH"
    echo "MEZOD_PUBLIC_IP $MEZOD_PUBLIC_IP"

    ### Github ###
    echo "GITHUB_TOKEN $GITHUB_TOKEN"

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
    init_genval
    setup_systemd_skip
    setup_systemd_sidecar
    setup_systemd_mezo
    systemd_restart
}

setenvs() {
    echo "Reading configuration from environment files"
    . ${ENVIRONMENT_FILE}
}

# default env file name - can be changed through -e/--envfile option
ENVIRONMENT_FILE="testnet.env"
healthcheck_flag=false
run_flag=false
show_variables_flag=false
cleanup_flag=false
backup_flag=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --health)
            healthcheck_flag=true
            shift
            ;;
       -s|--show-variables)
            show_variables_flag=true
            shift
            ;;
        -e|--envfile)
            ENVIRONMENT_FILE="$2"
            shift 2
            ;;
        -c|--cleanup)
            cleanup_flag=true
            shift
            ;;
        -b|--backup)
            backup_flag=true
            shift
            ;;
        -h|--help)
            setenvs
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

if [ $(id -u) -ne 0 ]; then
    echo "This script requires root privileges"
    exit 1
fi

# Execute actions based on flags
if [[ "$healthcheck_flag" == true ]]; then
    healthcheck
    exit 0
fi

if [[ "$show_variables_flag" == true ]]; then
    setenvs
    show_variables
fi

if [[ "$backup_flag" == true ]]; then
    setenvs
    backup
    exit 0
fi

if [[ "$cleanup_flag" == true ]]; then
    setenvs
    cleanup
    exit 0
fi

setenvs
main
