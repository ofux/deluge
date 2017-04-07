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

http req responds res
// or
tcp req responds res
// or
mqtt req responds res
// or
grpc req responds re
```
