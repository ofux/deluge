![deluge_title](https://user-images.githubusercontent.com/595505/27251395-53b0eb7e-5346-11e7-8b4f-efe8308c3eae.png)

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

### /v1/jobs GET-POST-DELETE

POST request body example *(text/plain)*:

```js
// The request body contains the script (written in deluge-DSL) to run
deluge("Some name", "10s", {
    "sc1": {
        "concurrent": 100,
        "delay": "2s"
    }
});

scenario("sc1", "Some scenario", function () {

    http("Some request", {
        "url": "http://localhost:8080/hello/toto"
    });

});
```

POST/GET response body example of an **unfinished** job *(application/json)*:

```json
{
    "ID": "0b0781d9-0cae-47e2-8196-a5cc6e24e086",
    "Name": "Some name",
    "Status": "InProgress",
    "GlobalDuration": 30000000000,
    "Scenarios": {
        "sc1": {
            "Name": "Some scenario",
            "IterationDuration": 0,
            "Status": "InProgress",
            "Errors": [],
            "Report": null
        }
    }
}
```

POST/GET response body example of a **finished** job *(application/json)*:

```json
{
    "ID": "0b0781d9-0cae-47e2-8196-a5cc6e24e086",
    "Name": "Some name",
    "Status": "DoneSuccess",
    "GlobalDuration": 30000000000,
    "Scenarios": {
        "sc1": {
            "Name": "Some scenario",
            "IterationDuration": 0,
            "Status": "DoneSuccess",
            "Errors": [],
            "Report": { "...": "..." }
        }
    }
}
```


## DSL

The DSL consists of a simple, extremely-easy-to-learn language with native support for emiting requests with different protocols.

Supported protocols to make some requests (out of the box) are:
- [x] HTTP
- [ ] TCP
- [ ] MQTT
- [ ] gRPC

Everything has been made so it is extremely easy to perform requests as show in this example:

```js
deluge("Some name", "5s", {
    "sc1": {
        "concurrent": 100,
        "delay": "2s"
    }
});

scenario("sc1", "Some scenario", function () {

    http("Some request", {
        "url": "http://localhost:8080/hello/foo"
    });

});
```

## Upcoming

- [ ] nice HTML report
- [ ] ability to distribute concurrent users of a scenario across multiple Deluge instances (possibly cross-datacenter)
- [ ] Docker image

## Some libs we use

- [x] [logrus](https://github.com/sirupsen/logrus) for the logs
- [x] [cobra](https://github.com/spf13/cobra) for the CLI
- [ ] [grpc-go](https://github.com/grpc/grpc-go) for gRPC
- [ ] [surgemq](https://github.com/influxdata/surgemq) for MQTT
- [x] [hdrhistogram](https://github.com/codahale/hdrhistogram) for HDR Histogram
