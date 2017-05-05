**Deluge** is an heavy load testing tool for web applications.

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
