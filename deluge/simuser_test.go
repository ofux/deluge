package deluge

import (
	"fmt"
	"github.com/ofux/deluge-dsl/lexer"
	"github.com/ofux/deluge-dsl/parser"
	"net/http"
	"net/http/httptest"
	"testing"
)

func NewSimUserTest(t *testing.T, js string) (*SimUser, Recorder) {
	l := lexer.New(js)
	p := parser.New(l)

	program, ok := p.ParseProgram()
	if !ok {
		PrintParserErrors(p.Errors())
		t.Fatal("Parsing error(s)")
	}

	recorder := NewRecorder(1)
	return NewSimUser("1", program, recorder), recorder
}

func checkSimUserStatus(t *testing.T, su *SimUser, status SimUserStatus) {
	if su.Status != status {
		t.Fatalf("Bad SimUser status %d, expected %d", su.Status, status)
	}
}

func TestSimUser_Assert(t *testing.T) {
	t.Run("Assert true", func(t *testing.T) {
		su, _ := NewSimUserTest(t, `
		assert(1+1 == 2)
		`)
		su.Run(0)
		checkSimUserStatus(t, su, DoneSuccess)
	})

	t.Run("Assert false", func(t *testing.T) {
		su, _ := NewSimUserTest(t, `
		assert(1+1 == 3)
		`)
		su.Run(0)
		checkSimUserStatus(t, su, DoneError)
	})
}

func TestSimUser_ExecHTTPRequest(t *testing.T) {
	t.Run("Simple HTTP GET request", func(t *testing.T) {
		callCount := 0
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if r.Method != "GET" {
				t.Errorf("Expected HTTP method to be %s, got %s", "GET", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, `{"foo":"bar"}`)
		}))
		defer ts.Close()

		url := ts.URL
		const reqName = "Some request"
		const recName = reqName + "->200"

		su, rec := NewSimUserTest(t, `
		http("`+reqName+`", {
			"url": "`+url+`"
		});
		`)
		su.Run(0)
		checkSimUserStatus(t, su, DoneSuccess)
		checkRecords(t, rec, recName, 1)

		if callCount != 1 {
			t.Errorf("Expected %d call(s), got %d", 1, callCount)
		}
	})

	t.Run("Simple HTTP POST request", func(t *testing.T) {
		callCount := 0
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if r.Method != "POST" {
				t.Errorf("Expected HTTP method to be %s, got %s", "POST", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, `{"foo":"bar"}`)
		}))
		defer ts.Close()

		url := ts.URL
		const reqName = "Some request"
		const recName = reqName + "->200"

		su, rec := NewSimUserTest(t, `
				http("`+reqName+`", {
					"url": "`+url+`",
					"method": "POST"
				});
				`)
		su.Run(0)
		checkSimUserStatus(t, su, DoneSuccess)
		checkRecords(t, rec, recName, 1)

		if callCount != 1 {
			t.Errorf("Expected %d call(s), got %d", 1, callCount)
		}
	})

	t.Run("Bad HTTP arguments", func(t *testing.T) {
		su, _ := NewSimUserTest(t, `
				http("foo");
				`)
		su.Run(0)
		checkSimUserStatus(t, su, DoneError)
	})

	t.Run("Bad HTTP name", func(t *testing.T) {
		su, _ := NewSimUserTest(t, `
				http(1, {
					"url": "http://plop.org",
					"method": "POST"
				});
				`)
		su.Run(0)
		checkSimUserStatus(t, su, DoneError)
	})

	t.Run("No HTTP url", func(t *testing.T) {
		su, _ := NewSimUserTest(t, `
				http("foo", {
					"method": "POST"
				});
				`)
		su.Run(0)
		checkSimUserStatus(t, su, DoneError)
	})

	t.Run("Bad HTTP url", func(t *testing.T) {
		su, _ := NewSimUserTest(t, `
				http("foo", {
					"url": 42
				});
				`)
		su.Run(0)
		checkSimUserStatus(t, su, DoneError)
	})

	t.Run("Bad HTTP url 2", func(t *testing.T) {
		su, _ := NewSimUserTest(t, `
				http("foo", {
					"url": "foobar"
				});
				`)
		su.Run(0)
		checkSimUserStatus(t, su, DoneError)
	})
}

func checkRecords(t *testing.T, rec Recorder, recName string, recCount int64) {
	rec.Close()
	records, err := rec.GetRecords()
	if err != nil {
		t.Fatal(err.Error())
	}
	record, ok := records[recName]
	if !ok {
		t.Fatalf("Expected to have some records for '%s'", recName)
	}
	if len(record) != 1 {
		t.Fatalf("Expected to have %d records for '%s', got %d", 1, recName, len(record))
	}
	if record[0].TotalCount() != recCount {
		t.Errorf("Expected to have totalCount = %d, got %d", recCount, record[0].TotalCount())
	}
}
