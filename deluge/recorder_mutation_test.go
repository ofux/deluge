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

		// WaitGroup operations mutations
		{
			Name:        "WaitGroup_NoWaitRecording",
			Description: "Don't wait for all records to be queued",
			Original:    "r.recordingWaitGroup.Wait()",
			Mutated:     "// r.recordingWaitGroup.Wait()",
		},
		{
			Name:        "WaitGroup_NoWaitProcessing",
			Description: "Don't wait for all records to be processed",
			Original:    "r.processingWaitGroup.Wait()",
			Mutated:     "// r.processingWaitGroup.Wait()",
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
			Description: "Change <= to < in iteration check",
			Original:    "if len(histograms) <= rec.iteration {",
			Mutated:     "if len(histograms) < rec.iteration {",
		},
		{
			Name:        "Index_Wrong_Length_Calc",
			Description: "Wrong length calculation when expanding histograms",
			Original:    "diff := rec.iteration + 1 - len(histograms)",
			Mutated:     "diff := rec.iteration - len(histograms)",
		},

		// Map operations mutations
		{
			Name:        "Map_DontUpdate",
			Description: "Don't update histograms map after expansion",
			Original:    "histograms = append(histograms, r.createHistograms(diff)...)\n\t\t\t\tr.histograms[rec.id] = histograms",
			Mutated:     "histograms = append(histograms, r.createHistograms(diff)...)\n\t\t\t\t// r.histograms[rec.id] = histograms",
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

		// Constructor mutations
		{
			Name:        "Constructor_NoAddToWaitGroup",
			Description: "Don't add to processing wait group",
			Original:    "r.processingWaitGroup.Add(1)",
			Mutated:     "// r.processingWaitGroup.Add(1)",
		},

		// Instant recording mutations
		{
			Name:        "Instant_Record_NoWaitGroup",
			Description: "Don't wait for recording to complete",
			Original:    "\tgo func() {\n\t\tdefer r.recordingWaitGroup.Done()\n\t\tr.recordsQueue <- Record{\n\t\t\titeration: iteration,\n\t\t\tid:        id,\n\t\t\tvalue:     value,\n\t\t}\n\t}()",
			Mutated:     "\tfunc() {\n\t\tdefer r.recordingWaitGroup.Done()\n\t\tr.recordsQueue <- Record{\n\t\t\titeration: iteration,\n\t\t\tid:        id,\n\t\t\tvalue:     value,\n\t\t}\n\t}()",
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
