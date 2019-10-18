package api

import (
	"encoding/json"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"reflect"
)

const (
	HeaderContentTypeKey      = "Content-Type"
	HeaderContentTypeJsonUTF8 = "application/json; charset=UTF-8"
)

type List struct {
	Elements interface{} `json:"elements"`
}

func ListOf(elements interface{}) List {
	switch reflect.TypeOf(elements).Kind() {
	case reflect.Slice:
		return List{Elements: elements}
	}
	panic(errors.Errorf("ListOf is expecting a slice but got %+v", elements))
}

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

// SendRawStringHTTPCode outputs string as-is with an HTTP code
func SendRawStringHTTPCode(w http.ResponseWriter, str string, code int) {
	w.Header().Set(HeaderContentTypeKey, HeaderContentTypeJsonUTF8)
	w.WriteHeader(code)
	_, err := w.Write([]byte(str))
	if err != nil {
		log.WithField("body", str).WithField("code", code).Error("error while writing body of response")
		// panic will cause the http.StatusInternalServerError to be send to users thanks to negroni recovery
		panic(err)
	}
}

func GetNonEmptyBody(w http.ResponseWriter, r *http.Request) ([]byte, bool) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		SendJSONError(w, err.Error(), http.StatusBadRequest)
		return nil, false
	}
	if len(body) == 0 {
		SendJSONError(w, "Missing body", http.StatusBadRequest)
		return nil, false
	}
	return body, true
}

func GetJSONBody(w http.ResponseWriter, r *http.Request, out interface{}) bool {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		SendJSONError(w, err.Error(), http.StatusBadRequest)
		return false
	}
	if len(body) == 0 {
		SendJSONError(w, "Missing body", http.StatusBadRequest)
		return false
	}
	err = json.Unmarshal(body, out)
	if err != nil {
		SendJSONError(w, err.Error(), http.StatusBadRequest)
		return false
	}
	return true
}
