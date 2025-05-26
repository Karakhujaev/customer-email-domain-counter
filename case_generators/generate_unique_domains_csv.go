package main

import (
	"encoding/csv"
	"fmt"
	"os"
)

func main() {
	const totalRows = 1_000_0000
	file, err := os.Create("case_unique_domains.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	writer.Write([]string{"first_name", "last_name", "email", "gender", "ip_address"})

	// Generate 1 million unique domains
	for i := 0; i < totalRows; i++ {
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

	fmt.Println("✅ Generated case_unique_domains.csv with 1 million unique domains.")
}
