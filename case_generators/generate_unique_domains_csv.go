package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	var rowCount int
	var fileName string

	fmt.Print("Enter number of rows to generate: ")
	_, err := fmt.Scanf("%d\n", &rowCount)
	if err != nil || rowCount <= 0 {
		fmt.Println("Invalid input. Please enter a positive integer for row count.")
		return
	}

	fmt.Print("Enter output file name for example -> data.csv: ")
	_, err = fmt.Scanf("%s\n", &fileName)
	if err != nil || fileName == "" {
		fmt.Println("Invalid input. Please provide a valid file name.")
		return
	}

	outputDir := "samples"
	err = os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		fmt.Printf("Failed to create directory '%s': %v\n", outputDir, err)
		return
	}

	fullPath := filepath.Join(outputDir, fileName)

	file, err := os.Create(fullPath)
	if err != nil {
		fmt.Printf("Failed to create file: %v\n", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"first_name", "last_name", "email", "gender", "ip_address"})

	for i := 0; i < rowCount; i++ {
		domain := fmt.Sprintf("domain%d.com", i)
		email := fmt.Sprintf("user%d@%s", i, domain)
		row := []string{
			fmt.Sprintf("First%d", i),
			fmt.Sprintf("Last%d", i),
			email,
			[]string{"Male", "Female"}[i%2],
			fmt.Sprintf("%d.%d.%d.%d", i%256, (i/256)%256, (i/65536)%256, (i/16777216)%256),
		}
		writer.Write(row)
	}

	fmt.Printf("File '%s' generated in 'samples/' directory with %d unique email domains.\n", fileName, rowCount)
}
