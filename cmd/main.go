// Package main provides the command-line interface for the organisation chart application.
// It processes transactions from a specified data directory and manages the organisation structure.
//
// Usage:
//
//	go run cmd/main.go -data <data_directory> [options]
//
// Required flags:
//
//	-data string
//	      Path to the data directory containing transactions
//
// Optional flags:
//
//	-init
//	      Initialize the database with government node
//	-type string
//	      Type of data to process: 'organisation' or 'person' (default: organisation)
//	-update_endpoint string
//	      Endpoint for the Update API (default "http://localhost:8080/entities")
//	-query_endpoint string
//	      Endpoint for the Query API (default "http://localhost:8081/v1/entities")
//
// Examples:
//
//  0. Get help:
//     go run cmd/main.go --help
//
//  1. Process organisation data with default settings:
//     go run cmd/main.go -data /path/to/data/directory
//
//  2. Process person data:
//     go run cmd/main.go -data /path/to/data/directory -type person
//
//  3. Initialize database and process organisation data:
//     go run cmd/main.go -data /path/to/data/directory -init
//
//  4. Use custom API endpoints:
//     go run cmd/main.go -data /path/to/data/directory -update_endpoint http://custom:8080/entities -query_endpoint http://custom:8081/v1/entities
//
// Process Types:
//   - organisation: Processes minister and department entities
//   - person: Processes citizen entities
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"orgchart_nexoan/api"
)

func main() {
	// Define command line flags with detailed descriptions
	dataDir := flag.String("data", "", "Path to the data directory containing transactions (required)")
	initDB := flag.Bool("init", false, "Initialize the database with government node before processing transactions")
	updateEndpoint := flag.String("update_endpoint", "http://localhost:8080/entities", "Endpoint for the Update API (default: http://localhost:8080/entities)")
	queryEndpoint := flag.String("query_endpoint", "http://localhost:8081/v1/entities", "Endpoint for the Query API (default: http://localhost:8081/v1/entities)")
	processType := flag.String("type", "organisation", "Type of data to process: 'organisation' or 'person' or 'document' or 'secretary' (default: organisation)")

	// Custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Process organisation chart transactions from a specified data directory.\n\n")
		fmt.Fprintf(os.Stderr, "Required flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  1. Process organisation data with default settings:\n")
		fmt.Fprintf(os.Stderr, "     %s -data /path/to/data/directory\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  2. Process person data:\n")
		fmt.Fprintf(os.Stderr, "     %s -data /path/to/data/directory -type person\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  3. Initialize database and process organisation data:\n")
		fmt.Fprintf(os.Stderr, "     %s -data /path/to/data/directory -init\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  4. Use custom API endpoints:\n")
		fmt.Fprintf(os.Stderr, "     %s -data /path/to/data/directory -update_endpoint http://custom:8080/entities -query_endpoint http://custom:8081/v1/entities\n\n", os.Args[0])
	}

	flag.Parse()

	// Validate data directory
	if *dataDir == "" {
		fmt.Fprintf(os.Stderr, "Error: Data directory path is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Validate process type
	if *processType != "organisation" && *processType != "person" && *processType != "document" && *processType != "secretary" {
		fmt.Fprintf(os.Stderr, "Error: Invalid process type. Must be 'organisation' or 'person' or 'document' or 'secretary'\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Ensure the data directory exists
	if _, err := os.Stat(*dataDir); os.IsNotExist(err) {
		log.Fatalf("Data directory does not exist: %s", *dataDir)
	}

	// Convert to absolute path
	absDataDir, err := filepath.Abs(*dataDir)
	if err != nil {
		log.Fatalf("Failed to get absolute path: %v", err)
	}

	// Create API client with configurable endpoints
	client := api.NewClient(*updateEndpoint, *queryEndpoint)

	// Initialize database if requested
	if *initDB {
		fmt.Println("Initializing database with government node...")
		government, err := client.CreateGovernmentNode()
		if err != nil {
			log.Fatalf("Failed to create government node: %v", err)
		}
		fmt.Printf("Successfully created government node with ID: %s\n", government.ID)
	}

	// Process transactions
	fmt.Printf("Processing %s transactions from directory: %s\n", *processType, absDataDir)
	if *processType == "document" {
		err = client.ProcessDocumentTransactions(absDataDir, *processType)
	} else if *processType == "secretary" {
		err = client.ProcessSecretaryAppointments(absDataDir)
	} else {
		err = client.ProcessTransactions(absDataDir, *processType)
	}

	if err != nil {
		log.Fatalf("Failed to process transactions: %v", err)
	}

	fmt.Println("Successfully processed all transactions")
}
