#!/bin/bash

CURDIR=$(pwd)
CATALOGDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
cd $CURDIR

# python ../catalog/gen.py 9999 100 > ./data/catalog.json

# sleep 1

# Crates 4 different proxies to the same running backend
export LOCALIP="192.168.1.20"

# docker run --rm --name 'dynlimits_0' -d \
#     -e DYNLIMITS_FORWARDTO_PORT=9876 \
#     -e DYNLIMITS_FORWARDTO_HOST=$LOCALIP \
#     -e DYNLIMITS_REDIS_ADDRESS=${LOCALIP}:6379 \
#     -v $CURDIR/data:/data \
#     -p 7900:7777 \
#     dhontecillas/dynlimits:0.1



export DYNLIMITS_FORWARDTO_PORT=9876
export DYNLIMITS_FORWARDTO_HOST=$LOCALIP
export DYNLIMITS_REDIS_ADDRESS=${LOCALIP}:6379
export DYNLIMITS_LISTEN_PORT=7900

# ../../../../proxy 1> stdout 2> stderr &
# ../../../../proxy

# Give some time to the server to load the catalog
sleep 30


# https://k6.io/docs/getting-started/running-k6
# https://github.com/loadimpact/k6#running-k6
# docker run -i --rm loadimpact/k6 run --vus 50 --duration 1m - < basic.js
docker run -i --rm loadimpact/k6 run --vus 50 --duration 1m - < ../basic.js

# docker stop dynlimits_0
