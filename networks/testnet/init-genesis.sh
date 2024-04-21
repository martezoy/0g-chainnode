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

kava version 2>/dev/null || export PATH=$PATH:$(go env GOPATH)/bin

set -e

IFS=","; declare -a IPS=($1); unset IFS

NUM_NODES=${#IPS[@]}
VLIDATOR_BALANCE=20000000000000ukava
FAUCET_BALANCE=20000000000000ukava
STAKING=2000000000000ukava

# Init configs
for ((i=0; i<$NUM_NODES; i++)) do
    HOMEDIR="$ROOT_DIR"/node$i
    
    # Change parameter token denominations to neuron
    GENESIS="$HOMEDIR"/config/genesis.json
    TMP_GENESIS="$HOMEDIR"/config/tmp_genesis.json

    # Init
    kava init "node$i" --home "$HOMEDIR" --chain-id "$CHAIN_ID" >/dev/null 2>&1

    # Replace stake with ukava
    sed -in-place='' 's/stake/ukava/g' "$GENESIS"

    # Replace the default evm denom of aphoton with ukava
    sed -in-place='' 's/aphoton/akava/g' "$GENESIS"

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

    # Add earn vault
    cat $GENESIS | jq '.app_state.earn.params.allowed_vaults =  [
        {
            denom: "usdx",
            strategies: ["STRATEGY_TYPE_HARD"],
        },
        {
            denom: "bkava",
            strategies: ["STRATEGY_TYPE_SAVINGS"],
        }]' >$TMP_GENESIS && mv $TMP_GENESIS $GENESIS

    # cat "$GENESIS" | jq '.app_state["staking"]["params"]["bond_denom"]="ukava"' >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
    # cat "$GENESIS" | jq '.app_state["gov"]["params"]["min_deposit"][0]["denom"]="ukava"' >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

    cat "$GENESIS" | jq '.app_state["staking"]["params"]["max_validators"]=200' >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"
    cat "$GENESIS" | jq '.app_state["slashing"]["params"]["signed_blocks_window"]="1000"' >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

    cat "$GENESIS" | jq '.app_state["consensus_params"]["block"]["time_iota_ms"]="3000"' >"$TMP_GENESIS" && mv "$TMP_GENESIS" "$GENESIS"

    # Change app.toml
    APP_TOML="$HOMEDIR"/config/app.toml
    sed -i 's/minimum-gas-prices = "0akava"/minimum-gas-prices = "1000000000akava"/' "$APP_TOML"
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
    NODE_ID=`kava tendermint show-node-id --home $ROOT_DIR/node$i`
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
        VALIDATOR="0gchain_9000_validator_$i"
        set +e
        ret=`kava keys list --keyring-backend os -n | grep $VALIDATOR`
        set -e
        if [[ "$ret" = "" ]]; then
            echo "Create validator key: $VALIDATOR"
            kava keys add $VALIDATOR --keyring-backend os --eth
        fi
    done
elif [[ "$OS_NAME" = "GNU/Linux" ]]; then
    # Create N validators for node0
    for ((i=0; i<$NUM_NODES; i++)) do
        yes $PASSWORD | kava keys add "0gchain_9000_validator_$i" --keyring-backend os --home "$ROOT_DIR"/node0 --eth
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
            yes $PASSWORD | kava add-genesis-account "0gchain_9000_validator_$j" $VLIDATOR_BALANCE --home "$ROOT_DIR/node$i"
        else
            kava add-genesis-account "0gchain_9000_validator_$j" $VLIDATOR_BALANCE --home "$ROOT_DIR/node$i"
        fi
    done
    kava add-genesis-account kava17n8707c20e8gge2tk2gestetjcs4536pdtf8y0 $FAUCET_BALANCE --home "$ROOT_DIR/node$i"
done

# Prepare genesis txs
mkdir -p "$ROOT_DIR"/gentxs
for ((i=0; i<$NUM_NODES; i++)) do
    if [[ "$OS_NAME" = "GNU/Linux" ]]; then
        yes $PASSWORD | kava gentx "0gchain_9000_validator_$i" $STAKING --home "$ROOT_DIR/node$i" --output-document "$ROOT_DIR/gentxs/node$i.json"
    else
        kava gentx "0gchain_9000_validator_$i" $STAKING --home "$ROOT_DIR/node$i" --output-document "$ROOT_DIR/gentxs/node$i.json"
    fi
done

# Create genesis at node0 and copy to other nodes
kava collect-gentxs --home "$ROOT_DIR/node0" --gentx-dir "$ROOT_DIR/gentxs" >/dev/null 2>&1
sed -i '/persistent_peers = /c\persistent_peers = ""' "$ROOT_DIR"/node0/config/config.toml
kava validate-genesis --home "$ROOT_DIR/node0"
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
