package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

//Author: Morteza Farrokhnejad
type DomainInfo struct {
	Domain      string
	HasMX       bool
	HasSPF      bool
	SPFRecord   string
	HasDMARC    bool
	DMARCRecord string
	MXRecords   []string
	Errors      []string
}

func main() {
	outputFile := flag.String("output", "", "Output CSV file (optional)")
	concurrency := flag.Int("concurrency", 10, "Number of concurrent goroutines")
	timeout := flag.Duration("timeout", 10*time.Second, "Timeout for DNS lookups")
	flag.Parse()

	var writer *csv.Writer
	if *outputFile != "" {
		file, err := os.Create(*outputFile)
		if err != nil {
			log.Fatalf("Error creating output file: %v\n", err)
		}
		defer file.Close()
		writer = csv.NewWriter(file)
		defer writer.Flush()
	}

	printHeader(writer)

	scanner := bufio.NewScanner(os.Stdin)
	domainChan := make(chan string)
	resultChan := make(chan DomainInfo)

	var wg sync.WaitGroup
	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for domain := range domainChan {
				resultChan <- checkDomain(domain, *timeout)
			}
		}()
	}

	go func() {
		for scanner.Scan() {
			domainChan <- scanner.Text()
		}
		close(domainChan)
	}()

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for result := range resultChan {
		printResult(writer, result)
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading from input: %v\n", err)
	}
}

func checkDomain(domain string, timeout time.Duration) DomainInfo {
	info := DomainInfo{Domain: domain}
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: timeout}
			return d.DialContext(ctx, network, address)
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Lookup MX records
	mxRecords, err := resolver.LookupMX(ctx, domain)
	if err != nil {
		info.Errors = append(info.Errors, fmt.Sprintf("MX lookup error: %v", err))
	} else {
		info.HasMX = len(mxRecords) > 0
		for _, mx := range mxRecords {
			info.MXRecords = append(info.MXRecords, mx.Host)
		}
	}

	// Lookup TXT records (for SPF)
	txtRecords, err := resolver.LookupTXT(ctx, domain)
	if err != nil {
		info.Errors = append(info.Errors, fmt.Sprintf("TXT lookup error: %v", err))
	} else {
		for _, record := range txtRecords {
			if strings.HasPrefix(record, "v=spf1") {
				info.HasSPF = true
				info.SPFRecord = record
				break
			}
		}
	}

	// Lookup DMARC records
	dmarcRecords, err := resolver.LookupTXT(ctx, "_dmarc."+domain)
	if err != nil {
		info.Errors = append(info.Errors, fmt.Sprintf("DMARC lookup error: %v", err))
	} else {
		for _, record := range dmarcRecords {
			if strings.HasPrefix(record, "v=DMARC1") {
				info.HasDMARC = true
				info.DMARCRecord = record
				break
			}
		}
	}

	return info
}

func printHeader(writer *csv.Writer) {
	header := []string{"Domain", "Has MX", "Has SPF", "SPF Record", "Has DMARC", "DMARC Record", "MX Records", "Errors"}
	if writer != nil {
		writer.Write(header)
	} else {
		fmt.Println(strings.Join(header, ","))
	}
}

func printResult(writer *csv.Writer, info DomainInfo) {
	record := []string{
		info.Domain,
		fmt.Sprintf("%v", info.HasMX),
		fmt.Sprintf("%v", info.HasSPF),
		info.SPFRecord,
		fmt.Sprintf("%v", info.HasDMARC),
		info.DMARCRecord,
		strings.Join(info.MXRecords, ";"),
		strings.Join(info.Errors, ";"),
	}

	if writer != nil {
		writer.Write(record)
	} else {
		fmt.Println(strings.Join(record, ","))
	}
}