# EmailVal
This Go program checks DNS records for multiple domains, including their MX, SPF, and DMARC records. It performs concurrent DNS lookups and allows outputting results to a CSV file or the console.

## Features


- **Concurrent DNS lookups**: The program checks domains concurrently using Goroutines.
- **MX Record Lookup**: Checks if a domain has MX (Mail Exchange) records.
- **SPF Record Lookup**: Looks for SPF (Sender Policy Framework) records in TXT records.
- **DMARC Record Lookup**: Looks for DMARC (Domain-based Message Authentication, Reporting, and Conformance) records.
- **CSV Output Support**: Option to export the results to a CSV file.

## Prerequisites
- Go installed (v1.16+)

## Installation

1. Clone the repository or copy the source code.
2. Navigate to the project directory.
3. Build the program:
   ```bash
   go build -o domain-checker
