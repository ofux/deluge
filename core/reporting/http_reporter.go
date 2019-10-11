package reporting

import (
	"github.com/ofux/deluge/core/recording"
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

func (r *HTTPReporter) Report(records *recording.HTTPRecordsOverTime) Report {
	report := &HTTPReport{
		Stats: &HTTPStatsOverTime{
			Global:       newHTTPStats(records.Global),
			PerIteration: make([]*HTTPStats, 0, 16),
		},
	}
	for _, v := range records.OverTime {
		report.Stats.PerIteration = append(report.Stats.PerIteration, newHTTPStats(v))
	}

	return report
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
