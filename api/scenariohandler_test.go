package api

import (
	"errors"
	"github.com/ofux/deluge/repov2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
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

		script := createScenario(t, scenarioKey, scenarioName)

		const script2 = `scenario("` + scenarioKey + `", "` + scenarioName + `", function () { let a=1; });`
		r := httptest.NewRequest("POST", "http://example.com/v1/scenarios", strings.NewReader(script2))
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusConflict, w.Code)
		scenario, ok := repov2.Instance.GetScenario(scenarioKey)
		require.True(t, ok)
		require.NotNil(t, scenario)
		assert.Equal(t, scenario.ID, scenarioKey)
		assert.Equal(t, scenario.Name, scenarioName)
		assert.Equal(t, scenario.Script, script)
		assert.NotEqual(t, scenario.Script, script2)
	})

	t.Run("Fails to save scenario in repository", func(t *testing.T) {
		repov2.Instance = &repoMock{
			InMemoryRepository: *repov2.NewInMemoryRepository(),
			SaveScenarioImpl: func(scenario *repov2.PersistedScenario) error {
				return errors.New("some error")
			},
		}
		w := httptest.NewRecorder()

		const script = `scenario("` + scenarioKey + `", "` + scenarioName + `", function () { let a=1; });`
		r := httptest.NewRequest("POST", "http://example.com/v1/scenarios", strings.NewReader(script))
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Empty body", func(t *testing.T) {
		repov2.Instance = &repoMock{
			InMemoryRepository: *repov2.NewInMemoryRepository(),
		}
		w := httptest.NewRecorder()

		r := httptest.NewRequest("POST", "http://example.com/v1/scenarios", nil)
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestScenarioHandler_Update(t *testing.T) {
	const scenarioKey = "myScenario"
	const scenarioName = "My scenario"

	var router = NewRouter(NewScenarioHandler())
	t.Run("Update an existing scenario", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		w := httptest.NewRecorder()

		script := createScenario(t, scenarioKey, scenarioName)

		const script2 = `scenario("` + scenarioKey + `", "` + scenarioName + `", function () { let a=1; });`
		r := httptest.NewRequest("PUT", "http://example.com/v1/scenarios/"+scenarioKey, strings.NewReader(script2))
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		scenario, ok := repov2.Instance.GetScenario(scenarioKey)
		require.True(t, ok)
		require.NotNil(t, scenario)
		assert.Equal(t, scenario.ID, scenarioKey)
		assert.Equal(t, scenario.Name, scenarioName)
		assert.NotEqual(t, scenario.Script, script)
		assert.Equal(t, scenario.Script, script2)
	})

	t.Run("Update an invalid scenario", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		w := httptest.NewRecorder()

		script := createScenario(t, scenarioKey, scenarioName)

		// missing scenario name
		const invalidScript = `scenario("` + scenarioKey + `", function () { });`

		r := httptest.NewRequest("PUT", "http://example.com/v1/scenarios/"+scenarioKey, strings.NewReader(invalidScript))
		router.ServeHTTP(w, r)

		assert.Equal(t, w.Code, http.StatusBadRequest)
		scenario, ok := repov2.Instance.GetScenario(scenarioKey)
		require.True(t, ok)
		require.NotNil(t, scenario)
		assert.Equal(t, scenario.ID, scenarioKey)
		assert.Equal(t, scenario.Name, scenarioName)
		assert.Equal(t, scenario.Script, script) // not invalidScript
	})

	t.Run("Update a non-existing scenario", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		w := httptest.NewRecorder()

		const script = `scenario("` + scenarioKey + `", "` + scenarioName + `", function () { });`

		r := httptest.NewRequest("PUT", "http://example.com/v1/scenarios/badID", strings.NewReader(script))
		router.ServeHTTP(w, r)

		assert.Equal(t, w.Code, http.StatusNotFound)
		_, ok := repov2.Instance.GetScenario(scenarioKey)
		require.False(t, ok)
	})

	t.Run("Fails to save scenario in repository", func(t *testing.T) {
		rep := &repoMock{
			InMemoryRepository: *repov2.NewInMemoryRepository(),
		}
		repov2.Instance = rep
		w := httptest.NewRecorder()

		createScenario(t, scenarioKey, scenarioName)

		rep.SaveScenarioImpl = func(scenario *repov2.PersistedScenario) error {
			return errors.New("some error")
		}

		const script = `scenario("` + scenarioKey + `", "` + scenarioName + `", function () { let a=1; });`
		r := httptest.NewRequest("PUT", "http://example.com/v1/scenarios/"+scenarioKey, strings.NewReader(script))
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Empty body", func(t *testing.T) {
		repov2.Instance = &repoMock{
			InMemoryRepository: *repov2.NewInMemoryRepository(),
		}
		w := httptest.NewRecorder()

		createScenario(t, scenarioKey, scenarioName)

		r := httptest.NewRequest("PUT", "http://example.com/v1/scenarios/"+scenarioKey, nil)
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestScenarioHandler_GetByID(t *testing.T) {
	const scenarioKey = "myScenario"
	const scenarioName = "My scenario"

	var router = NewRouter(NewScenarioHandler())
	t.Run("Get an existing scenario", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		w := httptest.NewRecorder()

		createScenario(t, scenarioKey, scenarioName)

		r := httptest.NewRequest(http.MethodGet, "http://example.com/v1/scenarios/"+scenarioKey, nil)
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		bbody, err := ioutil.ReadAll(w.Body)
		require.NoError(t, err)
		body := string(bbody)
		assert.Equal(t, `scenario("myScenario", "My scenario", function () { });`, body)
	})

	t.Run("Get a non-existing scenario", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		w := httptest.NewRecorder()

		r := httptest.NewRequest(http.MethodGet, "http://example.com/v1/scenarios/"+scenarioKey, nil)
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestScenarioHandler_DeleteByID(t *testing.T) {
	const scenarioKey = "myScenario"
	const scenarioName = "My scenario"

	var router = NewRouter(NewScenarioHandler())
	t.Run("Delete an existing scenario", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		w := httptest.NewRecorder()

		createScenario(t, scenarioKey, scenarioName)

		_, ok := repov2.Instance.GetScenario(scenarioKey)
		require.True(t, ok)

		r := httptest.NewRequest(http.MethodDelete, "http://example.com/v1/scenarios/"+scenarioKey, nil)
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		_, ok = repov2.Instance.GetScenario(scenarioKey)
		require.False(t, ok)
	})

	t.Run("Delete a non-existing scenario", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		w := httptest.NewRecorder()

		r := httptest.NewRequest(http.MethodDelete, "http://example.com/v1/scenarios/"+scenarioKey, nil)
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestScenarioHandler_GetAll(t *testing.T) {
	const scenarioKey1 = "myScenario1"
	const scenarioName1 = "My scenario 1"
	const scenarioKey2 = "myScenario2"
	const scenarioName2 = "My scenario 2"

	var router = NewRouter(NewScenarioHandler())
	t.Run("Get all scenarios", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		w := httptest.NewRecorder()

		createScenario(t, scenarioKey1, scenarioName1)
		createScenario(t, scenarioKey2, scenarioName2)

		r := httptest.NewRequest(http.MethodGet, "http://example.com/v1/scenarios", nil)
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		bbody, err := ioutil.ReadAll(w.Body)
		require.NoError(t, err)
		body := string(bbody)
		assert.JSONEq(t, `{"elements":[
			{
				"id": "`+scenarioKey1+`",
				"name": "`+scenarioName1+`"
			},{
				"id": "`+scenarioKey2+`",
				"name": "`+scenarioName2+`"
			}
		]}`, body)
	})

	t.Run("Get all scenarios on empty repository", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		w := httptest.NewRecorder()

		r := httptest.NewRequest(http.MethodGet, "http://example.com/v1/scenarios", nil)
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		bbody, err := ioutil.ReadAll(w.Body)
		require.NoError(t, err)
		body := string(bbody)
		assert.JSONEq(t, `{"elements":[]}`, body)
	})
}

func createScenario(t *testing.T, scenarioID, scenarioName string) string {
	t.Helper()
	script := `scenario("` + scenarioID + `", "` + scenarioName + `", function () { });`
	err := repov2.Instance.SaveScenario(&repov2.PersistedScenario{
		ID:     scenarioID,
		Name:   scenarioName,
		Script: `scenario("` + scenarioID + `", "` + scenarioName + `", function () { });`,
	})
	require.NoError(t, err)
	return script
}

type repoMock struct {
	repov2.InMemoryRepository

	SaveScenarioImpl func(scenario *repov2.PersistedScenario) error
}

func (r *repoMock) SaveScenario(scenario *repov2.PersistedScenario) error {
	if r.SaveScenarioImpl == nil {
		return r.InMemoryRepository.SaveScenario(scenario)
	}
	return r.SaveScenarioImpl(scenario)
}
