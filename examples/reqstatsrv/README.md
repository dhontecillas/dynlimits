**Disclaimer**: this is a sloppy test for dynlimit rate limit feature


First download the docker images:

```
docker pull dhontecillas/dynlimits:0.1
docker pull dhontecillas/reqstatsrv:0.1
```

Then run the `docker-compose.yml` file from this directory with:

```
docker-compose up
```

That will run the dummy http server
[**reqstatsrv**](https://github.com/dhontecillas/reqstatsrv)
on port `9876` and a **redis** instance to be used for the test.

Then open the `run.sh` file, and modify your IP. That is script has
the paths to be run from root directory of the project.

Run the file to have `dynlimits` instance running.

There is a request with the api token in `req_foo.sh`.

The limits are set in the `catalog.json` file.

Now you can test the rate limit using the `7777` port (the dynlimits
port), or without rate limit using the `9876` port.


# Catalog server

Inside the `catalog` folder there are the scripts to run a basic
test server using Python's http module. The [run.sh](./catalog/run.sh)
script will launch the http server, so any file in that directory
will be served, including [latest](./catalog/latest) that should
contain a JSON file with a catalog.

Inside the `catalog` folder there is a [gen.py](./catalog/gen.py) script
that allows to generate a long list of endpoints. See the docstring in
the file about how to use it:

```
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
```
