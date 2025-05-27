// Package customerimporter reads from a CSV file and returns a sorted (data
// structure of your choice) of email domains along with the number of customers
// with e-mail addresses for each domain. This should be able to be ran from the
// CLI and output the sorted domains to the terminal or to a file. Any errors
// should be logged (or handled). Performance matters (this is only ~3k lines,
// but could be 1m lines or run on a small machine).
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

type DomainResult struct {
	Domain string
	Count  int
}

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
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("failed to read mmap data: %w", err)
	}

	// Find header and determine email column index
	headerEnd := indexOfNewline(data, 0)
	if headerEnd == -1 {
		return fmt.Errorf("invalid CSV: no newline in file")
	}
	headerLine := data[:headerEnd]
	emailColIdx := findEmailColumnIndex(headerLine)
	if emailColIdx == -1 {
		return fmt.Errorf("email column not found in CSV header")
	}
	startPos := headerEnd + 1

	numWorkers := runtime.NumCPU() * 2
	chunks := chunkify(data, startPos, numWorkers)

	resultChan := make(chan map[string]int, numWorkers)
	var wg sync.WaitGroup

	for _, c := range chunks {
		wg.Add(1)
		go func(chunk []byte) {
			defer wg.Done()
			counts := processChunk(chunk, emailColIdx)
			resultChan <- counts
		}(c)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	finalCounts := make(map[string]int, 1_000_000)
	for counts := range resultChan {
		for domain, count := range counts {
			finalCounts[domain] += count
		}
	}

	sortedResults := sortDomains(finalCounts)
	return writeResults(outputPath, sortedResults)
}

func indexOfNewline(data []byte, start int) int {
	for i := start; i < len(data); i++ {
		if data[i] == '\n' {
			return i
		}
	}
	return -1
}

func findEmailColumnIndex(headerLine []byte) int {
	fields := strings.Split(string(headerLine), ",")
	for i, field := range fields {
		if strings.EqualFold(strings.TrimSpace(field), "email") {
			return i
		}
	}
	return -1
}

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

func processChunk(chunk []byte, emailColIdx int) map[string]int {
	counts := make(map[string]int, 50_000)
	lineStart := 0
	for lineStart < len(chunk) {
		lineEnd := indexOfNewline(chunk, lineStart)
		if lineEnd == -1 {
			lineEnd = len(chunk)
		}
		line := chunk[lineStart:lineEnd]
		if domain := parseDomainFromLine(line, emailColIdx); domain != "" {
			counts[domain]++
		}
		lineStart = lineEnd + 1
	}
	return counts
}

func parseDomainFromLine(line []byte, emailColIdx int) string {
	fieldStart := 0
	fieldNum := 0
	var email []byte

	for i := 0; i <= len(line); i++ {
		if i == len(line) || line[i] == ',' {
			if fieldNum == emailColIdx {
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
	for i := 0; i < len(domain); i++ {
		if domain[i] >= 'A' && domain[i] <= 'Z' {
			domain[i] += 'a' - 'A'
		}
	}
	return string(domain)
}

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

func writeResults(outputPath string, results []DomainResult) error {
	err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create output dir: %w", err)
	}

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

func generateOutputPath(csvPath string) string {
	dir := filepath.Dir(csvPath)
	filename := filepath.Base(csvPath)
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)
	return filepath.Join(dir, "outcomes", name+"_output"+ext)
}
