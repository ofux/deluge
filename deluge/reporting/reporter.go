package reporting

import (
	hdr "github.com/codahale/hdrhistogram"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/ofux/deluge/deluge/recording"
)

type Reporter interface {
	Report(recording.Recorder) (interface{}, error)
}

type Report struct {
	Name                 string
	Duration             duration.Duration
	ConcurrentUsersCount int
}

type Stats struct {
	CallCount              int64
	MinTime                int64
	MaxTime                int64
	MeanTime               float64
	ValueAtQuantiles       map[int]int64
	Distribution           []hdr.Bar
	CumulativeDistribution []hdr.Bracket
}

func newStatsFromHistogram(histo *hdr.Histogram) *Stats {
	stats := &Stats{
		CallCount:        histo.TotalCount(),
		MinTime:          histo.Min(),
		MaxTime:          histo.Max(),
		MeanTime:         histo.Mean(),
		ValueAtQuantiles: make(map[int]int64),
		//Distribution:           histo.Distribution(),
		CumulativeDistribution: histo.CumulativeDistribution(),
	}
	stats.ValueAtQuantiles[50] = histo.ValueAtQuantile(50)
	stats.ValueAtQuantiles[75] = histo.ValueAtQuantile(75)
	stats.ValueAtQuantiles[90] = histo.ValueAtQuantile(90)
	stats.ValueAtQuantiles[95] = histo.ValueAtQuantile(95)
	stats.ValueAtQuantiles[99] = histo.ValueAtQuantile(99)
	return stats
}
