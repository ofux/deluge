package recording

import (
	hdr "github.com/ofux/hdrhistogram"
)

func MergeHTTPRecordsOverTime(rec1, rec2 *HTTPRecordsOverTime) *HTTPRecordsOverTime {
	if rec1 == nil {
		return rec2
	}
	if rec2 == nil {
		return rec1
	}

	if len(rec1.OverTime) < len(rec2.OverTime) {
		rec1, rec2 = rec2, rec1
	}
	merged := &HTTPRecordsOverTime{
		Global: mergeHTTPRecords(rec1.Global, rec2.Global),
	}
	for i, v1 := range rec1.OverTime {
		if i < len(rec2.OverTime) {
			merged.OverTime = append(merged.OverTime, mergeHTTPRecords(v1, rec2.OverTime[i]))
		} else {
			merged.OverTime = append(merged.OverTime, v1)
		}
	}
	return merged
}

func mergeHTTPRecords(rec1, rec2 *HTTPRecord) *HTTPRecord {
	if rec1 == nil {
		return rec2
	}
	if rec2 == nil {
		return rec1
	}

	merged := &HTTPRecord{
		HTTPRequestRecord: *mergeHTTPRequestRecords(&rec1.HTTPRequestRecord, &rec2.HTTPRequestRecord),
		PerRequests:       make(map[string]*HTTPRequestRecord),
	}

	for k, v1 := range rec1.PerRequests {
		if v2, ok := rec2.PerRequests[k]; ok {
			merged.PerRequests[k] = mergeHTTPRequestRecords(v1, v2)
		} else {
			merged.PerRequests[k] = v1
		}
	}
	for k, v2 := range rec2.PerRequests {
		if _, ok := merged.PerRequests[k]; !ok {
			merged.PerRequests[k] = v2
		}
	}

	return merged
}

func mergeHTTPRequestRecords(rec1, rec2 *HTTPRequestRecord) *HTTPRequestRecord {
	merged := &HTTPRequestRecord{
		Global:    mergeHistograms(rec1.Global, rec2.Global),
		PerStatus: make(map[int]*hdr.Histogram),
		PerOkKo:   make(map[OkKo]*hdr.Histogram),
	}

	for k, h1 := range rec1.PerStatus {
		if h2, ok := rec2.PerStatus[k]; ok {
			merged.PerStatus[k] = mergeHistograms(h1, h2)
		} else {
			merged.PerStatus[k] = copyHistogram(h1)
		}
	}
	for k, h2 := range rec2.PerStatus {
		if _, ok := merged.PerStatus[k]; !ok {
			merged.PerStatus[k] = copyHistogram(h2)
		}
	}

	for k, h1 := range rec1.PerOkKo {
		if h2, ok := rec2.PerOkKo[k]; ok {
			merged.PerOkKo[k] = mergeHistograms(h1, h2)
		} else {
			merged.PerOkKo[k] = copyHistogram(h1)
		}
	}
	for k, h2 := range rec2.PerOkKo {
		if _, ok := merged.PerOkKo[k]; !ok {
			merged.PerOkKo[k] = copyHistogram(h2)
		}
	}

	return merged
}
