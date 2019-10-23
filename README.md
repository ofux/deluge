![deluge_title](https://user-images.githubusercontent.com/595505/27251395-53b0eb7e-5346-11e7-8b4f-efe8308c3eae.png)

[![Build Status](https://travis-ci.org/ofux/deluge.svg?branch=master)](https://travis-ci.org/ofux/deluge)
[![Go Report Card](https://goreportcard.com/badge/github.com/ofux/deluge)](https://goreportcard.com/report/github.com/ofux/deluge)
[![codecov](https://codecov.io/gh/ofux/deluge/branch/master/graph/badge.svg)](https://codecov.io/gh/ofux/deluge)

**Deluge** is a load testing tool for web applications, web APIs, IoT, or any TCP based application.

## Features

- It’s fast
- It's stupid simple
- No dependency hell, single binary made with go
- Rest API
- Pain-less, brain-less DSL
- Scales vertically and horizontally *(in progress)*
- Nice reporting *(in progress)*

## CLI

### Done

```sh
# Starts a worker listening on the given port without orchestrator
$ deluge start worker --port=8080

# Runs the deluge on the given worker/orchestrator. Uses REST API behind the scene.
$ deluge run <filename containing deluge's scenario(s)> <output filename> --remote=http://mydeluge.net:33033

# Silently starts a worker, runs deluge, write report and shutdown worker. Uses REST API behind the scene.
$ deluge run <filename containing deluge's scenario(s)> <output filename>
```

### In progress

```sh
# Starts an orchestrator listening on the given port
$ deluge start orchestrator --port=9090

# Starts a worker listening on the given port as a slave of the given orchestrator
$ deluge start worker --port=8080 --orchestrator=187.32.87.353:9090
```

## REST API

[Swagger documentation](https://app.swaggerhub.com/apis-docs/ofu/deluge-api/0.0.1)


## DSL

The DSL consists of a simple, extremely-easy-to-learn language with native support for emiting requests with different protocols.

### Scenarios

Scenarios are scripts executed at each iteration by each virtual user during the deluge. They are written in DelugeDSL.

Example of scenarios:

Simple scenario performing a single HTTP GET request.
```js
scenario("some-id", "Some scenario", function () {
    http("Some request", {
        "url": "http://localhost:8080/hello/foo"
    });
});
```

More complex scenario that uses session to store data between iterations.
```js
scenario("product", "Test the product entity", function (args, session) {

    let authenticate = function () {
        let resAuth = http("Authentication", {
            "url": "http://localhost:8080/oauth/token?log=1&b={\"access_token\":\"foooooooobar\"}",
            "headers": {
                "Content-Type": "application/x-www-form-urlencoded"
            },
            "body": urlParamsEncode({
                "username": "admin",
                "password": "admin",
                "grant_type": "password"
                // etc
            })
        });

        assert(resAuth["status"] == 200);
        let access_token = parseJson(resAuth["body"])["access_token"];
        assert(len(access_token) > 0);

        return access_token;
    };



    let baseUrl = "http://127.0.0.1:8080";
    if (args["baseUrl"]) {
        baseUrl = args["baseUrl"];
    }

    let httpCommonHeaders = {
        "Accept": "application/json",
        "Accept-Encoding": "gzip, deflate",
        "Accept-Language": "fr,fr-fr;q=0.8,en-us",
        "Connection": "keep-alive",
        "User-Agent": "Mozilla/5.0 etc."
    };


    let res1 = http("First unauthenticated request", {
        "url": baseUrl + "/api/v1/account?s=401",
        "headers": httpCommonHeaders
    });

    assert(res1["status"] == 401);

    pause("100ms");

    let access_token = session["access_token"];
    if (access_token == null) {
        access_token = authenticate();
        session["access_token"] = access_token;
    }

    pause("100ms");

    let res2 = http("Authenticated request", {
        "url": baseUrl + "/api/v1/account",
        "headers": merge(httpCommonHeaders, {
            "Authorization": "Bearer " + access_token
        })
    });

    assert(res2["status"] == 200);

    pause("100ms");

    for (let i=0; i < 2; i++) {
        let res3 = http("Get all products", {
            "url": baseUrl + "/api/v1/products",
            "headers": merge(httpCommonHeaders, {
                "Authorization": "Bearer " + access_token
            })
        });

        assert(res3["status"] == 200);

        let res4 = http("Create new product", {
            "url": baseUrl + "/api/v1/products?s=201",
            "method": "POST",
            "headers": merge(httpCommonHeaders, {
                "Authorization": "Bearer " + access_token
            }),
            "body": toJson({
                "ref": "SJ5"
                // etc
            })
        });

        assert(res4["status"] == 201);
    }

});
```

Supported protocols to make some requests (out of the box) are:
- [x] HTTP
- [ ] TCP
- [ ] MQTT
- [ ] gRPC

### Deluges

Deluges are scripts written in DelugeDSL that describe which scenrio(s) should be run and their respective configuration.

```js
deluge("some-deluge-id", "Some name", "5s", {
    "some-scenario-id": {
        "concurrent": 100,
        "delay": "2s",
        "args": {
            "baseUrl": "http://127.0.0.1:8080"
        }
    }
});
```

## TODO

- [ ] nice HTML report
- [ ] ability to distribute concurrent users of a scenario across multiple Deluge instances (ditributed workers)
- [ ] Docker image
