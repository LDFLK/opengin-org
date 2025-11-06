package main

import (
	"fmt"
	"log"
	"orgchart_nexoan/api"
)

func main() {
	fmt.Println("Starting Secretary Process...")
	// main logic for secretary processing
	// Initialize the API client with default endpoints
	updateEndpoint := "http://localhost:8080/entities"
	queryEndpoint := "http://localhost:8081/v1/entities"
	client := api.NewClient(updateEndpoint, queryEndpoint)
	
	// Process all secretary operations
	//including secratary processing  the secratary ...
	err := client.ProcessAllSecretaryOperations()
	if err != nil {
		log.Fatalf("Error processing secretary operations: %v", err)
	}
	
	fmt.Println("Secretary processing completed successfully!")
}