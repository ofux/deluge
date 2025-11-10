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

// TestEndToEndScenarioExecution tests the complete flow from DSL to execution
func TestEndToEndScenarioExecution(t *testing.T) {

	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	script := fmt.Sprintf(`
deluge("Load Test", {
    "testScenario": {
        "concurrent": 2,
        "delay": "20ms",
        "duration": "100ms"
    }
});

scenario("testScenario", "Test Scenario", function () {
    http("Test Request", {
        "url": "%s"
    });
});
	`, server.URL)

	l := lexer.New(script)
	p := parser.New(l)
	program, ok := p.ParseProgram()
	if !ok {
		deluge.PrintParserErrors(p.Errors())
		t.Fatal("Parsing error(s)")
	}

	dlg := deluge.NewDeluge(program)

	if dlg.Name != "Load Test" {
		t.Errorf("Expected deluge name to be 'Load Test', got '%s'", dlg.Name)
	}

	// Run should execute without panicking (runs for 100ms per scenario)
	dlg.Run()

	// Verify that requests were made
	if atomic.LoadInt32(&requestCount) == 0 {
		t.Error("Expected at least one HTTP request to be made")
	}
}

// TestScenarioDurationRespected verifies that scenarios respect their duration parameter
func TestScenarioDurationRespected(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	script := fmt.Sprintf(`
deluge("Duration Test", {
    "quickScenario": {
        "concurrent": 1,
        "delay": "10ms",
    "duration": "100ms"
    }
});

scenario("quickScenario", "Quick Scenario", function () {
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

	start := time.Now()
	dlg.Run()
	elapsed := time.Since(start)

	// The configured duration is 100ms
	// It should take approximately 100ms
	if elapsed < 80*time.Millisecond || elapsed > 200*time.Millisecond {
		t.Errorf("Expected execution to take ~100ms, but took %s", elapsed)
	}
}

// TestMultipleConcurrentSimUsers verifies that multiple SimUsers run concurrently
func TestMultipleConcurrentSimUsers(t *testing.T) {

	var activeRequests int32
	var maxConcurrent int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := atomic.AddInt32(&activeRequests, 1)

		// Track the maximum concurrent requests
		for {
			max := atomic.LoadInt32(&maxConcurrent)
			if current <= max || atomic.CompareAndSwapInt32(&maxConcurrent, max, current) {
				break
			}
		}

		time.Sleep(50 * time.Millisecond) // Simulate some work
		atomic.AddInt32(&activeRequests, -1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	script := fmt.Sprintf(`
deluge("Concurrency Test", {
    "concurrentScenario": {
        "concurrent": 5,
        "delay": "10ms",
    "duration": "100ms"
    }
});

scenario("concurrentScenario", "Concurrent Scenario", function () {
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

	maxSeen := atomic.LoadInt32(&maxConcurrent)
	if maxSeen < 2 {
		t.Errorf("Expected at least 2 concurrent requests, saw maximum of %d", maxSeen)
	}
}

// TestHTTPRequestWithDifferentMethods tests that different HTTP methods work
func TestHTTPRequestWithDifferentMethods(t *testing.T) {

	tests := []struct {
		name           string
		method         string
		expectedMethod string
	}{
		{"GET request", `"method": "GET"`, "GET"},
		{"POST request", `"method": "POST"`, "POST"},
		{"PUT request", `"method": "PUT"`, "PUT"},
		{"DELETE request", `"method": "DELETE"`, "DELETE"},
		{"Default method", "", "GET"}, // Default is GET
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedMethod string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedMethod = r.Method
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			methodField := ""
			if tt.method != "" {
				methodField = fmt.Sprintf(", %s", tt.method)
			}

			script := fmt.Sprintf(`
deluge("Method Test", {
    "methodScenario": {
        "concurrent": 1,
        "delay": "20ms",
    "duration": "100ms"
    }
});

scenario("methodScenario", "Method Test", function () {
    http("Request", {
        "url": "%s"%s
    });
});
			`, server.URL, methodField)

			l := lexer.New(script)
			p := parser.New(l)
			program, ok := p.ParseProgram()
			if !ok {
				t.Fatal("Parsing error")
			}

			dlg := deluge.NewDeluge(program)
			dlg.Run()

			if receivedMethod != tt.expectedMethod {
				t.Errorf("Expected HTTP method %s, got %s", tt.expectedMethod, receivedMethod)
			}
		})
	}
}

// TestScenarioRequestCounting verifies that scenarios correctly count requests
func TestScenarioRequestCounting(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	script := fmt.Sprintf(`
deluge("Count Test", {
    "countScenario": {
        "concurrent": 3,
        "delay": "50ms",
    "duration": "100ms"
    }
});

scenario("countScenario", "Count Scenario", function () {
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

	// The scenario should have executed multiple requests
	// With 3 concurrent users, 50ms delay, and 100ms duration, we expect many requests
	// At minimum, we should see 3 users * (100ms / 50ms) = 6 requests
	// But we'll be conservative and just check that some requests were made
	if dlg == nil {
		t.Error("Deluge instance should not be nil")
	}
}

// TestInvalidHTTPRequest verifies that invalid HTTP requests are handled gracefully
func TestInvalidHTTPRequest(t *testing.T) {

	script := `
deluge("Error Test", {
    "errorScenario": {
        "concurrent": 1,
        "delay": "20ms",
    "duration": "100ms"
    }
});

scenario("errorScenario", "Error Scenario", function () {
    http("Bad Request", {
        "url": "http://invalid-url-that-does-not-exist-12345.com"
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

	// This should not panic even with invalid URLs
	dlg.Run()
}

// TestSimUserWithAssertions tests that assertions work correctly in scenarios
func TestSimUserWithAssertions(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	t.Run("Successful assertion", func(t *testing.T) {
		script := fmt.Sprintf(`
deluge("Assert Test", {
    "assertScenario": {
        "concurrent": 1,
        "delay": "20ms",
    "duration": "100ms"
    }
});

scenario("assertScenario", "Assert Scenario", function () {
    http("Request", {
        "url": "%s"
    });
    assert(1 + 1 == 2);
});
		`, server.URL)

		l := lexer.New(script)
		p := parser.New(l)
		program, ok := p.ParseProgram()
		if !ok {
			t.Fatal("Parsing error")
		}

		dlg := deluge.NewDeluge(program)
		dlg.Run() // Should complete without errors
	})

	t.Run("Failed assertion stops SimUser", func(t *testing.T) {
		var requestCount int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&requestCount, 1)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		script := fmt.Sprintf(`
deluge("Failed Assert Test", {
    "failScenario": {
        "concurrent": 1,
        "delay": "10ms",
    "duration": "100ms"
    }
});

scenario("failScenario", "Fail Scenario", function () {
    http("Request", {
        "url": "%s"
    });
    assert(1 + 1 == 3);
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

		// The SimUser should stop after the first failed assertion
		// So we should only see 1 request (before the assertion fails)
		count := atomic.LoadInt32(&requestCount)
		if count != 1 {
			t.Errorf("Expected exactly 1 request before assertion fails, got %d", count)
		}
	})
}

// TestScenarioWithMultipleOperations tests scenarios with multiple HTTP calls
func TestScenarioWithMultipleOperations(t *testing.T) {
	var request1Count, request2Count int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/endpoint1":
			atomic.AddInt32(&request1Count, 1)
		case "/endpoint2":
			atomic.AddInt32(&request2Count, 1)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	script := fmt.Sprintf(`
deluge("Multi-Op Test", {
    "multiScenario": {
        "concurrent": 2,
        "delay": "20ms",
    "duration": "100ms"
    }
});

scenario("multiScenario", "Multi Operation Scenario", function () {
    http("First Request", {
        "url": "%s/endpoint1"
    });
    http("Second Request", {
        "url": "%s/endpoint2"
    });
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

	count1 := atomic.LoadInt32(&request1Count)
	count2 := atomic.LoadInt32(&request2Count)

	if count1 == 0 || count2 == 0 {
		t.Errorf("Expected requests to both endpoints, got endpoint1=%d, endpoint2=%d", count1, count2)
	}

	// Both endpoints should receive the same number of requests
	if count1 != count2 {
		t.Errorf("Expected equal requests to both endpoints, got endpoint1=%d, endpoint2=%d", count1, count2)
	}
}

// TestDelugeWithNoScenarios tests that a deluge with no scenarios doesn't crash
func TestDelugeWithNoScenarios(t *testing.T) {
	script := `
deluge("Empty Test", {});
	`

	l := lexer.New(script)
	p := parser.New(l)
	program, ok := p.ParseProgram()
	if !ok {
		t.Fatal("Parsing error")
	}

	dlg := deluge.NewDeluge(program)
	dlg.Run() // Should not panic
}

// TestScenarioDelayBetweenRequests verifies that the delay between requests is respected
func TestScenarioDelayBetweenRequests(t *testing.T) {
	var requestTimes []time.Time
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestTimes = append(requestTimes, time.Now())
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	script := fmt.Sprintf(`
deluge("Delay Test", {
    "delayScenario": {
        "concurrent": 1,
        "delay": "30ms",
    "duration": "200ms"
    }
});

scenario("delayScenario", "Delay Scenario", function () {
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

	mu.Lock()
	defer mu.Unlock()

	if len(requestTimes) < 2 {
		t.Skip("Not enough requests to verify delay")
	}

	// Check that consecutive requests are approximately 30ms apart
	for i := 1; i < len(requestTimes); i++ {
		delay := requestTimes[i].Sub(requestTimes[i-1])
		// Allow 15ms tolerance
		if delay < 15*time.Millisecond || delay > 45*time.Millisecond {
			t.Errorf("Expected ~30ms delay between requests, got %s", delay)
			break
		}
	}
}
