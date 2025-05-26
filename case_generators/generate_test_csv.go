package main

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
)

var domains = []string{
	"gmail.com", "yahoo.com", "outlook.com", "example.com",
	"protonmail.com", "aol.com", "mail.com", "icloud.com",
}

func randomEmail(id int) string {
	return fmt.Sprintf("user%d@%s", id, domains[rand.Intn(len(domains))])
}

func main() {
	const totalRows = 1_000_000
	file, err := os.Create("customers.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	writer.Write([]string{"first_name", "last_name", "email", "gender", "ip_address"})

	// Write rows
	for i := 0; i < totalRows; i++ {
		row := []string{
			fmt.Sprintf("First%d", i),
			fmt.Sprintf("Last%d", i),
			randomEmail(i),
			[]string{"Male", "Female"}[rand.Intn(2)],
			fmt.Sprintf("%d.%d.%d.%d", rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(256)),
		}
		writer.Write(row)
	}

	fmt.Println("✅ Generated case_million_rows.csv with 1 million rows.")
}
