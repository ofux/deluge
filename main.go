package main

import (
	"flag"
	"github.com/ofux/deluge/api"
)

func main() {
	var port = *flag.Int("listen", 33033, "server port")
	api.Serve(port)
}
