package api

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"net/http"
)

const (
	HeaderContentTypeKey      = "Content-Type"
	HeaderContentTypeJsonUTF8 = "application/json; charset=UTF-8"
)

// SendJSONWithHTTPCode outputs JSON with an HTTP code
func SendJSONWithHTTPCode(w http.ResponseWriter, d interface{}, code int) {
	w.Header().Set(HeaderContentTypeKey, HeaderContentTypeJsonUTF8)
	w.WriteHeader(code)
	if d != nil {
		err := json.NewEncoder(w).Encode(d)
		if err != nil {
			log.WithField("body", d).WithField("code", code).Error("error while encoding JSON body of response")
			// panic will cause the http.StatusInternalServerError to be send to users thanks to negroni recovery
			panic(err)
		}
	}
}

// SendJSONOk outputs a JSON with http.StatusOK code
func SendJSONOk(w http.ResponseWriter, d interface{}) {
	SendJSONWithHTTPCode(w, d, http.StatusOK)
}

// SendJSONError sends error with a custom message and error code
func SendJSONError(w http.ResponseWriter, error string, code int) {
	SendJSONWithHTTPCode(w, Error{error}, code)
}
