package core

import (
	"github.com/ofux/deluge/core/recording"
	"github.com/ofux/deluge/dsl/ast"
	"github.com/ofux/deluge/dsl/evaluator"
	"github.com/ofux/deluge/dsl/object"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// execHTTPRequest is the implementation of the built-in function 'http' that performs HTTP requests and
// records timing data.
func (su *simUser) execHTTPRequest(node ast.Node, args ...object.Object) object.Object {
	if oErr := evaluator.AssertArgsType(node, args, object.STRING_OBJ, object.HASH_OBJ); oErr != nil {
		return oErr
	}
	reqName := args[0].(*object.String).Value
	reqObj := args[1].(*object.Hash)

	req, errObj := createRequest(node, reqObj)
	if evaluator.IsError(errObj) {
		return errObj
	}

	su.log.Debugf("Performing HTTP request: %s %s", req.Method, req.URL.String())
	start := time.Now()
	res, err := su.client.Do(req)
	end := time.Now()

	if err != nil {
		su.log.Debugf("Request error: %s", err.Error())
		return evaluator.NewError(node, err.Error())
	}
	defer res.Body.Close()

	duration := end.Sub(start)
	su.log.Debugf("Response status: %s in %s", "res.Status", duration.String())
	su.httpRecorder.Record(&recording.HTTPRecordEntry{
		Iteration:  su.iteration,
		Name:       reqName,
		Value:      duration.Nanoseconds() / 100000,
		StatusCode: res.StatusCode,
	})

	return getResponseObject(node, res)
}

func createRequest(node ast.Node, reqObj *object.Hash) (*http.Request, *object.Error) {
	// Get URL
	url, _, err := reqObj.GetAsString("url")
	if err != nil {
		return nil, evaluator.NewError(node, "invalid HTTP request: %s", err.Error())
	}

	// Get Method
	var method = "GET"
	if m, ok, err := reqObj.GetAsString("method"); ok {
		if err != nil {
			return nil, evaluator.NewError(node, "invalid HTTP request: %s", err.Error())
		}
		method = m.Value
	}

	// Get Body
	var body io.Reader
	if b, ok, err := reqObj.GetAsString("body"); ok {
		if err != nil {
			return nil, evaluator.NewError(node, "invalid HTTP request: %s", err.Error())
		}
		body = strings.NewReader(b.Value)
	}

	// Create request
	req, err := http.NewRequest(method, url.Value, body)
	if err != nil {
		return nil, evaluator.NewError(node, err.Error())
	}

	// Add Headers
	if headers, ok, err := reqObj.GetAsHash("headers"); ok {
		if err != nil {
			return nil, evaluator.NewError(node, "invalid HTTP request: %s", err.Error())
		}
		for headerKey, headerVal := range headers.Pairs {
			headerValStr, ok := headerVal.(*object.String)
			if !ok {
				return nil, evaluator.NewError(node, "invalid HTTP header '%s': should be of type %s but was %s", headerKey, object.STRING_OBJ, headerVal.Type())
			}
			req.Header.Add(string(headerKey), headerValStr.Value)
		}
	}

	return req, nil
}

func getResponseObject(node ast.Node, res *http.Response) object.Object {
	resHeaders := getResponseHeaders(res)
	if evaluator.IsError(resHeaders) {
		return resHeaders
	}

	resBody := getResponseBody(node, res)
	if evaluator.IsError(resBody) {
		return resBody
	}

	return &object.Hash{
		Pairs: map[object.HashKey]object.Object{
			object.HashKey("status"):  &object.Integer{Value: int64(res.StatusCode)},
			object.HashKey("headers"): resHeaders,
			object.HashKey("body"):    resBody,
		},
		IsImmutable: true,
	}
}

func getResponseHeaders(res *http.Response) object.Object {
	resHeaders := make(map[object.HashKey]object.Object)
	for k := range res.Header {
		resHeaders[object.HashKey(k)] = &object.String{Value: res.Header.Get(k)}
	}
	return &object.Hash{
		Pairs:       resHeaders,
		IsImmutable: true,
	}
}

func getResponseBody(node ast.Node, res *http.Response) object.Object {
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return evaluator.NewError(node, err.Error())
	}
	return &object.String{Value: string(resBody)}
}
