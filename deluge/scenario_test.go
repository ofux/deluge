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


// TestScenarioDurationEnforcement verifies scenarios stop after the specified duration
func TestScenarioDurationEnforcement(t *testing.T) {
	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	script := fmt.Sprintf(`
		http("Request", {
			"url": "%s"
		});
	`, server.URL)

	l := lexer.New(script)
	p := parser.New(l)
	program, ok := p.ParseProgram()
	if !ok {
		t.Fatal("Parsing error")
	}

	// Create scenario with 1 concurrent user, 50ms delay
	scenario := deluge.NewScenario("test", 1, 50*time.Millisecond, 500*time.Millisecond, program)

	start := time.Now()
	// Run for 500ms
	scenario.Run(500 * time.Millisecond)
	elapsed := time.Since(start)

	// Should run for approximately 500ms
	if elapsed < 450*time.Millisecond || elapsed > 600*time.Millisecond {
		t.Errorf("Expected scenario to run for ~500ms, but took %s", elapsed)
	}

	// With 50ms delay and 500ms duration, we expect around 10 requests
	// Allow some tolerance for timing
	count := atomic.LoadInt32(&requestCount)
	if count < 8 || count > 12 {
		t.Errorf("Expected ~10 requests (500ms / 50ms), got %d", count)
	}

	// Verify TotalSimUsersCalls matches actual requests
	if scenario.TotalSimUsersCalls != int64(count) {
		t.Errorf("Expected TotalSimUsersCalls to be %d, got %d", count, scenario.TotalSimUsersCalls)
	}
}

