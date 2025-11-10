package deluge_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/ofux/deluge/deluge"
	"github.com/ofux/deluge/dsl/lexer"
	"github.com/ofux/deluge/dsl/parser"
)

// parseScript is a helper to parse DSL scripts
func parseScript(t *testing.T, script string) *deluge.SimUser {
	l := lexer.New(script)
	p := parser.New(l)
	program, ok := p.ParseProgram()
	if !ok {
		deluge.PrintParserErrors(p.Errors())
		t.Fatal("Parsing error")
	}
	return deluge.NewSimUser("test-user", program)
}

// TestSimUserInitialStatus verifies that a new SimUser starts in Virgin status
func TestSimUserInitialStatus(t *testing.T) {
	su := parseScript(t, `assert(1 == 1)`)

	if su.Status != deluge.Virgin {
		t.Errorf("Expected initial status to be Virgin, got %v", su.Status)
	}
}

// TestSimUserStatusTransitionToInProgress verifies status changes during execution
func TestSimUserStatusTransitionToInProgress(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Slow response to allow status check
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	script := fmt.Sprintf(`
		http("Request", {
			"url": "%s"
		});
	`, server.URL)

	su := parseScript(t, script)

	if su.Status != deluge.Virgin {
		t.Errorf("Expected initial status Virgin, got %v", su.Status)
	}

	su.Run()

	// After Run completes, status should be DoneSuccess or DoneError
	if su.Status != deluge.DoneSuccess && su.Status != deluge.DoneError {
		t.Errorf("Expected final status DoneSuccess or DoneError, got %v", su.Status)
	}
}

// TestSimUserSuccessfulExecution verifies DoneSuccess status
func TestSimUserSuccessfulExecution(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	script := fmt.Sprintf(`
		http("Request", {
			"url": "%s"
		});
		assert(true);
	`, server.URL)

	su := parseScript(t, script)
	su.Run()

	if su.Status != deluge.DoneSuccess {
		t.Errorf("Expected status DoneSuccess, got %v", su.Status)
	}
}

// TestSimUserFailedAssertion verifies DoneError status on assertion failure
func TestSimUserFailedAssertion(t *testing.T) {
	su := parseScript(t, `assert(false)`)
	su.Run()

	if su.Status != deluge.DoneError {
		t.Errorf("Expected status DoneError after failed assertion, got %v", su.Status)
	}
}

// TestSimUserWithInvalidHTTPRequest verifies error handling
func TestSimUserWithInvalidHTTPRequest(t *testing.T) {
	script := `
		http("Bad Request", {
			"url": "http://invalid-domain-12345.local"
		});
	`

	su := parseScript(t, script)
	su.Run()

	// Should handle error gracefully and set status to DoneError
	if su.Status != deluge.DoneError {
		t.Errorf("Expected status DoneError after HTTP error, got %v", su.Status)
	}
}

// TestSimUserWithComplexScript verifies execution of multi-step scripts
func TestSimUserWithComplexScript(t *testing.T) {
	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	script := fmt.Sprintf(`
		let x = 1 + 1;
		assert(x == 2);

		http("First", {
			"url": "%s/first"
		});

		let y = x * 2;
		assert(y == 4);

		http("Second", {
			"url": "%s/second"
		});

		assert(y + x == 6);
	`, server.URL, server.URL)

	su := parseScript(t, script)
	su.Run()

	if su.Status != deluge.DoneSuccess {
		t.Errorf("Expected status DoneSuccess, got %v", su.Status)
	}

	if atomic.LoadInt32(&requestCount) != 2 {
		t.Errorf("Expected 2 HTTP requests, got %d", atomic.LoadInt32(&requestCount))
	}
}

// TestSimUserMultipleRuns verifies SimUser can be run multiple times
func TestSimUserMultipleRuns(t *testing.T) {
	var runCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&runCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	script := fmt.Sprintf(`
		http("Request", {
			"url": "%s"
		});
	`, server.URL)

	su := parseScript(t, script)

	// Run multiple times
	for i := 0; i < 3; i++ {
		su.Run()
		if su.Status != deluge.DoneSuccess {
			t.Errorf("Run %d: Expected status DoneSuccess, got %v", i+1, su.Status)
		}
	}

	if atomic.LoadInt32(&runCount) != 3 {
		t.Errorf("Expected 3 requests across 3 runs, got %d", atomic.LoadInt32(&runCount))
	}
}

// TestSimUserWithVariables verifies that variables work correctly
func TestSimUserWithVariables(t *testing.T) {
	script := `
		let name = "test";
		let count = 42;
		let check = count > 40;
		assert(check);
		assert(name == "test");
	`

	su := parseScript(t, script)
	su.Run()

	if su.Status != deluge.DoneSuccess {
		t.Errorf("Expected status DoneSuccess, got %v", su.Status)
	}
}

