package core

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCompileDeluge(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			`deluge("myID", "200ms", {
				"myScenario": {
					"concurrent": 100,
					"delay": "100ms"
				}
			});`,
			"RUNTIME ERROR: Expected 4 arguments at",
		},
		{
			`deluge(1, "Some name", "200ms", {
				"myScenario": {
					"concurrent": 100,
					"delay": "100ms"
				}
			});`,
			"RUNTIME ERROR: Expected 1st argument to be a string with at least 3 characters at",
		},
		{
			`deluge("myID", 1, "200ms", {
				"myScenario": {
					"concurrent": 100,
					"delay": "100ms"
				}
			});`,
			"RUNTIME ERROR: Expected 2nd argument to be a string at",
		},
		{
			`deluge("myID", "Some name", 200, {
				"myScenario": {
					"concurrent": 100,
					"delay": "100ms"
				}
			});`,
			"RUNTIME ERROR: Expected 3rd argument to be a string at",
		},
		{
			`deluge("myID", "Some name", "200", {
				"myScenario": {
					"concurrent": 100,
					"delay": "100ms"
				}
			});`,
			"RUNTIME ERROR: Expected 3rd argument to be a valid duration at",
		},
		{
			`deluge("myID", "Some name", "200ms", "bad");`,
			"RUNTIME ERROR: Expected 4th argument to be an object at",
		},
		{
			`deluge("myID", "Some name", "200ms", {
				"myScenario": "bad"
			});`,
			"RUNTIME ERROR: Expected scenario configuration to be an object at",
		},
		{
			`deluge("myID", "Some name", "200ms", {
				"myScenario": {
					"delay": "100ms"
				}
			});`,
			"RUNTIME ERROR: Expected 'concurrent' value in configuration at",
		},
		{
			`deluge("myID", "Some name", "200ms", {
				"myScenario": {
					"concurrent": 100
				}
			});`,
			"RUNTIME ERROR: Expected 'delay' value in configuration at",
		},
		{
			`deluge("myID", "Some name", "200ms", {
				"myScenario": {
					"concurrent": "100",
					"delay": "100ms"
				}
			});`,
			"RUNTIME ERROR: Expected 'concurrent' value to be an integer in configuration at",
		},
		{
			`deluge("myID", "Some name", "200ms", {
				"myScenario": {
					"concurrent": 100,
					"delay": 100
				}
			});`,
			"RUNTIME ERROR: Expected 'delay' value to be a valid duration in configuration at",
		},
		{
			`deluge("myID", "Some name", "200ms", {
				"myScenario": {
					"concurrent": 100,
					"delay": "100"
				}
			});`,
			"RUNTIME ERROR: Expected 'delay' value to be a valid duration in configuration at",
		},
		{
			`deluge("myID", "Some name", "200ms", {
				"myScenario": {
					"concurrent": 100,
					"delay": "100ms",
					"args": "foobar"
				}
			});`,
			"RUNTIME ERROR: Expected 'args' to be an object at",
		},
		{
			`deluge("myID", "Some name", "200ms", {}); deluge("Some other name", "200ms", {});`,
			"RUNTIME ERROR: Expected only one deluge definition at",
		},
	}

	for _, tt := range tests {
		clearScenarioRepo()
		compileScenario(t, `
scenario("myScenario", "My scenario", function () {

    http("My request", {
        "url": "http://localhost:8080/hello/toto"
    });

});`)
		_, err := CompileDeluge(tt.input)
		require.Error(t, err)
		assert.Contains(t, err.Error(), tt.expected)
	}
}
