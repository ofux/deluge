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
	"testing"
)

func NewSimUserTest(t testing.TB, js string) *simUser {
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

	sc := &RunnableScenario{
		compiledScenario: &CompiledScenario{
			scenario: &ScenarioDefinition{
				ID:   "test-scenario",
				Name: "Test scenario",
			},
			script: script,
		},
		httpRecorder: recording.NewHTTPRecorder(1, 1),
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

func checkRecords(t *testing.T, rec *recording.HTTPRecorder, recName string, recCount int64) {
	rec.Close()
	records, err := rec.GetRecords()
	if err != nil {
		t.Fatal(err.Error())
	}
	record := records.OverTime[0]
	recordingtest.CheckHTTPRecord(t, record, recName, recCount, 200, recording.Ok)
}

func Benchmark_simUser_run(b *testing.B) {
	su := NewSimUserTest(b, `
		assert(1+1 == 2)
		`)

	for i := 0; i < b.N; i++ {
		su.run(0)
	}
}
