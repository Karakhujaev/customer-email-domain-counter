Customer Email Domain Counter

A high-performance CLI tool to process customer CSV files and count email domains.
Optimized for large datasets with efficient memory and file handling.

Architecture

<img width="800" alt="Screenshot 2025-05-27 at 11 50 36" src="https://github.com/user-attachments/assets/adddb7de-0cfa-4dc2-86b8-bc0d168cbba7" />



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

```bash
📁 customer-email-domain-counter/
│── 📁 customerimporter/              Core processing logic
│── 📁 samples/                        Sample CSV data          
│── main.go                           Entry point CLI
│── go.mod                             
│── go.sum 
```

Features

Efficient CSV parsing using buffered IO

Concurrent processing with chunking

Outputs domain frequency in descending order

Clean architecture with unit tests and benchmarks
