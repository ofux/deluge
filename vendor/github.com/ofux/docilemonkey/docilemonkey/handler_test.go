package docilemonkey

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()

	Handler(w, req)

	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected HTTP status code to be %d, got %d", http.StatusOK, resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err.Error())
	}

	if len(body) != 0 {
		t.Errorf("Expected body to be empty, got %s", string(body))
	}
}

func TestHandler_withTime(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com/foo?t=100ms", nil)
	w := httptest.NewRecorder()

	start := time.Now()
	Handler(w, req)
	elaspedTime := time.Now().Sub(start)

	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected HTTP status code to be %d, got %d", http.StatusOK, resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err.Error())
	}

	if len(body) != 0 {
		t.Errorf("Expected body to be empty, got %s", string(body))
	}

	if elaspedTime.Nanoseconds() < 100000000 {
		t.Fatalf("Responded in less than 100ms. Should have take more time")
	}
}

func TestHandler_withStatus(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com/foo?s=500", nil)
	w := httptest.NewRecorder()
	Handler(w, req)

	resp := w.Result()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected HTTP status code to be %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err.Error())
	}

	if len(body) != 0 {
		t.Errorf("Expected body to be empty, got %s", string(body))
	}
}

func TestHandler_withBody(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com/foo?b=test", nil)
	w := httptest.NewRecorder()
	Handler(w, req)

	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected HTTP status code to be %d, got %d", http.StatusOK, resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err.Error())
	}

	if string(body) != "test" {
		t.Errorf("Expected body to be '%s', got '%s'", "test", string(body))
	}
}

func TestHandler_withBodyBack(t *testing.T) {
	req := httptest.NewRequest("POST", "http://example.com/foo?bb=1", strings.NewReader("test"))
	w := httptest.NewRecorder()
	Handler(w, req)

	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected HTTP status code to be %d, got %d", http.StatusOK, resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err.Error())
	}

	if string(body) != "test" {
		t.Errorf("Expected body to be '%s', got '%s'", "test", string(body))
	}
}
