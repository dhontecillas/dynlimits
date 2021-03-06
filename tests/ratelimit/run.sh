#!/bin/bash

# https://k6.io/docs/getting-started/running-k6
# https://github.com/loadimpact/k6#running-k6
docker run -i loadimpact/k6 run --vus 500 --duration 40s - <basic.js
# docker run -i loadimpact/k6 run --vus 200 --duration 10s - <basic.js
