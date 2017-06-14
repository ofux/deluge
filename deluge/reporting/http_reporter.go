package reporting

import (
	"github.com/ofux/deluge/deluge/recording"
)

type HTTPReporter struct{}

type HTTPReport struct {
	Stats *HTTPStatsOverTime
}

type HTTPStatsOverTime struct {
	Global       *HTTPStats
	PerIteration []*HTTPStats
}

type HTTPStats struct {
	HTTPRequestStats
	PerRequests map[string]*HTTPRequestStats
}

type HTTPRequestStats struct {
	Global    *Stats
	PerStatus map[int]*Stats
	PerOkKo   map[recording.OkKo]*Stats
}

func (r *HTTPReporter) Report(httpRecorder *recording.HTTPRecorder) (Report, error) {
	records, err := httpRecorder.GetRecords()
	if err != nil {
		return nil, err
	}

	report := &HTTPReport{
		Stats: &HTTPStatsOverTime{
			Global:       newHTTPStats(records.Global),
			PerIteration: make([]*HTTPStats, 0, 16),
		},
	}
	for _, v := range records.PerIteration {
		report.Stats.PerIteration = append(report.Stats.PerIteration, newHTTPStats(v))
	}

	return report, nil
}

func newHTTPStats(rec *recording.HTTPRecord) *HTTPStats {
	st := &HTTPStats{
		HTTPRequestStats: *newHTTPRequestStats(&(rec.HTTPRequestRecord)),
		PerRequests:      make(map[string]*HTTPRequestStats),
	}
	for k, v := range rec.PerRequests {
		st.PerRequests[k] = newHTTPRequestStats(v)
	}
	return st
}

func newHTTPRequestStats(rec *recording.HTTPRequestRecord) *HTTPRequestStats {
	st := &HTTPRequestStats{
		Global:    newStatsFromHistogram(rec.Global),
		PerStatus: make(map[int]*Stats),
		PerOkKo:   make(map[recording.OkKo]*Stats),
	}
	for k, v := range rec.PerStatus {
		st.PerStatus[k] = newStatsFromHistogram(v)
	}
	for k, v := range rec.PerOkKo {
		st.PerOkKo[k] = newStatsFromHistogram(v)
	}
	return st
}
