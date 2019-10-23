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
	"time"
)

func TestDelugeHandler_Create(t *testing.T) {
	const scenarioKey = "myScenario"
	const scenarioName = "My scenario"

	const delugeKey = "myDeluge"
	const delugeName = "My deluge"
	const delugeDuration = 200 * time.Millisecond

	var router = NewRouter(NewDelugeHandler())

	t.Run("Create a valid deluge", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		createScenario(t, scenarioKey, scenarioName)
		w := httptest.NewRecorder()

		var script = `deluge("` + delugeKey + `", "` + delugeName + `", "` + delugeDuration.String() + `", {
				"` + scenarioKey + `": {
					"concurrent": 10,
					"delay": "10ms"
				}
			});`

		r := httptest.NewRequest(http.MethodPost, "http://example.com/v1/deluges", strings.NewReader(script))
		router.ServeHTTP(w, r)

		assert.Equal(t, w.Code, http.StatusCreated)
		deluge, ok := repov2.Instance.GetDeluge(delugeKey)
		require.True(t, ok)
		require.NotNil(t, deluge)
		assert.Equal(t, deluge.ID, delugeKey)
		assert.Equal(t, deluge.Name, delugeName)
		assert.Equal(t, deluge.GlobalDuration, delugeDuration)
		assert.Equal(t, deluge.Script, script)
		assert.Len(t, deluge.ScenarioIDs, 1)
		assert.Contains(t, deluge.ScenarioIDs, scenarioKey)
	})

	t.Run("Create a valid deluge with undefined scenario", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		createScenario(t, scenarioKey, scenarioName)
		w := httptest.NewRecorder()

		var script = `deluge("` + delugeKey + `", "` + delugeName + `", "` + delugeDuration.String() + `", {
				"` + scenarioKey + `": {
					"concurrent": 10,
					"delay": "10ms"
				},
				"IDontExist": {
					"concurrent": 10,
					"delay": "10ms"
				}
			});`

		r := httptest.NewRequest(http.MethodPost, "http://example.com/v1/deluges", strings.NewReader(script))
		router.ServeHTTP(w, r)

		assert.Equal(t, w.Code, http.StatusCreated)
		deluge, ok := repov2.Instance.GetDeluge(delugeKey)
		require.True(t, ok)
		require.NotNil(t, deluge)
		assert.Equal(t, deluge.ID, delugeKey)
		assert.Equal(t, deluge.Name, delugeName)
		assert.Equal(t, deluge.GlobalDuration, delugeDuration)
		assert.Equal(t, deluge.Script, script)
		assert.Len(t, deluge.ScenarioIDs, 2)
		assert.Contains(t, deluge.ScenarioIDs, scenarioKey)
		assert.Contains(t, deluge.ScenarioIDs, "IDontExist")
	})

	t.Run("Create an invalid deluge", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		createScenario(t, scenarioKey, scenarioName)
		w := httptest.NewRecorder()

		// missing scenario
		var script = `deluge("` + delugeKey + `", "` + delugeName + `", "` + delugeDuration.String() + `");`

		r := httptest.NewRequest(http.MethodPost, "http://example.com/v1/deluges", strings.NewReader(script))
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		_, ok := repov2.Instance.GetDeluge(delugeKey)
		require.False(t, ok)
	})

	t.Run("Create an existing deluge", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		createScenario(t, scenarioKey, scenarioName)
		w := httptest.NewRecorder()

		script := createDeluge(t, delugeKey, delugeName, scenarioKey)

		var script2 = `deluge("` + delugeKey + `", "someOtherName", "3s", {
				"anything": {
					"concurrent": 10,
					"delay": "10ms"
				}
			});`
		r := httptest.NewRequest(http.MethodPost, "http://example.com/v1/deluges", strings.NewReader(script2))
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusConflict, w.Code)
		deluge, ok := repov2.Instance.GetDeluge(delugeKey)
		require.True(t, ok)
		require.NotNil(t, deluge)
		assert.Equal(t, deluge.ID, delugeKey)
		assert.Equal(t, deluge.Name, delugeName)
		assert.Equal(t, deluge.GlobalDuration, delugeDuration)
		assert.Equal(t, deluge.Script, script)
		assert.Len(t, deluge.ScenarioIDs, 1)
		assert.Contains(t, deluge.ScenarioIDs, scenarioKey)
	})

	t.Run("Fails to save deluge in repository", func(t *testing.T) {
		repov2.Instance = &repoMock{
			InMemoryRepository: *repov2.NewInMemoryRepository(),
			SaveDelugeImpl: func(scenario *repov2.PersistedDeluge) error {
				return errors.New("some error")
			},
		}
		createScenario(t, scenarioKey, scenarioName)
		w := httptest.NewRecorder()

		var script = `deluge("` + delugeKey + `", "` + delugeName + `", "` + delugeDuration.String() + `", {
				"` + scenarioKey + `": {
					"concurrent": 10,
					"delay": "10ms"
				}
			});`
		r := httptest.NewRequest(http.MethodPost, "http://example.com/v1/deluges", strings.NewReader(script))
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Empty body", func(t *testing.T) {
		repov2.Instance = &repoMock{
			InMemoryRepository: *repov2.NewInMemoryRepository(),
		}
		createScenario(t, scenarioKey, scenarioName)
		w := httptest.NewRecorder()

		r := httptest.NewRequest(http.MethodPost, "http://example.com/v1/deluges", nil)
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDelugeHandler_Update(t *testing.T) {
	const scenarioKey = "myScenario"
	const scenarioName = "My scenario"

	const delugeKey = "myDeluge"
	const delugeName = "My deluge"
	const delugeDuration = 200 * time.Millisecond

	var router = NewRouter(NewDelugeHandler())

	t.Run("Update an existing deluge", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		createScenario(t, scenarioKey, scenarioName)
		w := httptest.NewRecorder()

		script := createDeluge(t, delugeKey, delugeName, scenarioKey)

		var script2 = `deluge("` + delugeKey + `", "someOtherName", "3s", {
				"anything": {
					"concurrent": 10,
					"delay": "10ms"
				}
			});`
		r := httptest.NewRequest(http.MethodPut, "http://example.com/v1/deluges/"+delugeKey, strings.NewReader(script2))
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		deluge, ok := repov2.Instance.GetDeluge(delugeKey)
		require.True(t, ok)
		require.NotNil(t, deluge)
		assert.Equal(t, deluge.ID, delugeKey)
		assert.Equal(t, deluge.Name, "someOtherName")
		assert.Equal(t, deluge.GlobalDuration, 3*time.Second)
		assert.NotEqual(t, deluge.Script, script)
		assert.Equal(t, deluge.Script, script2)
		assert.Len(t, deluge.ScenarioIDs, 1)
		assert.Contains(t, deluge.ScenarioIDs, "anything")
	})

	t.Run("Update an invalid deluge", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		createScenario(t, scenarioKey, scenarioName)
		w := httptest.NewRecorder()

		script := createDeluge(t, delugeKey, delugeName, scenarioKey)

		var script2 = `deluge("` + delugeKey + `", "someOtherName", "3s");`
		r := httptest.NewRequest(http.MethodPut, "http://example.com/v1/deluges/"+delugeKey, strings.NewReader(script2))
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		deluge, ok := repov2.Instance.GetDeluge(delugeKey)
		require.True(t, ok)
		require.NotNil(t, deluge)
		assert.Equal(t, deluge.ID, delugeKey)
		assert.Equal(t, deluge.Name, delugeName)
		assert.Equal(t, deluge.GlobalDuration, delugeDuration)
		assert.Equal(t, deluge.Script, script)
		assert.Len(t, deluge.ScenarioIDs, 1)
		assert.Contains(t, deluge.ScenarioIDs, scenarioKey)
	})

	t.Run("Update a non-existing deluge", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		createScenario(t, scenarioKey, scenarioName)
		w := httptest.NewRecorder()

		var script = `deluge("` + delugeKey + `", "` + delugeName + `", "` + delugeDuration.String() + `", {
				"` + scenarioKey + `": {
					"concurrent": 10,
					"delay": "10ms"
				}
			});`

		r := httptest.NewRequest("PUT", "http://example.com/v1/deluges/badID", strings.NewReader(script))
		router.ServeHTTP(w, r)

		assert.Equal(t, w.Code, http.StatusNotFound)
		_, ok := repov2.Instance.GetDeluge(delugeKey)
		require.False(t, ok)
	})

	t.Run("Fails to save deluge in repository", func(t *testing.T) {
		rep := &repoMock{
			InMemoryRepository: *repov2.NewInMemoryRepository(),
		}
		repov2.Instance = rep
		createScenario(t, scenarioKey, scenarioName)
		w := httptest.NewRecorder()

		var script = createDeluge(t, delugeKey, delugeName, scenarioKey)

		rep.SaveDelugeImpl = func(scenario *repov2.PersistedDeluge) error {
			return errors.New("some error")
		}

		r := httptest.NewRequest("PUT", "http://example.com/v1/deluges/"+delugeKey, strings.NewReader(script))
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Empty body", func(t *testing.T) {
		repov2.Instance = &repoMock{
			InMemoryRepository: *repov2.NewInMemoryRepository(),
		}
		createScenario(t, scenarioKey, scenarioName)
		w := httptest.NewRecorder()

		createDeluge(t, delugeKey, delugeName, scenarioKey)

		r := httptest.NewRequest("PUT", "http://example.com/v1/deluges/"+delugeKey, nil)
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestDelugeHandler_GetByID(t *testing.T) {
	const scenarioKey = "myScenario"
	const scenarioName = "My scenario"

	const delugeKey = "myDeluge"
	const delugeName = "My deluge"

	var router = NewRouter(NewDelugeHandler())

	t.Run("Get an existing deluge", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		createScenario(t, scenarioKey, scenarioName)
		w := httptest.NewRecorder()

		var script = createDeluge(t, delugeKey, delugeName, scenarioKey)

		r := httptest.NewRequest(http.MethodGet, "http://example.com/v1/deluges/"+delugeKey, nil)
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		bbody, err := ioutil.ReadAll(w.Body)
		require.NoError(t, err)
		body := string(bbody)
		assert.Equal(t, script, body)
	})

	t.Run("Get a non-existing scenario", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		createScenario(t, scenarioKey, scenarioName)
		w := httptest.NewRecorder()

		r := httptest.NewRequest(http.MethodGet, "http://example.com/v1/deluges/"+delugeKey, nil)
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestDelugeHandler_DeleteByID(t *testing.T) {
	const scenarioKey = "myScenario"
	const scenarioName = "My scenario"

	const delugeKey = "myDeluge"
	const delugeName = "My deluge"

	var router = NewRouter(NewDelugeHandler())

	t.Run("Delete an existing deluge", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		createScenario(t, scenarioKey, scenarioName)
		w := httptest.NewRecorder()

		createDeluge(t, delugeKey, delugeName, scenarioKey)

		_, ok := repov2.Instance.GetDeluge(delugeKey)
		require.True(t, ok)

		r := httptest.NewRequest(http.MethodDelete, "http://example.com/v1/deluges/"+delugeKey, nil)
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		_, ok = repov2.Instance.GetDeluge(delugeKey)
		require.False(t, ok)
	})

	t.Run("Delete a non-existing deluge", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		createScenario(t, scenarioKey, scenarioName)
		w := httptest.NewRecorder()

		r := httptest.NewRequest(http.MethodDelete, "http://example.com/v1/deluges/"+delugeKey, nil)
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestDelugeHandler_GetAll(t *testing.T) {
	const scenarioKey = "myScenario"
	const scenarioName = "My scenario"

	const delugeKey1 = "myDeluge1"
	const delugeName1 = "My deluge 1"
	const delugeKey2 = "myDeluge2"
	const delugeName2 = "My deluge 2"

	var router = NewRouter(NewDelugeHandler())

	t.Run("Get all deluges", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		createScenario(t, scenarioKey, scenarioName)
		w := httptest.NewRecorder()

		createDeluge(t, delugeKey1, delugeName1, scenarioKey)
		createDeluge(t, delugeKey2, delugeName2, scenarioKey)

		r := httptest.NewRequest(http.MethodGet, "http://example.com/v1/deluges", nil)
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		bbody, err := ioutil.ReadAll(w.Body)
		require.NoError(t, err)
		body := string(bbody)
		assert.JSONEq(t, `{"elements":[
			{
				"id": "`+delugeKey1+`",
				"name": "`+delugeName1+`"
			},{
				"id": "`+delugeKey2+`",
				"name": "`+delugeName2+`"
			}
		]}`, body)
	})

	t.Run("Get all deluges on empty repository", func(t *testing.T) {
		repov2.Instance = repov2.NewInMemoryRepository()
		createScenario(t, scenarioKey, scenarioName)
		w := httptest.NewRecorder()

		r := httptest.NewRequest(http.MethodGet, "http://example.com/v1/deluges", nil)
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		bbody, err := ioutil.ReadAll(w.Body)
		require.NoError(t, err)
		body := string(bbody)
		assert.JSONEq(t, `{"elements":[]}`, body)
	})
}

func createDeluge(t *testing.T, delugeID, delugeName, scenarioID string) string {
	t.Helper()
	script := `deluge("` + delugeID + `", "` + delugeName + `", "200ms", {
		"` + scenarioID + `": {
			"concurrent": 10,
			"delay": "10ms"
		}
	});`
	err := repov2.Instance.SaveDeluge(&repov2.PersistedDeluge{
		ID:             delugeID,
		Name:           delugeName,
		Script:         script,
		GlobalDuration: 200 * time.Millisecond,
		ScenarioIDs:    []string{scenarioID},
	})
	require.NoError(t, err)
	return script
}

func (r *repoMock) SaveDeluge(deluge *repov2.PersistedDeluge) error {
	if r.SaveDelugeImpl == nil {
		return r.InMemoryRepository.SaveDeluge(deluge)
	}
	return r.SaveDelugeImpl(deluge)
}
