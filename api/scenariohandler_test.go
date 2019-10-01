package api

import (
	"github.com/ofux/deluge/repov2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestScenarioHandler_Create(t *testing.T) {
	const scenarioKey = "myScenario"
	const scenarioName = "My scenario"

	var router = NewRouter(NewScenarioHandler())

	t.Run("Create a valid scenario", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		w := httptest.NewRecorder()

		const script = `scenario("` + scenarioKey + `", "` + scenarioName + `", function () { });`

		r := httptest.NewRequest("POST", "http://example.com/v1/scenarios", strings.NewReader(script))
		router.ServeHTTP(w, r)

		assert.Equal(t, w.Code, http.StatusCreated)
		scenario, ok := repov2.Instance.GetScenario(scenarioKey)
		require.True(t, ok)
		require.NotNil(t, scenario)
		assert.Equal(t, scenario.ID, scenarioKey)
		assert.Equal(t, scenario.Name, scenarioName)
		assert.Equal(t, scenario.Script, script)
	})

	t.Run("Create an invalid scenario", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		w := httptest.NewRecorder()

		// missing scenario name
		const script = `scenario("` + scenarioKey + `", function () { });`

		r := httptest.NewRequest("POST", "http://example.com/v1/scenarios", strings.NewReader(script))
		router.ServeHTTP(w, r)

		assert.Equal(t, w.Code, http.StatusBadRequest)
		_, ok := repov2.Instance.GetScenario(scenarioKey)
		require.False(t, ok)
	})

	t.Run("Create an existing scenario", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		w := httptest.NewRecorder()

		const script = `scenario("` + scenarioKey + `", "` + scenarioName + `", function () { });`
		err := repov2.Instance.SaveScenario(&repov2.PersistedScenario{
			ID:     scenarioKey,
			Name:   scenarioName,
			Script: script,
		})
		require.NoError(t, err)

		const script2 = `scenario("` + scenarioKey + `", "` + scenarioName + `", function () { let a=1; });`
		r := httptest.NewRequest("POST", "http://example.com/v1/scenarios", strings.NewReader(script2))
		router.ServeHTTP(w, r)

		assert.Equal(t, w.Code, http.StatusBadRequest)
		scenario, ok := repov2.Instance.GetScenario(scenarioKey)
		require.True(t, ok)
		require.NotNil(t, scenario)
		assert.Equal(t, scenario.ID, scenarioKey)
		assert.Equal(t, scenario.Name, scenarioName)
		assert.Equal(t, scenario.Script, script) // not script2
	})
}
