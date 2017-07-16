package core

import (
	"github.com/ofux/deluge/core/recording"
	"github.com/ofux/deluge/core/recording/recordingtest"
	"github.com/ofux/deluge/dsl/lexer"
	"github.com/ofux/deluge/dsl/parser"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func NewSimUserTest(t *testing.T, js string) *simUser {
	l := lexer.New(js)
	p := parser.New(l)

	script, ok := p.ParseProgram()
	if !ok {
		PrintParserErrors(p.Errors())
		t.Fatal("Parsing error(s)")
	}

	logger := log.New()
	// discard DSL logs for testing
	logger.Out = ioutil.Discard

	sc := &Scenario{
		Name:         "Test scenario",
		script:       script,
		httpRecorder: recording.NewHTTPRecorder(),
		log: logger.WithFields(log.Fields{
			"scenario": "Test scenario",
		}),
	}

	return newSimUser("1", sc)
}

func checkSimUserStatus(t *testing.T, su *simUser, status simUserStatus) {
	if su.status != status {
		t.Fatalf("Bad simUser status %d, expected %d", su.status, status)
	}
}

func checkSimUserError(t *testing.T, su *simUser, expectedError string) {
	require.NotNil(t, su.execError)
	assert.Equal(t, su.execError.Message, expectedError)
}

func TestSimUser_Assert(t *testing.T) {
	t.Run("Assert true", func(t *testing.T) {
		su := NewSimUserTest(t, `
		assert(1+1 == 2)
		`)
		su.run(0)
		checkSimUserStatus(t, su, UserDoneSuccess)
	})

	t.Run("Assert false", func(t *testing.T) {
		su := NewSimUserTest(t, `
		assert(1+1 == 3)
		`)
		su.run(0)
		checkSimUserStatus(t, su, UserDoneError)
	})
}

func TestSimUser_ExecHTTPRequest(t *testing.T) {
	t.Run("HTTP GET request", func(t *testing.T) {
		callCount := 0
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			assert.Equal(t, "GET", r.Method)
			body, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)
			assert.Len(t, body, 0)
		}))
		defer ts.Close()

		url := ts.URL
		const reqName = "Some request"

		su := NewSimUserTest(t, `
		http("`+reqName+`", {
			"url": "`+url+`"
		});
		`)
		su.run(0)
		checkSimUserStatus(t, su, UserDoneSuccess)
		checkRecords(t, su.httpRecorder, reqName, 1)
		assert.Equal(t, 1, callCount)
	})

	t.Run("HTTP DELETE request", func(t *testing.T) {
		callCount := 0
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			assert.Equal(t, "DELETE", r.Method)
			body, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)
			assert.Len(t, body, 0)
		}))
		defer ts.Close()

		url := ts.URL
		const reqName = "Some request"

		su := NewSimUserTest(t, `
				http("`+reqName+`", {
					"url": "`+url+`",
					"method": "DELETE"
				});
				`)
		su.run(0)
		checkSimUserStatus(t, su, UserDoneSuccess)
		checkRecords(t, su.httpRecorder, reqName, 1)
		assert.Equal(t, 1, callCount)
	})

	t.Run("HTTP POST request with a body", func(t *testing.T) {
		callCount := 0
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			assert.Equal(t, "POST", r.Method)
			body, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)
			assert.Equal(t, "test", string(body))
		}))
		defer ts.Close()

		url := ts.URL
		const reqName = "Some request"

		su := NewSimUserTest(t, `
				http("`+reqName+`", {
					"url": "`+url+`",
					"method": "POST",
					"body": "test"
				});
				`)
		su.run(0)
		checkSimUserStatus(t, su, UserDoneSuccess)
		checkRecords(t, su.httpRecorder, reqName, 1)
		assert.Equal(t, 1, callCount)
	})

	t.Run("HTTP GET request with headers", func(t *testing.T) {
		callCount := 0
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			assert.Equal(t, "GET", r.Method)
			body, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)
			assert.Len(t, body, 0)
			assert.Equal(t, "application/json", r.Header.Get("Accept"))
			assert.Equal(t, "Bearer foobar-bar-foo", r.Header.Get("Authorization"))
			assert.Equal(t, "", r.Header.Get("Foo"))
		}))
		defer ts.Close()

		url := ts.URL
		const reqName = "Some request"

		su := NewSimUserTest(t, `
		http("`+reqName+`", {
			"url": "`+url+`",
			"headers": {
				"Accept": "application/json",
				"Authorization": "Bearer foobar-bar-foo"
			}
		});
		`)
		su.run(0)
		checkSimUserStatus(t, su, UserDoneSuccess)
		checkRecords(t, su.httpRecorder, reqName, 1)
		assert.Equal(t, 1, callCount)
	})

	t.Run("Bad HTTP arguments", func(t *testing.T) {
		su := NewSimUserTest(t, `
				http("foo");
				`)
		su.run(0)
		checkSimUserStatus(t, su, UserDoneError)
		checkSimUserError(t, su, "wrong number of arguments. got=1, want=2")
	})

	t.Run("Bad HTTP name", func(t *testing.T) {
		su := NewSimUserTest(t, `
				http(1, {
					"url": "http://plop.org",
					"method": "POST"
				});
				`)
		su.run(0)
		checkSimUserStatus(t, su, UserDoneError)
		checkSimUserError(t, su, "wrong type of argument n°1. got=INTEGER, want=STRING")
	})

	t.Run("No HTTP url", func(t *testing.T) {
		su := NewSimUserTest(t, `
				http("foo", {
					"method": "POST"
				});
				`)
		su.run(0)
		checkSimUserStatus(t, su, UserDoneError)
		checkSimUserError(t, su, "invalid HTTP request: missing 'url' field")
	})

	t.Run("Bad HTTP url", func(t *testing.T) {
		su := NewSimUserTest(t, `
				http("foo", {
					"url": 42
				});
				`)
		su.run(0)
		checkSimUserStatus(t, su, UserDoneError)
		checkSimUserError(t, su, "invalid HTTP request: 'url' should be of type STRING but was INTEGER")
	})

	t.Run("Bad HTTP url 2", func(t *testing.T) {
		su := NewSimUserTest(t, `
				http("foo", {
					"url": "foobar"
				});
				`)
		su.run(0)
		checkSimUserStatus(t, su, UserDoneError)
		checkSimUserError(t, su, "Get foobar: unsupported protocol scheme \"\"")
	})

	t.Run("Bad HTTP method", func(t *testing.T) {
		su := NewSimUserTest(t, `
				http("foo", {
					"url": "http://plop.org",
					"method": "BAD BAD BAD"
				});
				`)
		su.run(0)
		checkSimUserStatus(t, su, UserDoneError)
		checkSimUserError(t, su, "net/http: invalid method \"BAD BAD BAD\"")
	})

	t.Run("Bad HTTP method type", func(t *testing.T) {
		su := NewSimUserTest(t, `
				http("foo", {
					"url": "http://plop.org",
					"method": 0
				});
				`)
		su.run(0)
		checkSimUserStatus(t, su, UserDoneError)
		checkSimUserError(t, su, "invalid HTTP request: 'method' should be of type STRING but was INTEGER")
	})

	t.Run("Bad HTTP body type", func(t *testing.T) {
		su := NewSimUserTest(t, `
				http("foo", {
					"url": "http://plop.org",
					"body": 1
				});
				`)
		su.run(0)
		checkSimUserStatus(t, su, UserDoneError)
		checkSimUserError(t, su, "invalid HTTP request: 'body' should be of type STRING but was INTEGER")
	})

	t.Run("Bad HTTP headers type", func(t *testing.T) {
		su := NewSimUserTest(t, `
				http("foo", {
					"url": "http://plop.org",
					"headers": 1
				});
				`)
		su.run(0)
		checkSimUserStatus(t, su, UserDoneError)
		checkSimUserError(t, su, "invalid HTTP request: 'headers' should be of type HASH but was INTEGER")
	})

	t.Run("Bad HTTP headers", func(t *testing.T) {
		su := NewSimUserTest(t, `
				http("foo", {
					"url": "http://plop.org",
					"headers": {
						"foo": 1
					}
				});
				`)
		su.run(0)
		checkSimUserStatus(t, su, UserDoneError)
		checkSimUserError(t, su, "invalid HTTP header 'foo': should be of type STRING but was INTEGER")
	})
}

func checkRecords(t *testing.T, rec *recording.HTTPRecorder, recName string, recCount int64) {
	rec.Close()
	records, err := rec.GetRecords()
	if err != nil {
		t.Fatal(err.Error())
	}
	record := records.PerIteration[0]
	recordingtest.CheckHTTPRecord(t, record, recName, recCount, 200, recording.Ok)
}
