#!/bin/sh

#
# This script starts the Mezod node.
#

set -o errexit # Exit on error
set -o nounset # Exit on use of an undefined variable

# Load predefined variables
# shellcheck disable=SC1091
. /vars.sh

# Start the Mezod node
(echo ${KEYRING_PASSWORD}; echo ${KEYRING_PASSWORD}) \
  | mezod start \
    --home=${MEZOD_HOME} \
    --metrics
