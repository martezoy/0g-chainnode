#!/bin/bash

function help() {
    echo "Usage: deploy.sh IP1,IP2,IP3 [options]"
    echo ""
    echo "  -i    Identity file"
    echo "  -k    Keyring password to create key (for Linux only)"
    echo "  -n    Network (default: testnet)"
    echo "  -c    Chain ID (default: \"zgtendermint_9000-1\")"
    echo ""
}

if [[ $# -eq 0 ]]; then
	help
	exit 1
fi

set -e

IP_LIST=$1
shift
PEM_FLAG=""
KEYRING_PASSWORD=""
NETWORK="testnet"
INIT_GENESIS_ENV=""

while [[ $# -gt 0 ]]; do
	case $1 in
	-i)
		PEM_FLAG="-i $2";
        shift; shift
		;;
	-k)
		KEYRING_PASSWORD=$2;
        shift; shift
		;;
    -n)
        NETWORK=$2
		INIT_GENESIS_ENV="$INIT_GENESIS_ENV export ROOT_DIR=$2;"
        shift; shift
		;;
    -c)
		INIT_GENESIS_ENV="$INIT_GENESIS_ENV export CHAIN_ID=$2;"
        shift; shift
		;;
	*)
        help
        echo "Unknown flag passed: \"$1\""
		exit 1
		;;
    esac
done

IFS=","; declare -a IPS=($IP_LIST); unset IFS
NUM_NODES=${#IPS[@]}

# Install dependent libraries and binary
for ((i=0; i<$NUM_NODES; i++)) do
    ssh $PEM_FLAG ubuntu@${IPS[$i]} "git clone https://github.com/0glabs/0g-chain.git; cd 0g-chain; ./networks/testnet/install.sh"
done

# Create genesis config on node0
ssh $PEM_FLAG ubuntu@${IPS[0]} "cd 0g-chain/networks/testnet; $INIT_GENESIS_ENV ./init-genesis.sh $IP_LIST $KEYRING_PASSWORD; tar czf ~/$NETWORK.tar.gz $NETWORK; rm -rf $NETWORK"
scp $PEM_FLAG ubuntu@${IPS[0]}:$NETWORK.tar.gz .
ssh $PEM_FLAG ubuntu@${IPS[0]} "rm $NETWORK.tar.gz"

# Copy genesis config to remote nodes
tar xzf $NETWORK.tar.gz
rm $NETWORK.tar.gz
cd $NETWORK
for ((i=0; i<$NUM_NODES; i++)) do
    tar czf node$i.tar.gz node$i
    scp $PEM_FLAG node$i.tar.gz ubuntu@${IPS[$i]}:~
    ssh $PEM_FLAG ubuntu@${IPS[$i]} "rm -rf kava-prod; tar xzf node$i.tar.gz; rm node$i.tar.gz; mv node$i kava-prod"
    rm node$i.tar.gz
done

echo -e "\n\nSucceeded to deploy on $NUM_NODES nodes!\n"