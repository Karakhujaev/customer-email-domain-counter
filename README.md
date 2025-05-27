Customer Email Domain Counter

A high-performance CLI tool to process customer CSV files and count email domains.
Optimized for large datasets with efficient memory and file handling.

Architecture


<img width="1024" alt="Screenshot 2025-05-27 at 11 52 34" src="https://github.com/user-attachments/assets/7f6b3727-cb15-4b61-a35c-a232ed37b175" />


<br>
<br>
<br>

1) Clone the Repository

```bash
git clone https://github.com/Karakhujaev/customer-email-domain-counter
cd customer-email-domain-counter
```

2) Run main.go file with sample file

```bash
go run main.go -file={file_path} 
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
