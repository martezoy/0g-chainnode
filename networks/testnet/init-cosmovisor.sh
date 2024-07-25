#!/bin/bash

if [ -z "$1" ]; then
    echo "Usage: $0 <0G Home>"
    exit 1
fi

go install cosmossdk.io/tools/cosmovisor/cmd/cosmovisor@latest

export DAEMON_NAME=0gchaind
echo "export DAEMON_NAME=0gchaind" >> ~/.profile
export DAEMON_HOME=$1
echo "export DAEMON_HOME=$1" >> ~/.profile
cosmovisor init $(whereis -b 0gchaind | awk '{print $2}')
mkdir $DAEMON_HOME/cosmovisor/backup
echo "export DAEMON_DATA_BACKUP_DIR=$DAEMON_HOME/cosmovisor/backup" >> ~/.profile
echo "export DAEMON_ALLOW_DOWNLOAD_BINARIES=true" >> ~/.profile
