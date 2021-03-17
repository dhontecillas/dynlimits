#!/bin/bash

CURDIR=$(pwd)
CATALOGDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
cd $CURDIR
python3 -m http.server --directory=$CATALOGDIR 9091
