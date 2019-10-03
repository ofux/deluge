package recordingtest

import (
	"github.com/ofux/deluge/core/recording"
	hdr "github.com/ofux/hdrhistogram"
	"testing"
)

func testHistogram(t *testing.T, histogram *hdr.Histogram, minTotalCount int64) {
	if histogram.TotalCount() < minTotalCount {
		t.Errorf("Expected to have TotalCount >= %d, got %d", minTotalCount, histogram.TotalCount())
	}
}

func testHistogramPerStatus(t *testing.T, rec *recording.HTTPRequestRecord, status int, minTotalCount int64) {
	resultPerStatus, ok := rec.PerStatus[status]
	if !ok {
		t.Fatalf("Expected to have some records for status %d", status)
	}
	testHistogram(t, resultPerStatus, minTotalCount)
}

func testHistogramPerOkKo(t *testing.T, rec *recording.HTTPRequestRecord, okKo recording.OkKo, minTotalCount int64) {
	resultPerOkKo, ok := rec.PerOkKo[okKo]
	if !ok {
		t.Fatalf("Expected to have some records for OK/KO %s", okKo)
	}
	testHistogram(t, resultPerOkKo, minTotalCount)
}

func testHistogramPerRequest(t *testing.T, rec *recording.HTTPRecord, reqName string, minTotalCount int64, status int, okKo recording.OkKo) {
	resultPerRequest, ok := rec.PerRequests[reqName]
	if !ok {
		t.Fatalf("Expected to have some records for request '%s'", reqName)
	}
	testHistogram(t, resultPerRequest.Global, minTotalCount)
	testHistogramPerOkKo(t, resultPerRequest, okKo, minTotalCount)
	testHistogramPerStatus(t, resultPerRequest, status, minTotalCount)
}

func CheckHTTPRecord(t *testing.T, rec *recording.HTTPRecord, reqName string, minTotalCount int64, status int, okKo recording.OkKo) {
	testHistogram(t, rec.Global, minTotalCount)
	testHistogramPerOkKo(t, &rec.HTTPRequestRecord, okKo, minTotalCount)
	testHistogramPerStatus(t, &rec.HTTPRequestRecord, status, minTotalCount)
	testHistogramPerRequest(t, rec, reqName, minTotalCount, status, okKo)
}
