package recording

import (
	"github.com/codahale/hdrhistogram"
	hdr "github.com/codahale/hdrhistogram"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestMergeHTTPRecordsOverTime(t *testing.T) {
	var someRecords = &HTTPRecordsOverTime{}

	type args struct {
		rec1 *HTTPRecordsOverTime
		rec2 *HTTPRecordsOverTime
	}
	tests := []struct {
		name string
		args args
		want *HTTPRecordsOverTime
	}{
		{
			name: "First record is nil",
			args: args{
				rec1: nil,
				rec2: someRecords,
			},
			want: someRecords,
		}, {
			name: "Second record is nil",
			args: args{
				rec1: someRecords,
				rec2: nil,
			},
			want: someRecords,
		}, {
			name: "Two complete histograms",
			args: args{
				rec1: &HTTPRecordsOverTime{
					Global: &HTTPRecord{
						HTTPRequestRecord: HTTPRequestRecord{
							Global: newFakeHistogram(t, 200, 300),
							PerStatus: map[int]*hdrhistogram.Histogram{
								200: newFakeHistogram(t, 200, 300),
								404: newFakeHistogram(t, 200, 400),
							},
							PerOkKo: map[OkKo]*hdrhistogram.Histogram{
								Ok: newFakeHistogram(t, 200, 300),
								Ko: newFakeHistogram(t, 200, 400),
							},
						},
						PerRequests: map[string]*HTTPRequestRecord{
							"req1": &HTTPRequestRecord{
								Global: newFakeHistogram(t, 200, 300),
								PerStatus: map[int]*hdrhistogram.Histogram{
									200: newFakeHistogram(t, 200, 300),
								},
								PerOkKo: map[OkKo]*hdrhistogram.Histogram{
									Ok: newFakeHistogram(t, 200, 300),
								},
							},
							"req2": &HTTPRequestRecord{
								Global: newFakeHistogram(t, 200, 300),
								PerStatus: map[int]*hdrhistogram.Histogram{
									200: newFakeHistogram(t, 200, 300),
									404: newFakeHistogram(t, 200, 400),
								},
								PerOkKo: map[OkKo]*hdrhistogram.Histogram{
									Ok: newFakeHistogram(t, 200, 300),
									Ko: newFakeHistogram(t, 200, 400),
								},
							},
						},
					},
					PerIteration: []*HTTPRecord{
						&HTTPRecord{
							HTTPRequestRecord: HTTPRequestRecord{
								Global: newFakeHistogram(t, 200, 300),
								PerStatus: map[int]*hdrhistogram.Histogram{
									200: newFakeHistogram(t, 200, 300),
								},
								PerOkKo: map[OkKo]*hdrhistogram.Histogram{
									Ok: newFakeHistogram(t, 200, 300),
								},
							},
							PerRequests: map[string]*HTTPRequestRecord{
								"req1": &HTTPRequestRecord{
									Global: newFakeHistogram(t, 200, 300),
									PerStatus: map[int]*hdrhistogram.Histogram{
										200: newFakeHistogram(t, 200, 300),
									},
									PerOkKo: map[OkKo]*hdrhistogram.Histogram{
										Ok: newFakeHistogram(t, 200, 300),
									},
								},
								"req2": &HTTPRequestRecord{
									Global: newFakeHistogram(t, 200, 300),
									PerStatus: map[int]*hdrhistogram.Histogram{
										200: newFakeHistogram(t, 200, 300),
									},
									PerOkKo: map[OkKo]*hdrhistogram.Histogram{
										Ok: newFakeHistogram(t, 200, 300),
									},
								},
							},
						}, &HTTPRecord{
							HTTPRequestRecord: HTTPRequestRecord{
								Global: newFakeHistogram(t, 200, 300),
								PerStatus: map[int]*hdrhistogram.Histogram{
									200: newFakeHistogram(t, 200, 300),
									404: newFakeHistogram(t, 200, 400),
								},
								PerOkKo: map[OkKo]*hdrhistogram.Histogram{
									Ok: newFakeHistogram(t, 200, 300),
									Ko: newFakeHistogram(t, 200, 400),
								},
							},
							PerRequests: map[string]*HTTPRequestRecord{
								"req1": &HTTPRequestRecord{
									Global:    newFakeHistogram(t, 200, 300),
									PerStatus: map[int]*hdrhistogram.Histogram{},
									PerOkKo:   map[OkKo]*hdrhistogram.Histogram{},
								},
								"req2": &HTTPRequestRecord{
									Global: newFakeHistogram(t, 200, 300),
									PerStatus: map[int]*hdrhistogram.Histogram{
										404: newFakeHistogram(t, 200, 400),
									},
									PerOkKo: map[OkKo]*hdrhistogram.Histogram{
										Ko: newFakeHistogram(t, 200, 400),
									},
								},
							},
						},
					},
				},
				rec2: &HTTPRecordsOverTime{
					Global: &HTTPRecord{
						HTTPRequestRecord: HTTPRequestRecord{
							Global: newFakeHistogram(t, 500, 600),
							PerStatus: map[int]*hdrhistogram.Histogram{
								200: newFakeHistogram(t, 500, 600),
								404: newFakeHistogram(t, 200, 400),
							},
							PerOkKo: map[OkKo]*hdrhistogram.Histogram{
								Ok: newFakeHistogram(t, 500, 600),
								Ko: newFakeHistogram(t, 200, 400),
							},
						},
						PerRequests: map[string]*HTTPRequestRecord{
							"req1": &HTTPRequestRecord{
								Global: newFakeHistogram(t, 500, 600),
								PerStatus: map[int]*hdrhistogram.Histogram{
									200: newFakeHistogram(t, 500, 600),
								},
								PerOkKo: map[OkKo]*hdrhistogram.Histogram{
									Ok: newFakeHistogram(t, 500, 600),
								},
							},
							"req3": &HTTPRequestRecord{
								Global: newFakeHistogram(t, 500, 600),
								PerStatus: map[int]*hdrhistogram.Histogram{
									200: newFakeHistogram(t, 500, 600),
									404: newFakeHistogram(t, 200, 400),
								},
								PerOkKo: map[OkKo]*hdrhistogram.Histogram{
									Ok: newFakeHistogram(t, 500, 600),
									Ko: newFakeHistogram(t, 200, 400),
								},
							},
						},
					},
					PerIteration: []*HTTPRecord{
						&HTTPRecord{
							HTTPRequestRecord: HTTPRequestRecord{
								Global: newFakeHistogram(t, 500, 600),
								PerStatus: map[int]*hdrhistogram.Histogram{
									200: newFakeHistogram(t, 500, 600),
								},
								PerOkKo: map[OkKo]*hdrhistogram.Histogram{
									Ok: newFakeHistogram(t, 500, 600),
								},
							},
							PerRequests: map[string]*HTTPRequestRecord{
								"req1": &HTTPRequestRecord{
									Global: newFakeHistogram(t, 500, 600),
									PerStatus: map[int]*hdrhistogram.Histogram{
										200: newFakeHistogram(t, 500, 600),
									},
									PerOkKo: map[OkKo]*hdrhistogram.Histogram{
										Ok: newFakeHistogram(t, 500, 600),
									},
								},
								"req3": &HTTPRequestRecord{
									Global: newFakeHistogram(t, 500, 600),
									PerStatus: map[int]*hdrhistogram.Histogram{
										200: newFakeHistogram(t, 500, 600),
									},
									PerOkKo: map[OkKo]*hdrhistogram.Histogram{
										Ok: newFakeHistogram(t, 500, 600),
									},
								},
							},
						}, &HTTPRecord{
							HTTPRequestRecord: HTTPRequestRecord{
								Global: newFakeHistogram(t, 500, 600),
								PerStatus: map[int]*hdrhistogram.Histogram{
									200: newFakeHistogram(t, 500, 600),
									404: newFakeHistogram(t, 200, 400),
								},
								PerOkKo: map[OkKo]*hdrhistogram.Histogram{
									Ok: newFakeHistogram(t, 500, 600),
									Ko: newFakeHistogram(t, 200, 400),
								},
							},
							PerRequests: map[string]*HTTPRequestRecord{
								"req1": &HTTPRequestRecord{
									Global:    newFakeHistogram(t, 500, 600),
									PerStatus: map[int]*hdrhistogram.Histogram{},
									PerOkKo:   map[OkKo]*hdrhistogram.Histogram{},
								},
								"req3": &HTTPRequestRecord{
									Global: newFakeHistogram(t, 500, 600),
									PerStatus: map[int]*hdrhistogram.Histogram{
										404: newFakeHistogram(t, 200, 400),
									},
									PerOkKo: map[OkKo]*hdrhistogram.Histogram{
										Ko: newFakeHistogram(t, 200, 400),
									},
								},
							},
						},
					},
				},
			},

			want: &HTTPRecordsOverTime{
				Global: &HTTPRecord{
					HTTPRequestRecord: HTTPRequestRecord{
						Global: newFakeHistogram(t, 200, 300, 500, 600),
						PerStatus: map[int]*hdrhistogram.Histogram{
							200: newFakeHistogram(t, 200, 300, 500, 600),
							404: newFakeHistogram(t, 200, 400, 200, 400),
						},
						PerOkKo: map[OkKo]*hdrhistogram.Histogram{
							Ok: newFakeHistogram(t, 200, 300, 500, 600),
							Ko: newFakeHistogram(t, 200, 400, 200, 400),
						},
					},
					PerRequests: map[string]*HTTPRequestRecord{
						"req1": &HTTPRequestRecord{
							Global: newFakeHistogram(t, 200, 300, 500, 600),
							PerStatus: map[int]*hdrhistogram.Histogram{
								200: newFakeHistogram(t, 200, 300, 500, 600),
							},
							PerOkKo: map[OkKo]*hdrhistogram.Histogram{
								Ok: newFakeHistogram(t, 200, 300, 500, 600),
							},
						},
						"req2": &HTTPRequestRecord{
							Global: newFakeHistogram(t, 200, 300),
							PerStatus: map[int]*hdrhistogram.Histogram{
								200: newFakeHistogram(t, 200, 300),
								404: newFakeHistogram(t, 200, 400),
							},
							PerOkKo: map[OkKo]*hdrhistogram.Histogram{
								Ok: newFakeHistogram(t, 200, 300),
								Ko: newFakeHistogram(t, 200, 400),
							},
						},
						"req3": &HTTPRequestRecord{
							Global: newFakeHistogram(t, 500, 600),
							PerStatus: map[int]*hdrhistogram.Histogram{
								200: newFakeHistogram(t, 500, 600),
								404: newFakeHistogram(t, 200, 400),
							},
							PerOkKo: map[OkKo]*hdrhistogram.Histogram{
								Ok: newFakeHistogram(t, 500, 600),
								Ko: newFakeHistogram(t, 200, 400),
							},
						},
					},
				},
				PerIteration: []*HTTPRecord{
					&HTTPRecord{
						HTTPRequestRecord: HTTPRequestRecord{
							Global: newFakeHistogram(t, 200, 300, 500, 600),
							PerStatus: map[int]*hdrhistogram.Histogram{
								200: newFakeHistogram(t, 200, 300, 500, 600),
							},
							PerOkKo: map[OkKo]*hdrhistogram.Histogram{
								Ok: newFakeHistogram(t, 200, 300, 500, 600),
							},
						},
						PerRequests: map[string]*HTTPRequestRecord{
							"req1": &HTTPRequestRecord{
								Global: newFakeHistogram(t, 200, 300, 500, 600),
								PerStatus: map[int]*hdrhistogram.Histogram{
									200: newFakeHistogram(t, 200, 300, 500, 600),
								},
								PerOkKo: map[OkKo]*hdrhistogram.Histogram{
									Ok: newFakeHistogram(t, 200, 300, 500, 600),
								},
							},
							"req2": &HTTPRequestRecord{
								Global: newFakeHistogram(t, 200, 300),
								PerStatus: map[int]*hdrhistogram.Histogram{
									200: newFakeHistogram(t, 200, 300),
								},
								PerOkKo: map[OkKo]*hdrhistogram.Histogram{
									Ok: newFakeHistogram(t, 200, 300),
								},
							},
							"req3": &HTTPRequestRecord{
								Global: newFakeHistogram(t, 500, 600),
								PerStatus: map[int]*hdrhistogram.Histogram{
									200: newFakeHistogram(t, 500, 600),
								},
								PerOkKo: map[OkKo]*hdrhistogram.Histogram{
									Ok: newFakeHistogram(t, 500, 600),
								},
							},
						},
					}, &HTTPRecord{
						HTTPRequestRecord: HTTPRequestRecord{
							Global: newFakeHistogram(t, 200, 300, 500, 600),
							PerStatus: map[int]*hdrhistogram.Histogram{
								200: newFakeHistogram(t, 200, 300, 500, 600),
								404: newFakeHistogram(t, 200, 400, 200, 400),
							},
							PerOkKo: map[OkKo]*hdrhistogram.Histogram{
								Ok: newFakeHistogram(t, 200, 300, 500, 600),
								Ko: newFakeHistogram(t, 200, 400, 200, 400),
							},
						},
						PerRequests: map[string]*HTTPRequestRecord{
							"req1": &HTTPRequestRecord{
								Global:    newFakeHistogram(t, 200, 300, 500, 600),
								PerStatus: map[int]*hdrhistogram.Histogram{},
								PerOkKo:   map[OkKo]*hdrhistogram.Histogram{},
							},
							"req2": &HTTPRequestRecord{
								Global: newFakeHistogram(t, 200, 300),
								PerStatus: map[int]*hdrhistogram.Histogram{
									404: newFakeHistogram(t, 200, 400),
								},
								PerOkKo: map[OkKo]*hdrhistogram.Histogram{
									Ko: newFakeHistogram(t, 200, 400),
								},
							},
							"req3": &HTTPRequestRecord{
								Global: newFakeHistogram(t, 500, 600),
								PerStatus: map[int]*hdrhistogram.Histogram{
									404: newFakeHistogram(t, 200, 400),
								},
								PerOkKo: map[OkKo]*hdrhistogram.Histogram{
									Ko: newFakeHistogram(t, 200, 400),
								},
							},
						},
					},
				},
			},
		},

		{
			name: "Two complete histograms with different iteration count",
			args: args{
				rec1: &HTTPRecordsOverTime{
					Global: &HTTPRecord{
						HTTPRequestRecord: HTTPRequestRecord{
							Global: newFakeHistogram(t, 200, 300),
							PerStatus: map[int]*hdrhistogram.Histogram{
								200: newFakeHistogram(t, 200, 300),
								404: newFakeHistogram(t, 200, 400),
							},
							PerOkKo: map[OkKo]*hdrhistogram.Histogram{
								Ok: newFakeHistogram(t, 200, 300),
								Ko: newFakeHistogram(t, 200, 400),
							},
						},
						PerRequests: map[string]*HTTPRequestRecord{
							"req1": &HTTPRequestRecord{
								Global: newFakeHistogram(t, 200, 300),
								PerStatus: map[int]*hdrhistogram.Histogram{
									200: newFakeHistogram(t, 200, 300),
								},
								PerOkKo: map[OkKo]*hdrhistogram.Histogram{
									Ok: newFakeHistogram(t, 200, 300),
								},
							},
							"req2": &HTTPRequestRecord{
								Global: newFakeHistogram(t, 200, 300),
								PerStatus: map[int]*hdrhistogram.Histogram{
									200: newFakeHistogram(t, 200, 300),
									404: newFakeHistogram(t, 200, 400),
								},
								PerOkKo: map[OkKo]*hdrhistogram.Histogram{
									Ok: newFakeHistogram(t, 200, 300),
									Ko: newFakeHistogram(t, 200, 400),
								},
							},
						},
					},
					PerIteration: []*HTTPRecord{
						&HTTPRecord{
							HTTPRequestRecord: HTTPRequestRecord{
								Global: newFakeHistogram(t, 200, 300),
								PerStatus: map[int]*hdrhistogram.Histogram{
									200: newFakeHistogram(t, 200, 300),
								},
								PerOkKo: map[OkKo]*hdrhistogram.Histogram{
									Ok: newFakeHistogram(t, 200, 300),
								},
							},
							PerRequests: map[string]*HTTPRequestRecord{
								"req1": &HTTPRequestRecord{
									Global: newFakeHistogram(t, 200, 300),
									PerStatus: map[int]*hdrhistogram.Histogram{
										200: newFakeHistogram(t, 200, 300),
									},
									PerOkKo: map[OkKo]*hdrhistogram.Histogram{
										Ok: newFakeHistogram(t, 200, 300),
									},
								},
								"req2": &HTTPRequestRecord{
									Global: newFakeHistogram(t, 200, 300),
									PerStatus: map[int]*hdrhistogram.Histogram{
										200: newFakeHistogram(t, 200, 300),
									},
									PerOkKo: map[OkKo]*hdrhistogram.Histogram{
										Ok: newFakeHistogram(t, 200, 300),
									},
								},
							},
						}, &HTTPRecord{
							HTTPRequestRecord: HTTPRequestRecord{
								Global: newFakeHistogram(t, 200, 300),
								PerStatus: map[int]*hdrhistogram.Histogram{
									200: newFakeHistogram(t, 200, 300),
									404: newFakeHistogram(t, 200, 400),
								},
								PerOkKo: map[OkKo]*hdrhistogram.Histogram{
									Ok: newFakeHistogram(t, 200, 300),
									Ko: newFakeHistogram(t, 200, 400),
								},
							},
							PerRequests: map[string]*HTTPRequestRecord{
								"req1": &HTTPRequestRecord{
									Global:    newFakeHistogram(t, 200, 300),
									PerStatus: map[int]*hdrhistogram.Histogram{},
									PerOkKo:   map[OkKo]*hdrhistogram.Histogram{},
								},
								"req2": &HTTPRequestRecord{
									Global: newFakeHistogram(t, 200, 300),
									PerStatus: map[int]*hdrhistogram.Histogram{
										404: newFakeHistogram(t, 200, 400),
									},
									PerOkKo: map[OkKo]*hdrhistogram.Histogram{
										Ko: newFakeHistogram(t, 200, 400),
									},
								},
							},
						},
					},
				},
				rec2: &HTTPRecordsOverTime{
					Global: &HTTPRecord{
						HTTPRequestRecord: HTTPRequestRecord{
							Global: newFakeHistogram(t, 500, 600),
							PerStatus: map[int]*hdrhistogram.Histogram{
								200: newFakeHistogram(t, 500, 600),
								404: newFakeHistogram(t, 200, 400),
							},
							PerOkKo: map[OkKo]*hdrhistogram.Histogram{
								Ok: newFakeHistogram(t, 500, 600),
								Ko: newFakeHistogram(t, 200, 400),
							},
						},
						PerRequests: map[string]*HTTPRequestRecord{
							"req1": &HTTPRequestRecord{
								Global: newFakeHistogram(t, 500, 600),
								PerStatus: map[int]*hdrhistogram.Histogram{
									200: newFakeHistogram(t, 500, 600),
								},
								PerOkKo: map[OkKo]*hdrhistogram.Histogram{
									Ok: newFakeHistogram(t, 500, 600),
								},
							},
							"req3": &HTTPRequestRecord{
								Global: newFakeHistogram(t, 500, 600),
								PerStatus: map[int]*hdrhistogram.Histogram{
									200: newFakeHistogram(t, 500, 600),
									404: newFakeHistogram(t, 200, 400),
								},
								PerOkKo: map[OkKo]*hdrhistogram.Histogram{
									Ok: newFakeHistogram(t, 500, 600),
									Ko: newFakeHistogram(t, 200, 400),
								},
							},
						},
					},
					PerIteration: []*HTTPRecord{
						&HTTPRecord{
							HTTPRequestRecord: HTTPRequestRecord{
								Global: newFakeHistogram(t, 500, 600),
								PerStatus: map[int]*hdrhistogram.Histogram{
									200: newFakeHistogram(t, 500, 600),
								},
								PerOkKo: map[OkKo]*hdrhistogram.Histogram{
									Ok: newFakeHistogram(t, 500, 600),
								},
							},
							PerRequests: map[string]*HTTPRequestRecord{
								"req1": &HTTPRequestRecord{
									Global: newFakeHistogram(t, 500, 600),
									PerStatus: map[int]*hdrhistogram.Histogram{
										200: newFakeHistogram(t, 500, 600),
									},
									PerOkKo: map[OkKo]*hdrhistogram.Histogram{
										Ok: newFakeHistogram(t, 500, 600),
									},
								},
								"req3": &HTTPRequestRecord{
									Global: newFakeHistogram(t, 500, 600),
									PerStatus: map[int]*hdrhistogram.Histogram{
										200: newFakeHistogram(t, 500, 600),
									},
									PerOkKo: map[OkKo]*hdrhistogram.Histogram{
										Ok: newFakeHistogram(t, 500, 600),
									},
								},
							},
						}, &HTTPRecord{
							HTTPRequestRecord: HTTPRequestRecord{
								Global: newFakeHistogram(t, 500, 600),
								PerStatus: map[int]*hdrhistogram.Histogram{
									200: newFakeHistogram(t, 500, 600),
									404: newFakeHistogram(t, 200, 400),
								},
								PerOkKo: map[OkKo]*hdrhistogram.Histogram{
									Ok: newFakeHistogram(t, 500, 600),
									Ko: newFakeHistogram(t, 200, 400),
								},
							},
							PerRequests: map[string]*HTTPRequestRecord{
								"req1": &HTTPRequestRecord{
									Global:    newFakeHistogram(t, 500, 600),
									PerStatus: map[int]*hdrhistogram.Histogram{},
									PerOkKo:   map[OkKo]*hdrhistogram.Histogram{},
								},
								"req3": &HTTPRequestRecord{
									Global: newFakeHistogram(t, 500, 600),
									PerStatus: map[int]*hdrhistogram.Histogram{
										404: newFakeHistogram(t, 200, 400),
									},
									PerOkKo: map[OkKo]*hdrhistogram.Histogram{
										Ko: newFakeHistogram(t, 200, 400),
									},
								},
							},
						}, &HTTPRecord{
							HTTPRequestRecord: HTTPRequestRecord{
								Global: newFakeHistogram(t, 500, 600),
								PerStatus: map[int]*hdrhistogram.Histogram{
									200: newFakeHistogram(t, 500, 600),
									404: newFakeHistogram(t, 200, 400),
								},
								PerOkKo: map[OkKo]*hdrhistogram.Histogram{
									Ok: newFakeHistogram(t, 500, 600),
									Ko: newFakeHistogram(t, 200, 400),
								},
							},
							PerRequests: map[string]*HTTPRequestRecord{
								"req1": &HTTPRequestRecord{
									Global:    newFakeHistogram(t, 500, 600),
									PerStatus: map[int]*hdrhistogram.Histogram{},
									PerOkKo:   map[OkKo]*hdrhistogram.Histogram{},
								},
								"req3": &HTTPRequestRecord{
									Global: newFakeHistogram(t, 500, 600),
									PerStatus: map[int]*hdrhistogram.Histogram{
										404: newFakeHistogram(t, 200, 400),
									},
									PerOkKo: map[OkKo]*hdrhistogram.Histogram{
										Ko: newFakeHistogram(t, 200, 400),
									},
								},
							},
						},
					},
				},
			},

			want: &HTTPRecordsOverTime{
				Global: &HTTPRecord{
					HTTPRequestRecord: HTTPRequestRecord{
						Global: newFakeHistogram(t, 200, 300, 500, 600),
						PerStatus: map[int]*hdrhistogram.Histogram{
							200: newFakeHistogram(t, 200, 300, 500, 600),
							404: newFakeHistogram(t, 200, 400, 200, 400),
						},
						PerOkKo: map[OkKo]*hdrhistogram.Histogram{
							Ok: newFakeHistogram(t, 200, 300, 500, 600),
							Ko: newFakeHistogram(t, 200, 400, 200, 400),
						},
					},
					PerRequests: map[string]*HTTPRequestRecord{
						"req1": &HTTPRequestRecord{
							Global: newFakeHistogram(t, 200, 300, 500, 600),
							PerStatus: map[int]*hdrhistogram.Histogram{
								200: newFakeHistogram(t, 200, 300, 500, 600),
							},
							PerOkKo: map[OkKo]*hdrhistogram.Histogram{
								Ok: newFakeHistogram(t, 200, 300, 500, 600),
							},
						},
						"req2": &HTTPRequestRecord{
							Global: newFakeHistogram(t, 200, 300),
							PerStatus: map[int]*hdrhistogram.Histogram{
								200: newFakeHistogram(t, 200, 300),
								404: newFakeHistogram(t, 200, 400),
							},
							PerOkKo: map[OkKo]*hdrhistogram.Histogram{
								Ok: newFakeHistogram(t, 200, 300),
								Ko: newFakeHistogram(t, 200, 400),
							},
						},
						"req3": &HTTPRequestRecord{
							Global: newFakeHistogram(t, 500, 600),
							PerStatus: map[int]*hdrhistogram.Histogram{
								200: newFakeHistogram(t, 500, 600),
								404: newFakeHistogram(t, 200, 400),
							},
							PerOkKo: map[OkKo]*hdrhistogram.Histogram{
								Ok: newFakeHistogram(t, 500, 600),
								Ko: newFakeHistogram(t, 200, 400),
							},
						},
					},
				},
				PerIteration: []*HTTPRecord{
					&HTTPRecord{
						HTTPRequestRecord: HTTPRequestRecord{
							Global: newFakeHistogram(t, 200, 300, 500, 600),
							PerStatus: map[int]*hdrhistogram.Histogram{
								200: newFakeHistogram(t, 200, 300, 500, 600),
							},
							PerOkKo: map[OkKo]*hdrhistogram.Histogram{
								Ok: newFakeHistogram(t, 200, 300, 500, 600),
							},
						},
						PerRequests: map[string]*HTTPRequestRecord{
							"req1": &HTTPRequestRecord{
								Global: newFakeHistogram(t, 200, 300, 500, 600),
								PerStatus: map[int]*hdrhistogram.Histogram{
									200: newFakeHistogram(t, 200, 300, 500, 600),
								},
								PerOkKo: map[OkKo]*hdrhistogram.Histogram{
									Ok: newFakeHistogram(t, 200, 300, 500, 600),
								},
							},
							"req2": &HTTPRequestRecord{
								Global: newFakeHistogram(t, 200, 300),
								PerStatus: map[int]*hdrhistogram.Histogram{
									200: newFakeHistogram(t, 200, 300),
								},
								PerOkKo: map[OkKo]*hdrhistogram.Histogram{
									Ok: newFakeHistogram(t, 200, 300),
								},
							},
							"req3": &HTTPRequestRecord{
								Global: newFakeHistogram(t, 500, 600),
								PerStatus: map[int]*hdrhistogram.Histogram{
									200: newFakeHistogram(t, 500, 600),
								},
								PerOkKo: map[OkKo]*hdrhistogram.Histogram{
									Ok: newFakeHistogram(t, 500, 600),
								},
							},
						},
					}, &HTTPRecord{
						HTTPRequestRecord: HTTPRequestRecord{
							Global: newFakeHistogram(t, 200, 300, 500, 600),
							PerStatus: map[int]*hdrhistogram.Histogram{
								200: newFakeHistogram(t, 200, 300, 500, 600),
								404: newFakeHistogram(t, 200, 400, 200, 400),
							},
							PerOkKo: map[OkKo]*hdrhistogram.Histogram{
								Ok: newFakeHistogram(t, 200, 300, 500, 600),
								Ko: newFakeHistogram(t, 200, 400, 200, 400),
							},
						},
						PerRequests: map[string]*HTTPRequestRecord{
							"req1": &HTTPRequestRecord{
								Global:    newFakeHistogram(t, 200, 300, 500, 600),
								PerStatus: map[int]*hdrhistogram.Histogram{},
								PerOkKo:   map[OkKo]*hdrhistogram.Histogram{},
							},
							"req2": &HTTPRequestRecord{
								Global: newFakeHistogram(t, 200, 300),
								PerStatus: map[int]*hdrhistogram.Histogram{
									404: newFakeHistogram(t, 200, 400),
								},
								PerOkKo: map[OkKo]*hdrhistogram.Histogram{
									Ko: newFakeHistogram(t, 200, 400),
								},
							},
							"req3": &HTTPRequestRecord{
								Global: newFakeHistogram(t, 500, 600),
								PerStatus: map[int]*hdrhistogram.Histogram{
									404: newFakeHistogram(t, 200, 400),
								},
								PerOkKo: map[OkKo]*hdrhistogram.Histogram{
									Ko: newFakeHistogram(t, 200, 400),
								},
							},
						},
					}, &HTTPRecord{
						HTTPRequestRecord: HTTPRequestRecord{
							Global: newFakeHistogram(t, 500, 600),
							PerStatus: map[int]*hdrhistogram.Histogram{
								200: newFakeHistogram(t, 500, 600),
								404: newFakeHistogram(t, 200, 400),
							},
							PerOkKo: map[OkKo]*hdrhistogram.Histogram{
								Ok: newFakeHistogram(t, 500, 600),
								Ko: newFakeHistogram(t, 200, 400),
							},
						},
						PerRequests: map[string]*HTTPRequestRecord{
							"req1": &HTTPRequestRecord{
								Global:    newFakeHistogram(t, 500, 600),
								PerStatus: map[int]*hdrhistogram.Histogram{},
								PerOkKo:   map[OkKo]*hdrhistogram.Histogram{},
							},
							"req3": &HTTPRequestRecord{
								Global: newFakeHistogram(t, 500, 600),
								PerStatus: map[int]*hdrhistogram.Histogram{
									404: newFakeHistogram(t, 200, 400),
								},
								PerOkKo: map[OkKo]*hdrhistogram.Histogram{
									Ko: newFakeHistogram(t, 200, 400),
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MergeHTTPRecordsOverTime(tt.args.rec1, tt.args.rec2); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergeHTTPRecordsOverTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func newFakeHistogram(t *testing.T, values ...int64) *hdr.Histogram {
	histo := hdr.New(0, 1000, 1)
	for _, v := range values {
		err := histo.RecordValue(v)
		require.NoError(t, err)
	}
	return histo
}
