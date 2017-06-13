# <img src="https://cloud.githubusercontent.com/assets/595505/26764853/527aa67c-496f-11e7-8cf6-c494373d4049.png" width="50"/> Deluge

**Deluge** is a load testing tool for web applications, web APIs, IoT, or any TCP based application.

## CLI

```sh
# Starts an orchestrator listening on the given port
$ deluge start orchestrator -port=9090

# Starts a worker listening on the given port as a slave of the given orchestrator
$ deluge start worker -port=8080 -orchestrator=187.32.87.353:9090

# Starts a worker listening on the given port without orchestrator
$ deluge start worker -port=8080

# Runs the deluge on the given worker/orchestrator. Uses REST API behind the scene.
$ deluge run <filename containing deluge and scenario(s)> -on-addr=187.32.87.353:9090

# Silently starts a worker, runs deluge, write report and shutdown worker. Uses REST API behind the scene.
$ deluge run <filename containing deluge and scenario(s)>
```
## REST API

- /deluges GET-POST-DELETE
  ```json
  # POST request
  {
      "name": "My Deluge",
      "scenarios": [
          {
              "name": "My Scenario 1",
              "concurrent": 100,
              "delay": "2s",
              "script": "base64-encoded-script"
          }
      ]
  }

  # POST response
  {
      "id": 876276,
      "name": "My Deluge",
      "status": "running",
      "scenarios": [
          {
              "name": "My Scenario 1",
              "concurrent": 100,
              "delay": "2s",
              "script": "base64-encoded-script"
          }
      ],
      "report": null
  }

  # GET response (of a finished job)
  {
      "id": 876276,
      "name": "My Deluge",
      "status": "done",
      "scenarios": [
          {
              "name": "My Scenario 1",
              "concurrent": 100,
              "delay": "2s",
              "script": "base64-encoded-script"
          }
      ],
      "report": {
          "...": "..."
      }
  }
  ```

## Roadmap

- [ ] core of the test runner, able to simulate concurrent users
- [ ] flexible yet simple DSL to write test scenarios
- [ ] recording, using HDRHistograms
- [ ] reporting, on a simple HTML page using JSON to export recorded data
- [ ] REST API (to run tests, get reports, etc.)
- [ ] CLI
- [ ] clustering or server mode to distribute concurrent users of a scenario across multiple Deluge instances (possibly cross-datacenter)
- [ ] ready to work Docker image


## DSL

The DSL consists of a simple but efficient language inspired from Javascript's syntax with native support of emiting requests on different protocols.

Supported protocols (out of the box) are:
- [ ] HTTP
- [ ] TCP
- [ ] MQTT
- [ ] gRPC

Everything has been made so it is extremely easy to perform requests as show in this example:

```js
let req = {
  "URL": "..."
};

let res = http(req);
// or
let res = tcp(req);
// or
let res = mqtt(req);
// or
let res = grpc(req);

if (res["status"] != 200) {
  fail();
}

assert(res["status"] == 200);

// launch async
async whatever();
async function(){
  // ...
}()
wait

async "group1" whatever();
async "group1" whatever();
wait "group1"
```

## Libs

- [ ] [logrus](https://github.com/sirupsen/logrus) for the logs
- [ ] [cobra](https://github.com/spf13/cobra) for the CLI
- [ ] [grpc-go](https://github.com/grpc/grpc-go) for gRPC
- [ ] [surgemq](https://github.com/influxdata/surgemq) for MQTT
- [ ] [hdrhistogram](https://github.com/codahale/hdrhistogram) for HDR Histogram
