"""
Usage:
    python gen.py [num_apikeys] [num_endpoints] [rate_limits]

    * num_apikeys: the number of api keys to generate
    * num_endpoints: the number of endpoints to generate
    * rate_limits: a comma separated list of values for the
        base api calls rate limits

Example:

    python gen.py 5 2 10,100,500

In order to give some variability of rate limit in different endpoints
inside the api key, the odd endpoints will have its corresponding
rate limit applied, and the even one twice the rate limit. 

The "base" rate limit for each api key, will cycle the list of rate limits

From the previous example:
    * 000000: 10 (for odd endpoint), 20 (for even endpoint)
    * 000001: 100 and 200
    * 000002: 500 and 1000
    * 000003: 10 and 20
"""
import json
import sys


ratelimits = [5, 20, 50, 100, 500, 1000, 5000]
num_endpoints = 10
num_apikeys = 2

if len(sys.argv) > 3:
    ratelimits = [int(x) for x in sys.argv[3].split(',')]

if len(sys.argv) > 2:
    num_endpoints = int(sys.argv[2])

if len(sys.argv) > 1:
    num_apikeys = int(sys.argv[1])

base = {
    "methods": ["GET", "POST", "DELETE", "PUT", "PATCH", "HEAD", "OPTIONS"],
	"paths": [],
	"endpoints": [],
	"apilimits": [],
}

for eidx in range(num_endpoints):
    base['paths'].append(f"/endpoint_{eidx}")
    base['endpoints'].append({"p": eidx, "m": 0})

num_rate_limits = len(ratelimits)
for akidx in range(num_apikeys):
    rl = ratelimits[akidx % num_rate_limits]
    limits = []
    for eidx in range(num_endpoints):
        limits.append({"ep": eidx, "rl": rl})
    base['apilimits'].append({"key": f"{akidx:06}", "limits": limits})

print(json.dumps(base, sort_keys=True, separators=(',', ':')))
# print(json.dumps(base, sort_keys=True, indent=4))
