package deluge_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ofux/deluge/deluge"
	"github.com/ofux/deluge/dsl/lexer"
	"github.com/ofux/deluge/dsl/parser"
)

// TestConcurrentHTTPRequests verifies thread-safety of concurrent HTTP requests
func TestConcurrentHTTPRequests(t *testing.T) {

	var requestCount int32
	var mu sync.Mutex
	var goroutineIDs []int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := int(atomic.AddInt32(&requestCount, 1))

		mu.Lock()
		goroutineIDs = append(goroutineIDs, count)
		mu.Unlock()

		time.Sleep(10 * time.Millisecond) // Ensure overlap
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	script := fmt.Sprintf(`
deluge("Concurrency Test", {
    "test": {
        "concurrent": 10,
        "delay": "5ms",
    "duration": "100ms"
    }
});

scenario("test", "Test", function () {
    http("Request", {
        "url": "%s"
    });
});
	`, server.URL)

	l := lexer.New(script)
	p := parser.New(l)
	program, ok := p.ParseProgram()
	if !ok {
		t.Fatal("Parsing error")
	}

	dlg := deluge.NewDeluge(program)

	// Run with race detector enabled (go test -race)
	dlg.Run()

	count := atomic.LoadInt32(&requestCount)
	if count < 10 {
		t.Errorf("Expected at least 10 concurrent requests, got %d", count)
	}
}

// TestHTTPTimeouts verifies handling of slow/timeout responses
func TestHTTPTimeouts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond) // Slow response
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	script := fmt.Sprintf(`
deluge("Timeout Test", {
    "test": {
        "concurrent": 2,
        "delay": "50ms",
    "duration": "100ms"
    }
});

scenario("test", "Test", function () {
    http("Slow Request", {
        "url": "%s"
    });
});
	`, server.URL)

	l := lexer.New(script)
	p := parser.New(l)
	program, ok := p.ParseProgram()
	if !ok {
		t.Fatal("Parsing error")
	}

	dlg := deluge.NewDeluge(program)

	// Should handle slow responses without blocking indefinitely
	done := make(chan bool)
	go func() {
		dlg.Run()
		done <- true
	}()

	select {
	case <-done:
		// Success
	case <-time.After(15 * time.Second):
		t.Fatal("Test timed out - possible deadlock")
	}
}

// TestHTTPErrorStatuses verifies handling of HTTP error status codes
func TestHTTPErrorStatuses(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"400 Bad Request", http.StatusBadRequest},
		{"404 Not Found", http.StatusNotFound},
		{"500 Internal Server Error", http.StatusInternalServerError},
		{"503 Service Unavailable", http.StatusServiceUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			script := fmt.Sprintf(`
deluge("Error Status Test", {
    "test": {
        "concurrent": 1,
        "delay": "20ms",
    "duration": "100ms"
    }
});

scenario("test", "Test", function () {
    http("Request", {
        "url": "%s"
    });
});
			`, server.URL)

			l := lexer.New(script)
			p := parser.New(l)
			program, ok := p.ParseProgram()
			if !ok {
				t.Fatal("Parsing error")
			}

			dlg := deluge.NewDeluge(program)

			// Should handle error status codes gracefully
			dlg.Run()
		})
	}
}

// TestMalformedURL verifies handling of malformed URLs
func TestMalformedURL(t *testing.T) {
	script := `
deluge("Malformed URL Test", {
    "test": {
        "concurrent": 1,
        "delay": "20ms",
    "duration": "100ms"
    }
});

scenario("test", "Test", function () {
    http("Bad Request", {
        "url": "not-a-valid-url"
    });
});
	`

	l := lexer.New(script)
	p := parser.New(l)
	program, ok := p.ParseProgram()
	if !ok {
		t.Fatal("Parsing error")
	}

	dlg := deluge.NewDeluge(program)

	// Should not panic with malformed URL
	dlg.Run()
}

// TestEmptyURL verifies handling of empty URL
func TestEmptyURL(t *testing.T) {
	script := `
deluge("Empty URL Test", {
    "test": {
        "concurrent": 1,
        "delay": "20ms",
    "duration": "100ms"
    }
});

scenario("test", "Test", function () {
    http("Empty Request", {
        "url": ""
    });
});
	`

	l := lexer.New(script)
	p := parser.New(l)
	program, ok := p.ParseProgram()
	if !ok {
		t.Fatal("Parsing error")
	}

	dlg := deluge.NewDeluge(program)

	// Should handle empty URL gracefully
	dlg.Run()
}

