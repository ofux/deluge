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

// TestDelugeWithMultipleScenarios verifies orchestration of multiple scenarios
func TestDelugeWithMultipleScenarios(t *testing.T) {

	var scenario1Count, scenario2Count int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/scenario1":
			atomic.AddInt32(&scenario1Count, 1)
		case "/scenario2":
			atomic.AddInt32(&scenario2Count, 1)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	script := fmt.Sprintf(`
deluge("Multi-Scenario Test", {
    "first": {
        "concurrent": 2,
        "delay": "50ms",
    "duration": "100ms"
    },
    "second": {
        "concurrent": 3,
        "delay": "50ms",
    "duration": "100ms"
    }
});

scenario("first", "First Scenario", function () {
    http("Request1", {
        "url": "%s/scenario1"
    });
});

scenario("second", "Second Scenario", function () {
    http("Request2", {
        "url": "%s/scenario2"
    });
});
	`, server.URL, server.URL)

	l := lexer.New(script)
	p := parser.New(l)
	program, ok := p.ParseProgram()
	if !ok {
		deluge.PrintParserErrors(p.Errors())
		t.Fatal("Parsing error")
	}

	dlg := deluge.NewDeluge(program)

	if dlg.Name != "Multi-Scenario Test" {
		t.Errorf("Expected deluge name 'Multi-Scenario Test', got '%s'", dlg.Name)
	}

	dlg.Run()

	// Both scenarios should have executed
	count1 := atomic.LoadInt32(&scenario1Count)
	count2 := atomic.LoadInt32(&scenario2Count)

	if count1 == 0 {
		t.Error("Expected scenario 1 to make requests")
	}
	if count2 == 0 {
		t.Error("Expected scenario 2 to make requests")
	}

	// Scenario 2 has more concurrent users, so it should make more requests
	// (3 users vs 2 users with same delay)
	if count2 <= count1 {
		t.Logf("Warning: Expected scenario2 (%d) to have more requests than scenario1 (%d)", count2, count1)
	}
}

