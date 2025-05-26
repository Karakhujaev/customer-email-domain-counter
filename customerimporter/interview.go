package customerimporter

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"golang.org/x/exp/mmap"
)

// holds a domain and the count of domain.
type DomainResult struct {
	Domain string
	Count  int
}

// reads CSV file, counts email domains with mmap and concurrency, writes sorted output.
func ProcessCSVFile(csvPath string) error {
	outputPath := generateOutputPath(csvPath)

	reader, err := mmap.Open(csvPath)
	if err != nil {
		return fmt.Errorf("failed to mmap CSV file: %w", err)
	}
	defer reader.Close()

	fileSize := int(reader.Len())
	data := make([]byte, fileSize)
	_, err = reader.ReadAt(data, 0)
	if err != nil && !errors.Is(err,io.EOF) {
		return fmt.Errorf("failed to read mmap data: %w", err)
	}

	// skip header line
	headerEnd := indexOfNewline(data, 0)
	if headerEnd == -1 {
		return fmt.Errorf("invalid CSV: no newline in file")
	}
	startPos := headerEnd + 1

	numWorkers := runtime.NumCPU() * 2
	chunks := chunkify(data, startPos, numWorkers)

	// channels for results and waitgroup for workers
	resultChan := make(chan map[string]int, numWorkers)
	var wg sync.WaitGroup

	// workers on chunks
	for _, c := range chunks {
		wg.Add(1)
		go func(chunk []byte) {
			defer wg.Done()
			counts := processChunk(chunk)
			resultChan <- counts
		}(c)
	}

	// close resultChan when all workers done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// merge all results from workers
	finalCounts := make(map[string]int, 1_000_000) 
	for counts := range resultChan {
		for domain, count := range counts {
			finalCounts[domain] += count
		}
	}

	// sort results
	sortedResults := sortDomains(finalCounts)

	// write output
	return writeResults(outputPath, sortedResults)
}

// indexOfNewline returns index of '\n' or -1 if none
func indexOfNewline(data []byte, start int) int {
	for i := start; i < len(data); i++ {
		if data[i] == '\n' {
			return i
		}
	}
	return -1
}

// chunkify splits data into n chunks aligned on newlines
func chunkify(data []byte, start int, n int) [][]byte {
	size := len(data) - start
	chunkSize := size / n

	chunks := make([][]byte, 0, n)
	chunkStart := start

	for i := 0; i < n; i++ {
		if chunkStart >= len(data) {
			break
		}
		end := chunkStart + chunkSize
		if i == n-1 || end >= len(data) {
			end = len(data)
			chunks = append(chunks, data[chunkStart:end])
			break
		}
		// move end forward to next newline to not split lines
		for end < len(data) && data[end] != '\n' {
			end++
		}
		if end < len(data) {
			end++ 
		}
		chunks = append(chunks, data[chunkStart:end])
		chunkStart = end
	}

	return chunks
}

// processChunk parses a chunk of CSV lines, counts domains
func processChunk(chunk []byte) map[string]int {
	counts := make(map[string]int, 50_000) 

	lineStart := 0
	for lineStart < len(chunk) {
		// find newline or end of chunk
		lineEnd := indexOfNewline(chunk, lineStart)
		if lineEnd == -1 {
			lineEnd = len(chunk)
		}

		line := chunk[lineStart:lineEnd]

		// Process line 
		if domain := parseDomainFromLine(line); domain != "" {
			counts[domain]++
		}

		lineStart = lineEnd + 1
	}

	return counts
}

// parseDomainFromLine manually parses the CSV line and extracts domain from 3rd column (email)
func parseDomainFromLine(line []byte) string {
	// minimal CSV parse by splitting on commas, ignoring quoted commas for simplicity here

	fieldStart := 0
	fieldNum := 0
	var email []byte

	for i := 0; i <= len(line); i++ {
		if i == len(line) || line[i] == ',' {
			if fieldNum == 2 {
				email = line[fieldStart:i]
				break
			}
			fieldNum++
			fieldStart = i + 1
		}
	}

	if len(email) == 0 {
		return ""
	}

	return extractDomain(email)
}

// extractDomain extracts and lowercases domain from email byte slice
func extractDomain(email []byte) string {
	atIndex := -1
	for i := len(email) - 1; i >= 0; i-- {
		if email[i] == '@' {
			atIndex = i
			break
		}
	}
	if atIndex == -1 || atIndex == len(email)-1 {
		return ""
	}

	domain := email[atIndex+1:]

	// validate domain: must contain '.'
	hasDot := false
	for _, b := range domain {
		if b == '.' {
			hasDot = true
			break
		}
	}
	if !hasDot || len(domain) < 3 {
		return ""
	}

	// lowercase in place without allocation
	for i := 0; i < len(domain); i++ {
		b := domain[i]
		if b >= 'A' && b <= 'Z' {
			domain[i] = b + 32
		}
	}

	return string(domain)
}

// sortDomains sorts by count desc, then domain asc
func sortDomains(counts map[string]int) []DomainResult {
	results := make([]DomainResult, 0, len(counts))
	for d, c := range counts {
		results = append(results, DomainResult{Domain: d, Count: c})
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Count != results[j].Count {
			return results[i].Count > results[j].Count
		}
		return results[i].Domain < results[j].Domain
	})
	return results
}

// writeResults writes CSV output with buffer
func writeResults(outputPath string, results []DomainResult) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	bufSize := 1024 * 1024
	writer := bufio.NewWriterSize(f, bufSize)
	defer writer.Flush()

	_, err = writer.WriteString("Domain,Count\n")
	if err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	var sb strings.Builder
	sb.Grow(64)
	for _, r := range results {
		sb.Reset()
		sb.WriteString(r.Domain)
		sb.WriteByte(',')
		sb.WriteString(fmt.Sprintf("%d", r.Count))
		sb.WriteByte('\n')
		if _, err = writer.WriteString(sb.String()); err != nil {
			return fmt.Errorf("failed to write result: %w", err)
		}
	}

	log.Printf("Results written to: %s", outputPath)
	return nil
}

// generateOutputPath appends "_output" before extension
func generateOutputPath(csvPath string) string {
	dir := filepath.Dir(csvPath)
	filename := filepath.Base(csvPath)
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)
	return filepath.Join(dir, "/outcomes/", name+"_output"+ext)
}
