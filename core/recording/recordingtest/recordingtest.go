package recordingtest

import (
	hdr "github.com/codahale/hdrhistogram"
	"github.com/ofux/deluge/core/recording"
	"testing"
)

func testHistogram(t *testing.T, histogram *hdr.Histogram, totalCount int64) {
	if histogram.TotalCount() != totalCount {
		t.Errorf("Expected to have totalCount = %d, got %d", totalCount, histogram.TotalCount())
	}
}

func testHistogramPerStatus(t *testing.T, rec *recording.HTTPRequestRecord, status int, totalCount int64) {
	resultPerStatus, ok := rec.PerStatus[status]
	if !ok {
		t.Fatalf("Expected to have some records for status %d", status)
	}
	testHistogram(t, resultPerStatus, totalCount)
}

func testHistogramPerOkKo(t *testing.T, rec *recording.HTTPRequestRecord, okKo recording.OkKo, totalCount int64) {
	resultPerOkKo, ok := rec.PerOkKo[okKo]
	if !ok {
		t.Fatalf("Expected to have some records for OK/KO %s", okKo)
	}
	testHistogram(t, resultPerOkKo, totalCount)
}

func testHistogramPerRequest(t *testing.T, rec *recording.HTTPRecord, reqName string, totalCount int64, status int, okKo recording.OkKo) {
	resultPerRequest, ok := rec.PerRequests[reqName]
	if !ok {
		t.Fatalf("Expected to have some records for request '%s'", reqName)
	}
	testHistogram(t, resultPerRequest.Global, totalCount)
	testHistogramPerOkKo(t, resultPerRequest, okKo, totalCount)
	testHistogramPerStatus(t, resultPerRequest, status, totalCount)
}

func CheckHTTPRecord(t *testing.T, rec *recording.HTTPRecord, reqName string, totalCount int64, status int, okKo recording.OkKo) {
	testHistogram(t, rec.Global, totalCount)
	testHistogramPerOkKo(t, &rec.HTTPRequestRecord, okKo, totalCount)
	testHistogramPerStatus(t, &rec.HTTPRequestRecord, status, totalCount)
	testHistogramPerRequest(t, rec, reqName, totalCount, status, okKo)
}
