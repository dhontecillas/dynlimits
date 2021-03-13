#!/bin/bash

CURDIR=$(pwd)
CATALOGDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
cd $CURDIR

# Crates 4 different proxies to the same running backend
export LOCALIP="192.168.1.20"

docker run -i --rm loadimpact/k6 run --vus 2000 --duration 1m - < baseline.js