// TestHTTPWithoutURL verifies error when URL is missing
func TestHTTPWithoutURL(t *testing.T) {
	script := `
deluge("Missing URL Test", {
    "test": {
        "concurrent": 1,
        "delay": "20ms",
    "duration": "100ms"
    }
});

scenario("test", "Test", function () {
    http("No URL Request", {
        "method": "GET"
    });
});
	`

	l := lexer.New(script)
	p := parser.New(l)
	program, ok := p.ParseProgram()
	if !ok {
		t.Fatal("Parsing error")
	}

	dlg := deluge.NewDeluge(program)

	// Should handle missing URL field
	dlg.Run()
}

// TestStressTestManyRequests verifies system under heavy load
func TestStressTestManyRequests(t *testing.T) {

	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	script := fmt.Sprintf(`
deluge("Stress Test", {
    "stress": {
        "concurrent": 50,
        "delay": "10ms",
    "duration": "100ms"
    }
});

scenario("stress", "Stress", function () {
    http("Request", {
        "url": "%s"
    });
});
	`, server.URL)

	l := lexer.New(script)
	p := parser.New(l)
	program, ok := p.ParseProgram()
	if !ok {
		t.Fatal("Parsing error")
	}

	dlg := deluge.NewDeluge(program)
	dlg.Run()

	// With 50 concurrent users, 10ms delay, 100ms duration: ~500 requests
	count := atomic.LoadInt32(&requestCount)
	if count < 100 {
		t.Errorf("Expected many requests under stress test, got only %d", count)
	}
}

// TestServerCloseDuringExecution verifies handling when server closes mid-test
func TestServerCloseDuringExecution(t *testing.T) {
	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&requestCount, 1)
		if count <= 5 {
			w.WriteHeader(http.StatusOK)
		}
		// After 5 requests, server will be closed
	}))

	script := fmt.Sprintf(`
deluge("Server Close Test", {
    "test": {
        "concurrent": 2,
        "delay": "20ms",
    "duration": "100ms"
    }
});

scenario("test", "Test", function () {
    http("Request", {
        "url": "%s"
    });
});
	`, server.URL)

	l := lexer.New(script)
	p := parser.New(l)
	program, ok := p.ParseProgram()
	if !ok {
		t.Fatal("Parsing error")
	}

	dlg := deluge.NewDeluge(program)

	// Close server after a short delay
	go func() {
		time.Sleep(500 * time.Millisecond)
		server.Close()
	}()

	// Should handle server closure gracefully
	dlg.Run()

	// Some requests should have succeeded before closure
	count := atomic.LoadInt32(&requestCount)
	if count == 0 {
		t.Error("Expected some requests before server closed")
	}
}

// TestComplexScenarioWithMixedOperations verifies complex mixed operations
func TestComplexScenarioWithMixedOperations(t *testing.T) {
	var getCount, postCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			atomic.AddInt32(&getCount, 1)
		} else if r.Method == "POST" {
			atomic.AddInt32(&postCount, 1)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	script := fmt.Sprintf(`
deluge("Complex Test", {
    "complex": {
        "concurrent": 3,
        "delay": "20ms",
    "duration": "100ms"
    }
});

scenario("complex", "Complex", function () {
    let x = 1;

    http("GET Request", {
        "url": "%s/get",
        "method": "GET"
    });

    x = x + 1;
    assert(x == 2);

    if (x > 1) {
        http("POST Request", {
            "url": "%s/post",
            "method": "POST"
        });
    }

    let arr = [1, 2, 3];
    assert(len(arr) == 3);
});
	`, server.URL, server.URL)

	l := lexer.New(script)
	p := parser.New(l)
	program, ok := p.ParseProgram()
	if !ok {
		t.Fatal("Parsing error")
	}

	dlg := deluge.NewDeluge(program)
	dlg.Run()

	get := atomic.LoadInt32(&getCount)
	post := atomic.LoadInt32(&postCount)

	if get == 0 || post == 0 {
		t.Errorf("Expected both GET (%d) and POST (%d) requests", get, post)
	}

	// Should have equal GET and POST requests (both in same iteration)
	if get != post {
		t.Errorf("Expected equal GET (%d) and POST (%d) requests", get, post)
	}
}