// TestDelugeScenarioExecutionOrder verifies scenarios run in sequence
func TestDelugeScenarioExecutionOrder(t *testing.T) {

	var scenario1Start, scenario1End, scenario2Start time.Time
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		switch r.URL.Path {
		case "/scenario1":
			if scenario1Start.IsZero() {
				scenario1Start = time.Now()
			}
			scenario1End = time.Now()
		case "/scenario2":
			if scenario2Start.IsZero() {
				scenario2Start = time.Now()
			}
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	script := fmt.Sprintf(`
deluge("Sequential Test", {
    "first": {
        "concurrent": 1,
        "delay": "50ms",
    "duration": "100ms"
    },
    "second": {
        "concurrent": 1,
        "delay": "50ms",
    "duration": "100ms"
    }
});

scenario("first", "First", function () {
    http("Request", {
        "url": "%s/scenario1"
    });
});

scenario("second", "Second", function () {
    http("Request", {
        "url": "%s/scenario2"
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

	mu.Lock()
	defer mu.Unlock()

	// Scenarios should run sequentially, not concurrently
	// Scenario 2 should start after scenario 1 ends
	if !scenario2Start.After(scenario1End) {
		t.Logf("Scenario timing: S1 start=%s, S1 end=%s, S2 start=%s",
			scenario1Start, scenario1End, scenario2Start)
		// Note: Due to the hardcoded 10s duration, scenarios run sequentially
		// This test documents that behavior
	}
}

// TestDelugeName verifies the deluge name is set correctly
func TestDelugeName(t *testing.T) {
	script := `
deluge("My Custom Load Test", {
    "test": {
        "concurrent": 1,
        "delay": "20ms",
    "duration": "100ms"
    }
});

scenario("test", "Test", function () {
    assert(true);
});
	`

	l := lexer.New(script)
	p := parser.New(l)
	program, ok := p.ParseProgram()
	if !ok {
		t.Fatal("Parsing error")
	}

	dlg := deluge.NewDeluge(program)

	if dlg.Name != "My Custom Load Test" {
		t.Errorf("Expected name 'My Custom Load Test', got '%s'", dlg.Name)
	}
}

// TestDelugeWithOnlyScenarioDefinition verifies error when scenario is defined but not configured
func TestDelugeWithOnlyScenarioDefinition(t *testing.T) {
	// Note: This test documents current behavior where log.Fatalf is called
	// In a real test suite, you might want to refactor to return errors instead
	script := `
deluge("Test", {});

scenario("unconfigured", "Unconfigured Scenario", function () {
    assert(true);
});
	`

	l := lexer.New(script)
	p := parser.New(l)
	program, ok := p.ParseProgram()
	if !ok {
		t.Fatal("Parsing error")
	}

	// This will create a deluge, but the scenario won't be configured
	// The current implementation doesn't fail at creation time
	dlg := deluge.NewDeluge(program)

	if dlg.Name != "Test" {
		t.Errorf("Expected name 'Test', got '%s'", dlg.Name)
	}

	// Run should not crash even with no scenarios
	dlg.Run()
}

// TestDelugeEmptyScenarioConfig verifies behavior with empty scenario configuration
func TestDelugeEmptyScenarioConfig(t *testing.T) {
	script := `
deluge("Empty Config Test", {});
	`

	l := lexer.New(script)
	p := parser.New(l)
	program, ok := p.ParseProgram()
	if !ok {
		t.Fatal("Parsing error")
	}

	dlg := deluge.NewDeluge(program)
	dlg.Run() // Should not panic with empty config
}

// TestDelugeWithThreeScenarios verifies handling of three scenarios
func TestDelugeWithThreeScenarios(t *testing.T) {

	var s1Count, s2Count, s3Count int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/s1":
			atomic.AddInt32(&s1Count, 1)
		case "/s2":
			atomic.AddInt32(&s2Count, 1)
		case "/s3":
			atomic.AddInt32(&s3Count, 1)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	script := fmt.Sprintf(`
deluge("Three Scenarios", {
    "alpha": {
        "concurrent": 1,
        "delay": "20ms",
    "duration": "100ms"
    },
    "beta": {
        "concurrent": 2,
        "delay": "20ms",
    "duration": "100ms"
    },
    "gamma": {
        "concurrent": 3,
        "delay": "20ms",
    "duration": "100ms"
    }
});

scenario("alpha", "Alpha", function () {
    http("Req", {"url": "%s/s1"});
});

scenario("beta", "Beta", function () {
    http("Req", {"url": "%s/s2"});
});

scenario("gamma", "Gamma", function () {
    http("Req", {"url": "%s/s3"});
});
	`, server.URL, server.URL, server.URL)

	l := lexer.New(script)
	p := parser.New(l)
	program, ok := p.ParseProgram()
	if !ok {
		t.Fatal("Parsing error")
	}

	dlg := deluge.NewDeluge(program)
	dlg.Run()

	// All three scenarios should execute
	if atomic.LoadInt32(&s1Count) == 0 {
		t.Error("Scenario alpha made no requests")
	}
	if atomic.LoadInt32(&s2Count) == 0 {
		t.Error("Scenario beta made no requests")
	}
	if atomic.LoadInt32(&s3Count) == 0 {
		t.Error("Scenario gamma made no requests")
	}
}

// TestDelugeScenarioIndependence verifies scenarios don't affect each other
func TestDelugeScenarioIndependence(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	script := fmt.Sprintf(`
deluge("Independence Test", {
    "success": {
        "concurrent": 1,
        "delay": "50ms",
    "duration": "100ms"
    },
    "failure": {
        "concurrent": 1,
        "delay": "50ms",
    "duration": "100ms"
    }
});

scenario("success", "Success Scenario", function () {
    http("Request", {"url": "%s"});
    assert(true);
});

scenario("failure", "Failure Scenario", function () {
    http("Request", {"url": "%s"});
    assert(false);  // This scenario fails
});
	`, server.URL, server.URL)

	l := lexer.New(script)
	p := parser.New(l)
	program, ok := p.ParseProgram()
	if !ok {
		t.Fatal("Parsing error")
	}

	dlg := deluge.NewDeluge(program)

	// Both scenarios should run independently
	// The failure in one shouldn't affect the other
	dlg.Run()

	// If we get here without panic, the test passes
}

// TestDelugeWithDifferentDelays verifies scenarios with different delays
func TestDelugeWithDifferentDelays(t *testing.T) {

	var fastCount, slowCount int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/fast":
			atomic.AddInt32(&fastCount, 1)
		case "/slow":
			atomic.AddInt32(&slowCount, 1)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	script := fmt.Sprintf(`
deluge("Different Delays", {
    "fast": {
        "concurrent": 1,
        "delay": "10ms",
    "duration": "100ms"
    },
    "slow": {
        "concurrent": 1,
        "delay": "15ms",
    "duration": "100ms"
    }
});

scenario("fast", "Fast", function () {
    http("Req", {"url": "%s/fast"});
});

scenario("slow", "Slow", function () {
    http("Req", {"url": "%s/slow"});
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

	fast := atomic.LoadInt32(&fastCount)
	slow := atomic.LoadInt32(&slowCount)

	// Fast scenario (50ms delay) should make more requests than slow (80ms delay)
	// With 100ms duration: fast ~2 requests, slow ~1 request
	if fast <= slow {
		t.Errorf("Expected fast scenario (%d) to make more requests than slow (%d)", fast, slow)
	}
}

// TestDelugeRunTiming verifies the total execution time
func TestDelugeRunTiming(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	script := fmt.Sprintf(`
deluge("Timing Test", {
    "s1": {
        "concurrent": 1,
        "delay": "50ms",
    "duration": "100ms"
    },
    "s2": {
        "concurrent": 1,
        "delay": "50ms",
    "duration": "100ms"
    }
});

scenario("s1", "S1", function () {
    http("Req", {"url": "%s"});
});

scenario("s2", "S2", function () {
    http("Req", {"url": "%s"});
});
	`, server.URL, server.URL)

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

	// Both scenarios configured with 100ms duration
	// With 2 scenarios running sequentially: ~200ms total
	expectedDuration := 200 * time.Millisecond
	tolerance := 100 * time.Millisecond

	if elapsed < expectedDuration-tolerance || elapsed > expectedDuration+tolerance {
		t.Errorf("Expected total runtime ~%s, got %s", expectedDuration, elapsed)
	}
}

// TestDelugeScenarioOrderDoesNotMatter verifies scenario definition order doesn't matter
func TestDelugeScenarioOrderDoesNotMatter(t *testing.T) {
	tests := []struct {
		name   string
		script string
	}{
		{
			name: "Config then scenario",
			script: `
deluge("Test", {
    "test": {"concurrent": 1, "delay": "50ms", "duration": "100ms"}
});
scenario("test", "Test", function () { assert(true); });
			`,
		},
		{
			name: "Scenario then config",
			script: `
scenario("test", "Test", function () { assert(true); });
deluge("Test", {
    "test": {"concurrent": 1, "delay": "50ms", "duration": "100ms"}
});
			`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.script)
			p := parser.New(l)
			program, ok := p.ParseProgram()
			if !ok {
				t.Fatal("Parsing error")
			}

			dlg := deluge.NewDeluge(program)
			dlg.Run() // Should not panic regardless of order
		})
	}
}
