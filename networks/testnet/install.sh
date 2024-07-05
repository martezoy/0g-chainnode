#!/bin/bash

if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    # Install dependent libraries
    go version 2>/dev/null || sudo snap install go --classic
    jq --version 2>/dev/null || sudo snap install jq
    make --version 2>/dev/null || sudo apt install make -y
    gcc --version 2>/dev/null || (sudo apt-get update; sudo apt install gcc -y)
elif [[ "$OSTYPE" == "darwin"* ]]; then
     # Install dependent libraries
     go version 2>/dev/null || brew install go
     jq --version 2>/dev/null || brew install jq
     make --version 2>/dev/null || brew install make
     gcc --version 2>/dev/null || brew install gcc
fi

# Build binary
export PATH=$PATH:$(go env GOPATH)/bin
0gchaind version 2>/dev/null
if [[ $? -ne 0 ]]; then
    # Make under root dir
    SCRIPT_DIR=`dirname "${BASH_SOURCE[0]}"`
    cd $SCRIPT_DIR/../..
    rm -rf $(go env GOPATH)/bin/0gchaind
    make install
    # Add gopath to path
    echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.profile
fi