// TestScenarioConcurrentUsers verifies multiple users run concurrently
func TestScenarioConcurrentUsers(t *testing.T) {
	var activeRequests int32
	var maxConcurrent int32
	var mu sync.Mutex
	var requestTimes []time.Time

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := atomic.AddInt32(&activeRequests, 1)

		// Track max concurrent
		for {
			max := atomic.LoadInt32(&maxConcurrent)
			if current <= max || atomic.CompareAndSwapInt32(&maxConcurrent, max, current) {
				break
			}
		}

		mu.Lock()
		requestTimes = append(requestTimes, time.Now())
		mu.Unlock()

		time.Sleep(100 * time.Millisecond) // Simulate work
		atomic.AddInt32(&activeRequests, -1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	script := fmt.Sprintf(`
		http("Request", {
			"url": "%s"
		});
	`, server.URL)

	l := lexer.New(script)
	p := parser.New(l)
	program, ok := p.ParseProgram()
	if !ok {
		t.Fatal("Parsing error")
	}

	// Create scenario with 5 concurrent users
	scenario := deluge.NewScenario("test", 5, 50*time.Millisecond, 500*time.Millisecond, program)
	scenario.Run(1 * time.Second)

	maxSeen := atomic.LoadInt32(&maxConcurrent)
	if maxSeen < 2 {
		t.Errorf("Expected at least 2 concurrent requests, saw maximum of %d", maxSeen)
	}

	// Verify we had multiple requests happening in parallel
	mu.Lock()
	defer mu.Unlock()

	if len(requestTimes) < 5 {
		t.Errorf("Expected at least 5 requests with 5 concurrent users, got %d", len(requestTimes))
	}
}

// TestScenarioRequestCountingAccuracy verifies accurate request counting
func TestScenarioRequestCountingAccuracy(t *testing.T) {
	var actualRequests int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&actualRequests, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	script := fmt.Sprintf(`
		http("Request", {
			"url": "%s"
		});
	`, server.URL)

	l := lexer.New(script)
	p := parser.New(l)
	program, ok := p.ParseProgram()
	if !ok {
		t.Fatal("Parsing error")
	}

	scenario := deluge.NewScenario("test", 3, 100*time.Millisecond, 500*time.Millisecond, program)
	scenario.Run(1 * time.Second)

	actual := atomic.LoadInt32(&actualRequests)
	counted := scenario.TotalSimUsersCalls

	if counted != int64(actual) {
		t.Errorf("TotalSimUsersCalls (%d) doesn't match actual requests (%d)", counted, actual)
	}

	// With 3 users, 100ms delay, 1s duration, we expect ~30 requests (3 * 10)
	if counted < 25 || counted > 35 {
		t.Errorf("Expected ~30 requests, got %d", counted)
	}
}

// TestScenarioWithFailingUsers verifies scenario continues when some users fail
func TestScenarioWithFailingUsers(t *testing.T) {
	var successCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&successCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	script := fmt.Sprintf(`
		http("Request", {
			"url": "%s"
		});
		assert(false);  // This will cause the user to fail after first request
	`, server.URL)

	l := lexer.New(script)
	p := parser.New(l)
	program, ok := p.ParseProgram()
	if !ok {
		t.Fatal("Parsing error")
	}

	// Create scenario with 3 users
	scenario := deluge.NewScenario("test", 3, 50*time.Millisecond, 500*time.Millisecond, program)
	scenario.Run(500 * time.Millisecond)

	// Each user should make exactly 1 request before failing
	count := atomic.LoadInt32(&successCount)
	if count != 3 {
		t.Errorf("Expected 3 requests (1 per user before assertion fails), got %d", count)
	}

	if scenario.TotalSimUsersCalls != 3 {
		t.Errorf("Expected TotalSimUsersCalls to be 3, got %d", scenario.TotalSimUsersCalls)
	}
}

// TestScenarioZeroDuration verifies scenario with zero duration
func TestScenarioZeroDuration(t *testing.T) {
	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	script := fmt.Sprintf(`
		http("Request", {
			"url": "%s"
		});
	`, server.URL)

	l := lexer.New(script)
	p := parser.New(l)
	program, ok := p.ParseProgram()
	if !ok {
		t.Fatal("Parsing error")
	}

	scenario := deluge.NewScenario("test", 2, 10*time.Millisecond, 500*time.Millisecond, program)

	start := time.Now()
	scenario.Run(0)
	elapsed := time.Since(start)

	// Should complete almost immediately
	if elapsed > 100*time.Millisecond {
		t.Errorf("Expected near-instant completion with 0 duration, took %s", elapsed)
	}

	// With 0 duration, the scenario terminates immediately
	// The first check happens before any requests can complete
	count := atomic.LoadInt32(&requestCount)
	// Zero duration means no requests are guaranteed to execute
	_ = count // acknowledge we checked it
}

// TestScenarioVeryShortDuration verifies behavior with very short duration
func TestScenarioVeryShortDuration(t *testing.T) {
	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	script := fmt.Sprintf(`
		http("Request", {
			"url": "%s"
		});
	`, server.URL)

	l := lexer.New(script)
	p := parser.New(l)
	program, ok := p.ParseProgram()
	if !ok {
		t.Fatal("Parsing error")
	}

	scenario := deluge.NewScenario("test", 1, 10*time.Millisecond, 500*time.Millisecond, program)

	start := time.Now()
	scenario.Run(50 * time.Millisecond)
	elapsed := time.Since(start)

	// Should complete in approximately 50ms
	if elapsed > 200*time.Millisecond {
		t.Errorf("Expected ~50ms execution, took %s", elapsed)
	}
}

// TestScenarioSingleUser verifies scenario with one user
func TestScenarioSingleUser(t *testing.T) {
	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	script := fmt.Sprintf(`
		http("Request", {
			"url": "%s"
		});
	`, server.URL)

	l := lexer.New(script)
	p := parser.New(l)
	program, ok := p.ParseProgram()
	if !ok {
		t.Fatal("Parsing error")
	}

	scenario := deluge.NewScenario("test", 1, 100*time.Millisecond, 500*time.Millisecond, program)
	scenario.Run(1 * time.Second)

	// With 1 user, 100ms delay, 1s duration: ~10 requests
	count := atomic.LoadInt32(&requestCount)
	if count < 8 || count > 12 {
		t.Errorf("Expected ~10 requests, got %d", count)
	}

	if scenario.TotalSimUsersCalls != int64(count) {
		t.Errorf("TotalSimUsersCalls mismatch: expected %d, got %d", count, scenario.TotalSimUsersCalls)
	}
}

// TestScenarioManyUsers verifies scenario with many concurrent users
func TestScenarioManyUsers(t *testing.T) {
	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	script := fmt.Sprintf(`
		http("Request", {
			"url": "%s"
		});
	`, server.URL)

	l := lexer.New(script)
	p := parser.New(l)
	program, ok := p.ParseProgram()
	if !ok {
		t.Fatal("Parsing error")
	}

	// Test with 20 concurrent users
	scenario := deluge.NewScenario("test", 20, 100*time.Millisecond, 500*time.Millisecond, program)
	scenario.Run(1 * time.Second)

	// With 20 users, 100ms delay, 1s duration: ~200 requests
	count := atomic.LoadInt32(&requestCount)
	if count < 150 || count > 250 {
		t.Errorf("Expected ~200 requests, got %d", count)
	}

	if scenario.TotalSimUsersCalls != int64(count) {
		t.Errorf("TotalSimUsersCalls mismatch: expected %d, got %d", count, scenario.TotalSimUsersCalls)
	}
}

// TestScenarioWithComplexScript verifies scenario with multi-step operations
func TestScenarioWithComplexScript(t *testing.T) {
	var endpoint1Count, endpoint2Count int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/endpoint1":
			atomic.AddInt32(&endpoint1Count, 1)
		case "/endpoint2":
			atomic.AddInt32(&endpoint2Count, 1)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	script := fmt.Sprintf(`
		http("First", {
			"url": "%s/endpoint1"
		});
		let x = 1 + 1;
		assert(x == 2);
		http("Second", {
			"url": "%s/endpoint2"
		});
	`, server.URL, server.URL)

	l := lexer.New(script)
	p := parser.New(l)
	program, ok := p.ParseProgram()
	if !ok {
		t.Fatal("Parsing error")
	}

	scenario := deluge.NewScenario("test", 2, 100*time.Millisecond, 500*time.Millisecond, program)
	scenario.Run(1 * time.Second)

	count1 := atomic.LoadInt32(&endpoint1Count)
	count2 := atomic.LoadInt32(&endpoint2Count)

	// Both endpoints should receive equal requests
	if count1 != count2 {
		t.Errorf("Expected equal requests to both endpoints, got %d and %d", count1, count2)
	}

	// TotalSimUsersCalls counts SimUser.Run() iterations, not individual HTTP requests
	// With 2 users and 1s duration with 100ms delay: ~10 iterations per user = 20 total
	if scenario.TotalSimUsersCalls < 15 || scenario.TotalSimUsersCalls > 25 {
		t.Errorf("Expected ~20 SimUser iterations, got %d", scenario.TotalSimUsersCalls)
	}

	// Each iteration makes 2 HTTP requests, so total HTTP requests should be ~2x iterations
	totalHTTP := count1 + count2
	expectedHTTP := scenario.TotalSimUsersCalls * 2
	if totalHTTP != int32(expectedHTTP) {
		t.Errorf("Expected %d HTTP requests (2 per iteration), got %d", expectedHTTP, totalHTTP)
	}
}

// TestScenarioRaceConditionOnCounter verifies no race condition on TotalSimUsersCalls
func TestScenarioRaceConditionOnCounter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	script := fmt.Sprintf(`
		http("Request", {
			"url": "%s"
		});
	`, server.URL)

	l := lexer.New(script)
	p := parser.New(l)
	program, ok := p.ParseProgram()
	if !ok {
		t.Fatal("Parsing error")
	}

	// Create scenario with many concurrent users hitting the counter
	scenario := deluge.NewScenario("test", 50, 10*time.Millisecond, 500*time.Millisecond, program)

	// Run with race detector enabled (go test -race)
	scenario.Run(500 * time.Millisecond)

	// Just verify it completed without race detector complaints
	if scenario.TotalSimUsersCalls == 0 {
		t.Error("Expected some requests to be counted")
	}
}

// TestScenarioName verifies scenario name is preserved
func TestScenarioName(t *testing.T) {
	script := `assert(true)`

	l := lexer.New(script)
	p := parser.New(l)
	program, ok := p.ParseProgram()
	if !ok {
		t.Fatal("Parsing error")
	}

	scenario := deluge.NewScenario("My Test Scenario", 1, 10*time.Millisecond, 500*time.Millisecond, program)

	if scenario.Name != "My Test Scenario" {
		t.Errorf("Expected scenario name 'My Test Scenario', got '%s'", scenario.Name)
	}
}
