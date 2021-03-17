#!/bin/bash

CURDIR=$(pwd)
CATALOGDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
cd $CURDIR

export LOCALIP="192.168.1.20"

docker run --rm --name 'dynlimits_dyn0' -d \
    -e DYNLIMITS_FORWARDTO_PORT=9876 \
    -e DYNLIMITS_FORWARDTO_HOST=$LOCALIP \
    -e DYNLIMITS_REDIS_ADDRESS=${LOCALIP}:6379 \
    -e DYNLIMITS_CATALOG_SERVER_URL="http://${LOCALIP}:9091" \
    -e DYNLIMITS_CATALOG_SERVER_APIKEY="foo" \
    -e DYNLIMITS_CATALOG_SERVER_POLLSECS=2 \
    -v $CURDIR/data:/data \
    -p 7900:7777 \
    dhontecillas/dynlimits:0.1

docker run -i --name 'k6_0' --rm loadimpact/k6 run --vus 1 --duration 4m - <k6.js
docker stop dynlimits_dyn0
