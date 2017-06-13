package main

import (
	"flag"
	"fmt"
	"github.com/ofux/docilemonkey/docilemonkey"
	"net/http"
)

func main() {
	var addr = *flag.String("listen", ":8080", "server port")

	http.HandleFunc("/", docilemonkey.Handler)

	fmt.Println(`====================================================
Docile Monkey server will respond to any relative URL.

Query params:
	s: will be the HTTP status code of the response (ex: 200, or 500)
	t: time to wait before responding (ex: 10s, or 200ms)
	b: will be the body of the response
	bb: if parameter 'b' has no value and if bb=1 then the body of the request (if any) will be sent back in the response

Example of request:
	GET http://localhost:8080/foo/bar?s=201&t=500ms&b={"hello":"world"}
`)
	fmt.Printf("Listening on %s...\n", addr)
	fmt.Println(`====================================================`)
	http.ListenAndServe(addr, nil)
}
