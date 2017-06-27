package reporting

import (
	hdr "github.com/codahale/hdrhistogram"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewStatsFromHistogram(t *testing.T) {

	t.Run("Get stats from simple histogram", func(t *testing.T) {
		histo := hdr.New(0, 100, 2)
		histo.RecordValue(1)
		histo.RecordValue(10)
		histo.RecordValue(50)
		histo.RecordValue(100)

		stats := newStatsFromHistogram(histo)

		assert.Equal(t, &Stats{
			CallCount: 4,
			MinTime:   1,
			MaxTime:   100,
			MeanTime:  40.25,
			ValueAtQuantiles: map[int]int64{
				50: 10,
				75: 50,
				90: 100,
				95: 100,
				99: 100,
			},
			CumulativeDistribution: []hdr.Bracket{
				{
					Quantile: 0,
					Count:    1,
					ValueAt:  1,
				},
				{
					Quantile: 50,
					Count:    2,
					ValueAt:  10,
				},
				{
					Quantile: 75,
					Count:    3,
					ValueAt:  50,
				},
				{
					Quantile: 87.5,
					Count:    4,
					ValueAt:  100,
				},
				{
					Quantile: 100,
					Count:    4,
					ValueAt:  100,
				},
			},
		}, stats)
	})

	t.Run("Get stats from empty histogram", func(t *testing.T) {
		histo := hdr.New(0, 100, 2)

		stats := newStatsFromHistogram(histo)

		assert.Equal(t, &Stats{
			CallCount: 0,
			MinTime:   0,
			MaxTime:   0,
			MeanTime:  0,
			ValueAtQuantiles: map[int]int64{
				50: 0,
				75: 0,
				90: 0,
				95: 0,
				99: 0,
			},
			CumulativeDistribution: []hdr.Bracket{
				{
					Quantile: 100,
					Count:    0,
					ValueAt:  0,
				},
			},
		}, stats)
	})
}
