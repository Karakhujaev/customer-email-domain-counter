Customer Email Domain Counter

A high-performance CLI tool to process customer CSV files and count email domains.
Optimized for large datasets with efficient memory and file handling.


1) Clone the Repository

```bash
git clone https://github.com/Karakhujaev/customer-email-domain-counter
cd customer-email-domain-counter
```

2) Run main.go file with sample file

```bash
go run main.go -file=samples/customers.csv 
```

3) Run Tests with Temporary File Samples

```bash
go test -v -cover ./customerimporter/...
go test -bench=. ./customerimporter/...
```

Project Structure

├── main.go                    Entry point CLI
├── customerimporter/          Core processing logic
│   ├── interview.go           CSV parsing and domain counting
│   └── interview_test.go      Tests and benchmarks
├── samples/                   Sample CSV data
│   └── customers.csv
├── go.mod                     Module definition
└── go.sum                     Dependency checksums


Features

Efficient CSV parsing using buffered IO

Concurrent processing with chunking

Outputs domain frequency in descending order

Clean architecture with unit tests and benchmarks
