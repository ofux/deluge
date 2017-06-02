**Deluge** is a load testing tool for web applications, web APIs, IoT, or any TCP based application.

## Roadmap

- [ ] core of the test runner, able to simulate concurrent users
- [ ] flexible yet simple DSL to write test scenarios
- [ ] recording, using HDRHistograms
- [ ] reporting, on a simple HTML page using JSON to export recorded data
- [ ] REST API (to run tests, get reports, etc.)
    - /scenarios GET-POST-PUT-DELETE
      ```json
      {
          "id": "sc1",
          "name": "My Scenario 1",
          "script": "base64-encoded-script"
      }
      ```
    - /deluges GET-POST-PUT-DELETE
      ```json
      {
          "id": "deluge1",
          "name": "My Deluge",
          "scenarios": {
              "sc1": {
                  "concurrent": 100,
                  "delay": "2s"
              }
          }
      }
      ```
    - /runs GET-POST-DELETE
      ```json
      {
          "id": 876276,
          "deluge_id": "deluge1",
          "report": {
              "...": "..."
          }
      }
      ```
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