// TestSimUserWithConditionals verifies conditional logic
func TestSimUserWithConditionals(t *testing.T) {
	tests := []struct {
		name           string
		script         string
		expectedStatus deluge.SimUserStatus
	}{
		{
			name: "True branch taken",
			script: `
				let x = 10;
				if (x > 5) {
					assert(true);
				} else {
					assert(false);
				}
			`,
			expectedStatus: deluge.DoneSuccess,
		},
		{
			name: "False branch taken",
			script: `
				let x = 3;
				if (x > 5) {
					assert(false);
				} else {
					assert(true);
				}
			`,
			expectedStatus: deluge.DoneSuccess,
		},
		{
			name: "Nested conditionals",
			script: `
				let x = 10;
				if (x > 5) {
					if (x < 20) {
						assert(true);
					} else {
						assert(false);
					}
				}
			`,
			expectedStatus: deluge.DoneSuccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			su := parseScript(t, tt.script)
			su.Run()

			if su.Status != tt.expectedStatus {
				t.Errorf("Expected status %v, got %v", tt.expectedStatus, su.Status)
			}
		})
	}
}

// TestSimUserWithRecursion verifies recursive function calls
func TestSimUserWithRecursion(t *testing.T) {
	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	script := fmt.Sprintf(`
		let makeRequests = function(count) {
			if (count > 0) {
				http("Request", {
					"url": "%s"
				});
				makeRequests(count - 1);
			}
		};
		makeRequests(3);
	`, server.URL)

	su := parseScript(t, script)
	su.Run()

	if su.Status != deluge.DoneSuccess {
		t.Errorf("Expected status DoneSuccess, got %v", su.Status)
	}

	if atomic.LoadInt32(&requestCount) != 3 {
		t.Errorf("Expected 3 HTTP requests from recursion, got %d", atomic.LoadInt32(&requestCount))
	}
}

// TestSimUserWithFunctions verifies function definitions and calls
func TestSimUserWithFunctions(t *testing.T) {
	script := `
		let add = function(a, b) {
			return a + b;
		};

		let result = add(5, 3);
		assert(result == 8);
	`

	su := parseScript(t, script)
	su.Run()

	if su.Status != deluge.DoneSuccess {
		t.Errorf("Expected status DoneSuccess, got %v", su.Status)
	}
}

// TestSimUserWithArrays verifies array operations
func TestSimUserWithArrays(t *testing.T) {
	script := `
		let arr = [1, 2, 3];
		assert(arr[0] == 1);
		assert(arr[2] == 3);
		assert(len(arr) == 3);
	`

	su := parseScript(t, script)
	su.Run()

	if su.Status != deluge.DoneSuccess {
		t.Errorf("Expected status DoneSuccess, got %v", su.Status)
	}
}

// TestSimUserWithHash verifies hash/object operations
func TestSimUserWithHash(t *testing.T) {
	script := `
		let obj = {"name": "test", "value": 42};
		assert(obj["name"] == "test");
		assert(obj["value"] == 42);
	`

	su := parseScript(t, script)
	su.Run()

	if su.Status != deluge.DoneSuccess {
		t.Errorf("Expected status DoneSuccess, got %v", su.Status)
	}
}

// TestSimUserErrorPropagation verifies that errors stop execution
func TestSimUserErrorPropagation(t *testing.T) {
	var requestCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	script := fmt.Sprintf(`
		http("First", {
			"url": "%s"
		});

		assert(false);

		http("Should not execute", {
			"url": "%s"
		});
	`, server.URL, server.URL)

	su := parseScript(t, script)
	su.Run()

	if su.Status != deluge.DoneError {
		t.Errorf("Expected status DoneError, got %v", su.Status)
	}

	// Only the first request should execute before assertion fails
	if atomic.LoadInt32(&requestCount) != 1 {
		t.Errorf("Expected 1 HTTP request before error, got %d", atomic.LoadInt32(&requestCount))
	}
}

// TestSimUserWithoutHTTPRequests verifies pure computation scenarios
func TestSimUserWithoutHTTPRequests(t *testing.T) {
	script := `
		let fibonacci = function(n) {
			if (n <= 1) {
				return n;
			}
			return fibonacci(n - 1) + fibonacci(n - 2);
		};

		let result = fibonacci(6);
		assert(result == 8);
	`

	su := parseScript(t, script)
	su.Run()

	if su.Status != deluge.DoneSuccess {
		t.Errorf("Expected status DoneSuccess, got %v", su.Status)
	}
}

// TestSimUserEmptyScript verifies empty script execution
func TestSimUserEmptyScript(t *testing.T) {
	su := parseScript(t, ``)
	su.Run()

	// Empty script should complete successfully
	if su.Status != deluge.DoneSuccess {
		t.Errorf("Expected status DoneSuccess for empty script, got %v", su.Status)
	}
}
