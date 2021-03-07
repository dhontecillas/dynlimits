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
