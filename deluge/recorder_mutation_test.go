//go:build !race

package deluge

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// Mutation represents a code change to test
type Mutation struct {
	Name        string
	Description string
	Original    string
	Mutated     string
}

// TestMutations runs mutation testing on recorder.go
func TestMutations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping mutation tests in short mode")
	}

	filePath := "recorder.go"

	// Read original file
	originalContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read %s: %v", filePath, err)
	}

	defer func() {
		// Ensure the original file is restored even in case of panic
		if r := recover(); r != nil {
			_ = os.WriteFile(filePath, originalContent, 0644)
			panic(r)
		} else {
			_ = os.WriteFile(filePath, originalContent, 0644)
		}
	}()

	mutations := []Mutation{

		// State transitions mutations
		{
			Name:        "State_SkipRecording",
			Description: "Don't set RECORDING state in processRecords",
			Original:    "r.recording = RECORDING",
			Mutated:     "r.recording = READY",
		},
		{
			Name:        "State_SkipTerminated",
			Description: "Don't set TERMINATED state after close",
			Original:    "r.recording = TERMINATED",
			Mutated:     "// r.recording = TERMINATED",
		},

		// Condition mutations
		{
			Name:        "Condition_CloseWrongState",
			Description: "Change Close condition from RECORDING to READY",
			Original:    "if r.recording == RECORDING {",
			Mutated:     "if r.recording == READY {",
		},
		{
			Name:        "Condition_GetRecordsWrongState",
			Description: "Change GetRecords condition to allow non-TERMINATED state",
			Original:    "if r.recording != TERMINATED {",
			Mutated:     "if r.recording == READY {",
		},

		// Channel operations mutations
		{
			Name:        "Channel_NoClose",
			Description: "Don't close the records queue",
			Original:    "close(r.recordsQueue)",
			Mutated:     "// close(r.recordsQueue)",
		},
		{
			Name:        "Channel_NoWait",
			Description: "Don't wait for termination",
			Original:    "<-r.onTerminated",
			Mutated:     "// <-r.onTerminated",
		},
		{
			Name:        "Channel_NoSignal",
			Description: "Don't signal termination",
			Original:    "r.onTerminated <- struct{}{}",
			Mutated:     "// r.onTerminated <- struct{}{}",
		},

		// Iteration/indexing mutations
		{
			Name:        "Index_OffByOne_Plus",
			Description: "Add 1 to iteration index when recording",
			Original:    "histogram := histograms[rec.iteration]",
			Mutated:     "histogram := histograms[rec.iteration+1]",
		},
		{
			Name:        "Index_OffByOne_Minus",
			Description: "Subtract 1 from iteration check",
			Original:    "if len(histograms) < rec.iteration+1 {",
			Mutated:     "if len(histograms) < rec.iteration {",
		},
		{
			Name:        "Index_Wrong_Length_Calc",
			Description: "Wrong length calculation when expanding histograms",
			Original:    "histograms = append(histograms, createHistograms(rec.iteration+1-len(histograms))...)",
			Mutated:     "histograms = append(histograms, createHistograms(rec.iteration-len(histograms))...)",
		},

		// Map operations mutations
		{
			Name:        "Map_DontUpdate",
			Description: "Don't update histograms map after expansion",
			Original:    "histograms = append(histograms, createHistograms(rec.iteration+1-len(histograms))...)\n\t\t\t\tr.histograms[rec.id] = histograms",
			Mutated:     "histograms = append(histograms, createHistograms(rec.iteration+1-len(histograms))...)\n\t\t\t\t// r.histograms[rec.id] = histograms",
		},

		// Value recording mutations
		{
			Name:        "Record_DontRecordValue",
			Description: "Skip recording the value in histogram",
			Original:    "histogram.RecordValue(rec.value)",
			Mutated:     "// histogram.RecordValue(rec.value)",
		},
		{
			Name:        "Record_WrongValue",
			Description: "Record wrong value (0 instead of actual)",
			Original:    "histogram.RecordValue(rec.value)",
			Mutated:     "histogram.RecordValue(0)",
		},

		// Loop mutations
		{
			Name:        "Loop_BreakEarly",
			Description: "Break after first record",
			Original:    "for {\n\t\t\trec, ok := <-r.recordsQueue\n\t\t\tif !ok {\n\t\t\t\treturn\n\t\t\t}",
			Mutated:     "for {\n\t\t\trec, ok := <-r.recordsQueue\n\t\t\tif !ok {\n\t\t\t\treturn\n\t\t\t}\n\t\t\treturn",
		},
		{
			Name:        "Loop_IgnoreChannelClosed",
			Description: "Ignore channel closed signal",
			Original:    "if !ok {\n\t\t\t\treturn\n\t\t\t}",
			Mutated:     "if !ok {\n\t\t\t\t// return\n\t\t\t}",
		},

		// HDR Histogram creation mutations
		{
			Name:        "Histogram_WrongPrecision",
			Description: "Change histogram precision from 3 to 1",
			Original:    "histograms[i] = hdr.New(0, 3600000000, 3)",
			Mutated:     "histograms[i] = hdr.New(0, 3600000000, 1)",
		},
		{
			Name:        "Histogram_WrongMaxValue",
			Description: "Change histogram max value",
			Original:    "histograms[i] = hdr.New(0, 3600000000, 3)",
			Mutated:     "histograms[i] = hdr.New(0, 1000000, 3)",
		},
	}

	results := make(map[string]bool)
	var survived []string
	var killed []string
	var skipped []string

	for _, mutation := range mutations {
		t.Run(mutation.Name, func(t *testing.T) {
			// Check if mutation can be applied
			if !strings.Contains(string(originalContent), mutation.Original) {
				t.Errorf("Original code not found. Looking for: %q", mutation.Original)
				results[mutation.Name] = false
				skipped = append(skipped, mutation.Name)
				return
			}

			// Apply mutation
			mutatedContent := strings.Replace(string(originalContent), mutation.Original, mutation.Mutated, 1)
			if err := os.WriteFile(filePath, []byte(mutatedContent), 0644); err != nil {
				t.Fatalf("Failed to write mutated file: %v", err)
			}

			// Ensure we restore the original file
			defer func() {
				if err := os.WriteFile(filePath, originalContent, 0644); err != nil {
					t.Fatalf("Failed to restore original file: %v", err)
				}
			}()

			// Run tests with a short timeout to catch hanging mutations
			cmd := exec.Command("go", "test", "-run", "Test__Verifier__", "-timeout", "20s")
			output, err := cmd.CombinedOutput()

			// Tests should FAIL when mutation is introduced (mutation is "killed")
			if err != nil {
				t.Logf("KILLED: %s\n   Tests correctly failed with this mutation\n   %s", mutation.Description, mutation.Name)
				results[mutation.Name] = true
				killed = append(killed, mutation.Name)
			} else {
				t.Errorf("SURVIVED: %s\n   Tests passed despite mutation - possible coverage gap!\n   Mutation: %s\n   Output: %s",
					mutation.Description, mutation.Name, string(output))
				results[mutation.Name] = false
				survived = append(survived, mutation.Name)
			}
		})
	}

	// Print summary
	t.Log("\n" + strings.Repeat("=", 80))
	t.Log("MUTATION TESTING SUMMARY")
	t.Log(strings.Repeat("=", 80))
	t.Logf("Total mutations: %d", len(mutations))
	t.Logf("Killed (detected): %d", len(killed))
	t.Logf("Survived (not detected): %d", len(survived))
	t.Logf("Skipped: %d", len(skipped))
	t.Logf("Mutation Score: %.1f%%", float64(len(killed))/float64(len(mutations)-len(skipped))*100)

	if len(survived) > 0 {
		t.Log("\n⚠️  MUTATIONS THAT SURVIVED (coverage gaps):")
		for _, name := range survived {
			t.Logf("  - %s", name)
		}
	}

	if len(skipped) > 0 {
		t.Log("\n⚠️  SKIPPED MUTATIONS:")
		for _, name := range skipped {
			t.Logf("  - %s", name)
		}
	}

	t.Log(strings.Repeat("=", 80))
}
