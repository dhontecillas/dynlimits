export LOCALIP=192.168.1.20
docker run --rm -e DYNLIMITS_FORWARDTO_PORT=9876 -e DYNLIMITS_FORWARDTO_HOST=$LOCALIP -e DYNLIMITS_REDIS_ADDRESS=${LOCALIP}:6379 -v $PWD/examples/reqstatsrv:/data -p 7777:7777 dhontecillas/dynlimits:0.1 
