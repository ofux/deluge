package reporting

import (
	"github.com/ofux/deluge/core/recording"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHTTPReporter_Report(t *testing.T) {
	recorder := recording.NewHTTPRecorder(1)

	for i := 0; i < 3; i++ {
		recorder.Record(&recording.HTTPRecordEntry{
			Iteration:  i,
			Name:       "foo",
			Value:      1000,
			StatusCode: 200,
		})
		recorder.Record(&recording.HTTPRecordEntry{
			Iteration:  i,
			Name:       "foo",
			Value:      14000,
			StatusCode: 200,
		})
		recorder.Record(&recording.HTTPRecordEntry{
			Iteration:  i,
			Name:       "foo",
			Value:      100,
			StatusCode: 400,
		})
		recorder.Record(&recording.HTTPRecordEntry{
			Iteration:  i,
			Name:       "foo",
			Value:      6000,
			StatusCode: 500,
		})
		recorder.Record(&recording.HTTPRecordEntry{
			Iteration:  i,
			Name:       "bar",
			Value:      20000,
			StatusCode: 201,
		})
		recorder.Record(&recording.HTTPRecordEntry{
			Iteration:  i,
			Name:       "bar",
			Value:      40000,
			StatusCode: 500,
		})
	}

	recorder.Close()

	reporter := &HTTPReporter{}

	recs, err := recorder.GetRecords()
	require.NoError(t, err)

	report := reporter.Report(recs)
	rep := report.(*HTTPReport)
	assert.Equal(t, int64(18), rep.Stats.Global.Global.CallCount)
	assert.Equal(t, int64(100), rep.Stats.Global.Global.MinTime)
	assert.InDelta(t, int64(40000), rep.Stats.Global.Global.MaxTime, 100)

	assert.Len(t, rep.Stats.Global.PerRequests, 2)
	assert.Len(t, rep.Stats.Global.PerStatus, 4)
	assert.Len(t, rep.Stats.Global.PerOkKo, 2)
	assert.Equal(t, int64(12), rep.Stats.Global.PerRequests["foo"].Global.CallCount)
	assert.Equal(t, int64(6), rep.Stats.Global.PerRequests["foo"].PerOkKo[recording.Ok].CallCount)
	assert.Equal(t, int64(6), rep.Stats.Global.PerRequests["foo"].PerOkKo[recording.Ko].CallCount)
	assert.Len(t, rep.Stats.Global.PerRequests["foo"].PerStatus, 3)
	assert.Equal(t, int64(6), rep.Stats.Global.PerRequests["foo"].PerStatus[200].CallCount)
	assert.Equal(t, int64(3), rep.Stats.Global.PerRequests["foo"].PerStatus[400].CallCount)
	assert.Equal(t, int64(3), rep.Stats.Global.PerRequests["foo"].PerStatus[500].CallCount)

	assert.Equal(t, int64(6), rep.Stats.Global.PerRequests["bar"].Global.CallCount)
	assert.Equal(t, int64(3), rep.Stats.Global.PerRequests["bar"].PerOkKo[recording.Ok].CallCount)
	assert.Equal(t, int64(3), rep.Stats.Global.PerRequests["bar"].PerOkKo[recording.Ko].CallCount)
	assert.Len(t, rep.Stats.Global.PerRequests["bar"].PerStatus, 2)
	assert.Equal(t, int64(3), rep.Stats.Global.PerRequests["bar"].PerStatus[201].CallCount)
	assert.Equal(t, int64(3), rep.Stats.Global.PerRequests["bar"].PerStatus[500].CallCount)
}
