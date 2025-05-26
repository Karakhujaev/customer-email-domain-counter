package customerimporter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// creates temp CSV file with given content
func createTempCSVFile(t testing.TB, content string) (string, func()) {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "testcsv_*.csv")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}
	return tmpFile.Name(), func() { os.Remove(tmpFile.Name()) }
}

// TestProcessCSVFile tests ProcessCSVFile correctness with small input
func TestProcessCSVFile(t *testing.T) {
	csvContent := `id,name,email
1,John,john@example.com
2,Jane,jane@domain.org
3,Bob,bob@example.com
`

	path, cleanup := createTempCSVFile(t, csvContent)
	defer cleanup()

	// create output directory to match generateOutputPath's directory
	outDir := filepath.Dir(generateOutputPath(path))
	if err := os.MkdirAll(outDir, 0755); err != nil {
		t.Fatalf("failed to create output directory: %v", err)
	}

	err := ProcessCSVFile(path)
	if err != nil {
		t.Fatalf("ProcessCSVFile failed: %v", err)
	}

	outPath := generateOutputPath(path)
	outContent, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	outStr := string(outContent)
	if !strings.Contains(outStr, "example.com,2") {
		t.Errorf("expected example.com count 2 in output, got:\n%s", outStr)
	}
	if !strings.Contains(outStr, "domain.org,1") {
		t.Errorf("expected domain.org count 1 in output, got:\n%s", outStr)
	}

	// clean up output file as well
	if err := os.Remove(outPath); err != nil {
		t.Logf("failed to remove output file: %v", err)
	}
}


// benchmarkProcessCSVFile benchmarks
func BenchmarkProcessCSVFile(b *testing.B) {
	var builder strings.Builder
	builder.WriteString("id,name,email\n")

	for i := 0; i < 100_000; i++ {// 100k lines with same domain
		builder.WriteString(fmt.Sprintf("%d,User%d,user%d@example.com\n", i, i, i))
	}
	content := builder.String()

	path, cleanup := createTempCSVFile(b, content)
	defer cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := ProcessCSVFile(path)
		if err != nil {
			b.Fatalf("ProcessCSVFile failed: %v", err)
		}
	}
}
