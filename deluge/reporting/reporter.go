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
	Distribution           []hdr.Bar
	CumulativeDistribution []hdr.Bracket
}

func newStatsFromHistogram(histo *hdr.Histogram) *Stats {
	return &Stats{
		CallCount:              histo.TotalCount(),
		MinTime:                histo.Min(),
		MaxTime:                histo.Max(),
		MeanTime:               histo.Mean(),
		Distribution:           histo.Distribution(),
		CumulativeDistribution: histo.CumulativeDistribution(),
	}
}
