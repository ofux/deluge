package core

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCompileScenario(t *testing.T) {
	tests := []struct {
		input      string
		expectedID string
	}{
		{
			`
			scenario("myId", "My scenario", function () {})`,
			"myId",
		},
	}

	for _, tt := range tests {
		compiled, err := CompileScenario(tt.input)
		assert.NoError(t, err)
		assert.Equal(t, tt.expectedID, compiled.GetScenarioDefinition().ID)
	}
}

func TestCompileScenario_With_Scenario_Errors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			`
			scenario("My scenario", function () {})`,
			"RUNTIME ERROR: Expected 3 arguments at",
		},
		{
			`
			scenario(1, "My scenario", function () {})`,
			"RUNTIME ERROR: Expected 1st argument to be a string with at least 3 characters at",
		},
		{
			`
			scenario("myScenario", 1, function () {})`,
			"RUNTIME ERROR: Expected 2nd argument to be a string at",
		},
		{
			`
			scenario("myScenario", "My scenario", "bad")`,
			"RUNTIME ERROR: Expected 3rd argument to be a function at",
		},
		{
			`
			scenario("myScenario1", "My scenario 1", function () {});
			scenario("myScenario2", "My scenario 2", function () {});`,
			"RUNTIME ERROR: Expected only one deluge definition at",
		},
	}

	for _, tt := range tests {
		_, err := CompileScenario(tt.input)
		require.Error(t, err)
		assert.Contains(t, err.Error(), tt.expected)
	}
}

func BenchmarkCompileScenario(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := CompileScenario(`
scenario("myScenario", "My scenario", function () {

    http("My request", {
        "url": "http://localhost:8080/hello/toto"
    });

});`)

		if err != nil {
			b.Fatal(err)
		}
	}
}
