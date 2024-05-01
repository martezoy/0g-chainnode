#!/bin/bash

ROOT_DIR=${ROOT_DIR:-testnet}
CHAIN_ID=${CHAIN_ID:-zgtendermint_16600-1}

# Usage: init-genesis.sh IP1,IP2,IP3 KEYRING_PASSWORD
OS_NAME=`uname -o`
USAGE="Usage: ${BASH_SOURCE[0]} IP1,IP2,IP3"
if [[ "$OS_NAME" = "GNU/Linux" ]]; then
    USAGE="$USAGE KEYRING_PASSWORD"
fi

if [[ $# -eq 0 ]]; then
    echo "IP list not specified"
    echo $USAGE
    exit 1
fi

if [[ "$OS_NAME" = "GNU/Linux" ]]; then
    if [[ $# -eq 1 ]]; then
        echo "Keyring password not specified"
        echo $USAGE
        exit 1
    fi

    PASSWORD=$2
fi

0gchaind version 2>/dev/null || export PATH=$PATH:$(go env GOPATH)/bin

set -e

IFS=","; declare -a IPS=($1); unset IFS

NUM_NODES=${#IPS[@]}
VLIDATOR_BALANCE=15000000000000000000000000neuron
FAUCET_BALANCE=40000000000000000000000000neuron
STAKING=10000000000000000000000000neuron

# Init configs
for ((i=0; i<$NUM_NODES; i++)) do
    HOMEDIR="$ROOT_DIR"/node$i
    
    # Change parameter token denominations to neuron
    GENESIS="$HOMEDIR"/config/genesis.json
    TMP_GENESIS="$HOMEDIR"/config/tmp_genesis.json

    # Init
    0gchaind init "node$i" --home "$HOMEDIR" --chain-id "$CHAIN_ID" >/dev/null 2>&1

    # Replace stake with neuron
    sed -in-place='' 's/stake/neuron/g' "$GENESIS"

    # Replace the default evm denom of aphoton with neuron
    sed -in-place='' 's/aphoton/neuron/g' "$GENESIS"

    cat $GENESIS | jq '.consensus_params.block.max_gas = "25000000"' >$TMP_GENESIS && mv $TMP_GENESIS $GENESIS

    # Zero out the total supply so it gets recalculated during InitGenesis
    cat $GENESIS | jq '.app_state.bank.supply = []' >$TMP_GENESIS && mv $TMP_GENESIS $GENESIS

    # Disable fee market
    cat $GENESIS | jq '.app_state.feemarket.params.no_base_fee = true' >$TMP_GENESIS && mv $TMP_GENESIS $GENESIS

    # Disable london fork
    cat $GENESIS | jq '.app_state.evm.params.chain_config.london_block = null' >$TMP_GENESIS && mv $TMP_GENESIS $GENESIS
    cat $GENESIS | jq '.app_state.evm.params.chain_config.arrow_glacier_block = null' >$TMP_GENESIS && mv $TMP_GENESIS $GENESIS
    cat $GENESIS | jq '.app_state.evm.params.chain_config.gray_glacier_block = null' >$TMP_GENESIS && mv $TMP_GENESIS $GENESIS
    cat $GENESIS | jq '.app_state.evm.params.chain_config.merge_netsplit_block = null' >$TMP_GENESIS && mv $TMP_GENESIS $GENESIS
    cat $GENESIS | jq '.app_state.evm.params.chain_config.shanghai_block = null' >$TMP_GENESIS && mv $TMP_GENESIS $GENESIS
    cat $GENESIS | jq '.app_state.evm.params.chain_config.cancun_block = null' >$TMP_GENESIS && mv $TMP_GENESIS $GENESIS

    # cat "$GENESIS" | jq '.app_state["staking"]["params"]["bond_denom"]="a0gi"' >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
    # cat "$GENESIS" | jq '.app_state["gov"]["params"]["min_deposit"][0]["denom"]="a0gi"' >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

    cat "$GENESIS" | jq '.app_state["staking"]["params"]["max_validators"]=200' >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
    cat "$GENESIS" | jq '.app_state["slashing"]["params"]["signed_blocks_window"]="1000"' >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

    cat "$GENESIS" | jq '.app_state["consensus_params"]["block"]["time_iota_ms"]="3000"' >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

    # Change app.toml
    APP_TOML="$HOMEDIR"/config/app.toml
    sed -i 's/minimum-gas-prices = "0neuron"/minimum-gas-prices = "1000000000neuron"/' "$APP_TOML"
    sed -i '/\[json-rpc\]/,/^\[/ s/enable = false/enable = true/' "$APP_TOML"
    sed -i '/\[json-rpc\]/,/^\[/ s/address = "127.0.0.1:8545"/address = "0.0.0.0:8545"/' "$APP_TOML"

    # Set evm tracer to json
    sed -in-place='' 's/tracer = ""/tracer = "json"/g' "$APP_TOML"

    # Enable full error trace to be returned on tx failure
    sed -in-place='' '/iavl-cache-size/a\
trace = true' "$APP_TOML"
done

# Update seeds in config.toml
SEEDS=""
for ((i=0; i<$NUM_NODES; i++)) do
    if [[ $i -gt 0 ]]; then SEEDS=$SEEDS,; fi
    NODE_ID=`0gchaind tendermint show-node-id --home $ROOT_DIR/node$i`
    SEEDS=$SEEDS$NODE_ID@${IPS[$i]}:26656
done

for ((i=0; i<$NUM_NODES; i++)) do
    sed -i "/seeds = /c\seeds = \"$SEEDS\"" "$ROOT_DIR"/node$i/config/config.toml
done

# Prepare validators
#
# Note, keyring backend `file` works bad on Windows, and `add-genesis-account`
# do not supports --keyring-dir flag. As a result, we use keyring backend `os`,
# which is the default value.
#
# Where key stored:
# - Windows: Windows credentials management.
# - Linux: under `--home` specified folder.
if [[ "$OS_NAME" = "Msys" ]]; then
    for ((i=0; i<$NUM_NODES; i++)) do
        VALIDATOR="0gchain_validator_$i"
        set +e
        ret=`0gchaind keys list --keyring-backend os -n | grep $VALIDATOR`
        set -e
        if [[ "$ret" = "" ]]; then
            echo "Create validator key: $VALIDATOR"
            0gchaind keys add $VALIDATOR --keyring-backend os --eth
        fi
    done
elif [[ "$OS_NAME" = "GNU/Linux" ]]; then
    # Create N validators for node0
    for ((i=0; i<$NUM_NODES; i++)) do
        yes $PASSWORD | 0gchaind keys add "0gchain_validator_$i" --keyring-backend os --home "$ROOT_DIR"/node0 --eth
    done

    # Copy validators to other nodes
    for ((i=1; i<$NUM_NODES; i++)) do
        cp "$ROOT_DIR"/node0/keyhash "$ROOT_DIR"/node$i
        cp "$ROOT_DIR"/node0/*.address "$ROOT_DIR"/node$i
        cp "$ROOT_DIR"/node0/*.info "$ROOT_DIR"/node$i
    done
else
    echo -e "\n\nOS: $OS_NAME"
    echo "Unsupported OS to generate keys for validators!!!"
    exit 1
fi

# Add all validators in genesis
for ((i=0; i<$NUM_NODES; i++)) do
    for ((j=0; j<$NUM_NODES; j++)) do
        if [[ "$OS_NAME" = "GNU/Linux" ]]; then
            yes $PASSWORD | 0gchaind add-genesis-account "0gchain_validator_$j" $VLIDATOR_BALANCE --home "$ROOT_DIR/node$i"
        else
            0gchaind add-genesis-account "0gchain_validator_$j" $VLIDATOR_BALANCE --home "$ROOT_DIR/node$i"
        fi
    done
    0gchaind add-genesis-account 0g17n8707c20e8gge2tk2gestetjcs4536p4fhqcs $FAUCET_BALANCE --home "$ROOT_DIR/node$i"
done

# Prepare genesis txs
mkdir -p "$ROOT_DIR"/gentxs
for ((i=0; i<$NUM_NODES; i++)) do
    if [[ "$OS_NAME" = "GNU/Linux" ]]; then
        yes $PASSWORD | 0gchaind gentx "0gchain_validator_$i" $STAKING --home "$ROOT_DIR/node$i" --output-document "$ROOT_DIR/gentxs/node$i.json"
    else
        0gchaind gentx "0gchain_validator_$i" $STAKING --home "$ROOT_DIR/node$i" --output-document "$ROOT_DIR/gentxs/node$i.json"
    fi
done

# Create genesis at node0 and copy to other nodes
0gchaind collect-gentxs --home "$ROOT_DIR/node0" --gentx-dir "$ROOT_DIR/gentxs" >/dev/null 2>&1
0gchaind validate-genesis --home "$ROOT_DIR/node0"
for ((i=1; i<$NUM_NODES; i++)) do
    cp "$ROOT_DIR"/node0/config/genesis.json "$ROOT_DIR"/node$i/config/genesis.json
done

# For linux, backup keys for all validators
if [[ "$OS_NAME" = "GNU/Linux" ]]; then
    mkdir -p "$ROOT_DIR"/keyring-os

    cp "$ROOT_DIR"/node0/keyhash "$ROOT_DIR"/keyring-os
    cp "$ROOT_DIR"/node0/*.address "$ROOT_DIR"/keyring-os
    cp "$ROOT_DIR"/node0/*.info "$ROOT_DIR"/keyring-os

    for ((i=0; i<$NUM_NODES; i++)) do
        rm -f "$ROOT_DIR"/node$i/keyhash "$ROOT_DIR"/node$i/*.address "$ROOT_DIR"/node$i/*.info
    done
fi

echo -e "\n\nSucceeded to init genesis!\n"
