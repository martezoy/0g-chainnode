#!/usr/bin/env bash

for D in ../x/*; do
  if [ -d "${D}" ]; then
    rm -rf "./$(echo $D | awk -F/ '{print $NF}')"
  fi
done