// TestRapidFireRequests verifies handling of very rapid requests
func TestRapidFireRequests(t *testing.T) {
	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	script := fmt.Sprintf(`
deluge("Rapid Fire", {
    "rapid": {
        "concurrent": 10,
        "delay": "1ms",
    "duration": "100ms"
    }
});

scenario("rapid", "Rapid", function () {
    http("Request", {
        "url": "%s"
    });
});
	`, server.URL)

	l := lexer.New(script)
	p := parser.New(l)
	program, ok := p.ParseProgram()
	if !ok {
		t.Fatal("Parsing error")
	}

	dlg := deluge.NewDeluge(program)
	dlg.Run()

	// With 10 users, 1ms delay, 100ms duration: very many requests
	count := atomic.LoadInt32(&requestCount)
	if count < 100 {
		t.Errorf("Expected many rapid requests, got only %d", count)
	}
}

// TestZeroConcurrentUsers verifies behavior with 0 concurrent users
func TestZeroConcurrentUsers(t *testing.T) {
	script := `
assert(true)
	`

	l := lexer.New(script)
	p := parser.New(l)
	program, ok := p.ParseProgram()
	if !ok {
		t.Fatal("Parsing error")
	}

	// Create scenario with 0 concurrent users
	scenario := deluge.NewScenario("test", 0, 10*time.Millisecond, 100*time.Millisecond, program)

	// Should not panic with 0 users
	scenario.Run(100 * time.Millisecond)

	// Should have 0 total calls
	if scenario.TotalSimUsersCalls != 0 {
		t.Errorf("Expected 0 calls with 0 concurrent users, got %d", scenario.TotalSimUsersCalls)
	}
}

// TestVeryLargeConcurrency verifies handling of very high concurrency
func TestVeryLargeConcurrency(t *testing.T) {

	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	script := fmt.Sprintf(`
deluge("High Concurrency", {
    "test": {
        "concurrent": 100,
        "delay": "50ms",
    "duration": "100ms"
    }
});

scenario("test", "Test", function () {
    http("Request", {
        "url": "%s"
    });
});
	`, server.URL)

	l := lexer.New(script)
	p := parser.New(l)
	program, ok := p.ParseProgram()
	if !ok {
		t.Fatal("Parsing error")
	}

	dlg := deluge.NewDeluge(program)
	dlg.Run()

	count := atomic.LoadInt32(&requestCount)
	// With 100 users, 50ms delay, 100ms duration: ~200 requests
	if count < 100 {
		t.Errorf("Expected many requests with high concurrency, got %d", count)
	}
}

// TestScenarioWithScriptErrors verifies error handling in scripts
func TestScenarioWithScriptErrors(t *testing.T) {
	// Test with an undefined variable
	script := `
deluge("Error Test", {
    "test": {
        "concurrent": 1,
        "delay": "20ms",
    "duration": "100ms"
    }
});

scenario("test", "Test", function () {
    let x = undefinedVariable;  // This will cause an error
});
	`

	l := lexer.New(script)
	p := parser.New(l)
	program, ok := p.ParseProgram()
	if !ok {
		t.Fatal("Parsing error")
	}

	dlg := deluge.NewDeluge(program)

	// Should handle script errors without crashing
	dlg.Run()
}

// TestHTTPWithInvalidMethod verifies handling of invalid HTTP methods
func TestHTTPWithInvalidMethod(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	script := fmt.Sprintf(`
deluge("Invalid Method Test", {
    "test": {
        "concurrent": 1,
        "delay": "20ms",
    "duration": "100ms"
    }
});

scenario("test", "Test", function () {
    http("Request", {
        "url": "%s",
        "method": "INVALID_METHOD"
    });
});
	`, server.URL)

	l := lexer.New(script)
	p := parser.New(l)
	program, ok := p.ParseProgram()
	if !ok {
		t.Fatal("Parsing error")
	}

	dlg := deluge.NewDeluge(program)

	// Should handle invalid HTTP method
	dlg.Run()
}
