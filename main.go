package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"customerimporter/customerimporter"
)

func main() {
	// Command line flags
	csvPath := flag.String("file", "", "Path to the CSV file to process (required)")
	help := flag.Bool("help", false, "Show help message")
	flag.Parse()

	if *help || *csvPath == "" {
		showUsage()
		return
	}

	// Validate file existence
	if _, err := os.Stat(*csvPath); os.IsNotExist(err) {
		log.Fatalf("File does not exist: %s", *csvPath)
	}

	// Validate .csv extension
	if ext := filepath.Ext(*csvPath); ext != ".csv" {
		log.Fatalf("File must have .csv extension, got: %s", ext)
	}

	fmt.Printf("Processing CSV file: %s\n", *csvPath)
	start := time.Now()

	// Process file
	if err := customerimporter.ProcessCSVFile(*csvPath); err != nil {
		log.Fatalf("Error processing CSV file: %v", err)
	}

	fmt.Printf("Processing completed in: %v\n", time.Since(start))
}

func showUsage() {
	fmt.Println("Customer Domain Counter")
	fmt.Println("=======================")
	fmt.Println("Processes a CSV file and counts email domains.")
	fmt.Println("Results sorted by count (desc) are saved in an output CSV file.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Printf("  %s -file=<path_to_csv_file>\n", os.Args[0])
	fmt.Println()
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Printf("  %s -file=customers.csv\n", os.Args[0])
	fmt.Printf("  %s -file=/path/to/data.csv\n", os.Args[0])
	fmt.Println()
	fmt.Println("Output:")
	fmt.Println("  Creates a CSV file with '_output' suffix before extension.")
	fmt.Println("  Example: customers.csv -> customers_output.csv")
}
