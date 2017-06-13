package docilemonkey

import (
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

// Handler that will respond what you ask it to respond.
// Possible query params are :
// 		s: will be the HTTP status code of the response (ex: 200, or 500)
// 		t: time to wait before responding (ex: 10s, or 200ms)
// 		b: will be the body of the response
// 		bb: if parameter 'b' has no value and if bb=1 then the body of the request (if any) will be sent back in the response
func Handler(w http.ResponseWriter, r *http.Request) {
	if paramTime := r.URL.Query().Get("t"); paramTime != "" {
		if t, err := time.ParseDuration(paramTime); err == nil {
			time.Sleep(t)
		}
	}

	if paramStatus := r.URL.Query().Get("s"); paramStatus != "" {
		if status, err := strconv.ParseInt(paramStatus, 10, 0); err == nil {
			w.WriteHeader(int(status))
		}
	}

	if paramBody := r.URL.Query().Get("b"); paramBody != "" {
		w.Write([]byte(paramBody))
	} else if paramBodyBack := r.URL.Query().Get("bb"); paramBodyBack == "1" {
		if body, err := ioutil.ReadAll(r.Body); err == nil {
			w.Write(body)
		}
	}
}
