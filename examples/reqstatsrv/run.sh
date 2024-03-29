# Crates 4 different procies to the same running backend
#
#
export LOCALIP=192.168.1.20

CURDIR=$(pwd)
CATALOGDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
cd $CURDIR

docker run --rm --name 'dynlimits_0' \
    -e DYNLIMITS_FORWARDTO_PORT=9876 \
    -e DYNLIMITS_FORWARDTO_HOST=$LOCALIP \
    -e DYNLIMITS_REDIS_ADDRESS=${LOCALIP}:6379 \
    -v $CATALOGDIR:/data \
    -p 7900:7777 \
    dhontecillas/dynlimits:0.1
