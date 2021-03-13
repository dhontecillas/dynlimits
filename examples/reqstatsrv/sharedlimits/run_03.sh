#!/bin/bash

CURDIR=$(pwd)
CATALOGDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
cd $CURDIR

# Crates 4 different proxies to the same running backend
export LOCALIP="192.168.1.20"

docker run --rm --name 'dynlimits_3' -d \
    -e DYNLIMITS_FORWARDTO_PORT=9876 \
    -e DYNLIMITS_FORWARDTO_HOST=$LOCALIP \
    -e DYNLIMITS_REDIS_ADDRESS=${LOCALIP}:6379 \
    -v $CURDIR/data:/data \
    -p 7903:7777 \
    dhontecillas/dynlimits:0.1

docker run -i --rm loadimpact/k6 run --vus 1 --duration 1m - <k6_03.js
docker stop dynlimits_3
