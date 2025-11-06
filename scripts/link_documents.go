package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"orgchart_nexoan/api"
	"orgchart_nexoan/models"
)

// DocumentLink represents a document relationship from CSV
type DocumentLink struct {
	Parent       string
	Child        string
	Relationship string
	StartDate    string
}

// LinkDocuments processes document linking CSV files and creates relationships
func LinkDocuments(client *api.Client, csvFilePath string) error {
	// Read CSV file
	links, err := readDocumentLinksCSV(csvFilePath)
	if err != nil {
		return fmt.Errorf("failed to read CSV file %s: %w", csvFilePath, err)
	}

	fmt.Printf("Processing %d document links from %s\n", len(links), csvFilePath)

	// Process each link
	successCount := 0
	errorCount := 0

	for i, link := range links {
		fmt.Printf("Processing link %d/%d: %s -> %s (%s)\n",
			i+1, len(links), link.Parent, link.Child, link.Relationship)

		err := createDocumentRelationship(client, link)
		if err != nil {
			log.Printf("Error creating relationship %s -> %s: %v", link.Parent, link.Child, err)
			errorCount++
		} else {
			successCount++
		}
	}

	fmt.Printf("Completed processing %s: %d successful, %d errors\n",
		csvFilePath, successCount, errorCount)

	return nil
}

// readDocumentLinksCSV reads a CSV file and returns document links
func readDocumentLinksCSV(filePath string) ([]DocumentLink, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file must have at least a header and one data row")
	}

	// Skip header row
	var links []DocumentLink
	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) < 4 {
			return nil, fmt.Errorf("row %d has insufficient columns (expected 4, got %d)", i+1, len(record))
		}

		link := DocumentLink{
			Parent:       strings.TrimSpace(record[0]),
			Child:        strings.TrimSpace(record[1]),
			Relationship: strings.TrimSpace(record[2]),
			StartDate:    strings.TrimSpace(record[3]),
		}

		// Validate required fields
		if link.Parent == "" || link.Child == "" || link.Relationship == "" || link.StartDate == "" {
			return nil, fmt.Errorf("row %d has empty required fields", i+1)
		}

		links = append(links, link)
	}

	return links, nil
}

// createDocumentRelationship creates a relationship between two documents
func createDocumentRelationship(client *api.Client, link DocumentLink) error {
	// Parse the start date
	date, err := time.Parse("2006-01-02", link.StartDate)
	if err != nil {
		return fmt.Errorf("failed to parse date %s: %w", link.StartDate, err)
	}
	dateISO := date.Format(time.RFC3339)

	// Find parent document entity
	parentEntity, err := findDocumentByName(client, link.Parent)
	if err != nil {
		return fmt.Errorf("failed to find parent document '%s': %w", link.Parent, err)
	}

	// Find child document entity
	childEntity, err := findDocumentByName(client, link.Child)
	if err != nil {
		return fmt.Errorf("failed to find child document '%s': %w", link.Child, err)
	}

	// Create the relationship
	err = createRelationship(client, parentEntity.ID, childEntity.ID, link.Relationship, dateISO)
	if err != nil {
		return fmt.Errorf("failed to create relationship: %w", err)
	}

	fmt.Printf("Successfully created relationship: %s (%s) -> %s (%s) [%s]\n",
		link.Parent, parentEntity.ID, link.Child, childEntity.ID, link.Relationship)

	return nil
}

// findDocumentByName searches for a document entity by name
func findDocumentByName(client *api.Client, documentName string) (*models.Entity, error) {
	// Search for document entities with the given name
	searchCriteria := &models.SearchCriteria{
		Kind: &models.Kind{
			Major: "Document",
		},
		Name: documentName,
	}

	results, err := client.SearchEntities(searchCriteria)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("document not found: %s", documentName)
	}

	if len(results) > 1 {
		return nil, fmt.Errorf("multiple documents found with name: %s", documentName)
	}

	// Convert SearchResult to Entity
	result := results[0]
	entity := &models.Entity{
		ID:         result.ID,
		Kind:       result.Kind,
		Created:    result.Created,
		Terminated: result.Terminated,
		Name: models.TimeBasedValue{
			Value: result.Name,
		},
		Metadata:      []models.MetadataEntry{},
		Attributes:    []models.AttributeEntry{},
		Relationships: []models.RelationshipEntry{},
	}

	return entity, nil
}

// createRelationship creates a relationship between two entities
func createRelationship(client *api.Client, parentID, childID, relationshipType, startTime string) error {
	// Generate unique relationship ID with random number- rand package automatically seeds with a random number
	randomNum := rand.Intn(999999)
	uniqueRelationshipID := fmt.Sprintf("%s_%s_%06d", parentID, childID, randomNum)

	// Create the relationship entity update
	parentEntity := &models.Entity{
		ID:         parentID,
		Kind:       models.Kind{},
		Created:    "",
		Terminated: "",
		Name:       models.TimeBasedValue{},
		Metadata:   []models.MetadataEntry{},
		Attributes: []models.AttributeEntry{},
		Relationships: []models.RelationshipEntry{
			{
				Key: uniqueRelationshipID,
				Value: models.Relationship{
					RelatedEntityID: childID,
					StartTime:       startTime,
					EndTime:         "",
					ID:              uniqueRelationshipID,
					Name:            relationshipType,
				},
			},
		},
	}

	// Update the parent entity to add the relationship
	_, updateErr := client.UpdateEntity(parentID, parentEntity)
	if updateErr != nil {
		return fmt.Errorf("failed to update parent entity: %w", updateErr)
	}

	return nil
}

// testAPIConnection tests the API connection with retry logic
func testAPIConnection(client *api.Client) error {
	const maxRetries = 5
	const retryDelay = 500 * time.Millisecond

	for attempt := 1; attempt <= maxRetries; attempt++ {
		fmt.Printf("Connection test attempt %d/%d...\n", attempt, maxRetries)

		// Try a simple search to test the connection
		// Search for any document to test the query API
		searchCriteria := &models.SearchCriteria{
			Kind: &models.Kind{
				Major: "Document",
			},
		}

		_, err := client.SearchEntities(searchCriteria)
		if err != nil {
			if attempt == maxRetries {
				return fmt.Errorf("API connection failed after %d attempts: %w", maxRetries, err)
			}
			fmt.Printf("Connection test failed, retrying in %v...\n", retryDelay)
			time.Sleep(retryDelay)
			continue
		}

		// Connection successful
		fmt.Printf("Connection test successful on attempt %d\n", attempt)

		// Add a small delay to ensure the connection is fully established
		time.Sleep(200 * time.Millisecond)
		return nil
	}

	return fmt.Errorf("unexpected error: should not reach this point")
}

func main() {
	// Configuration
	updateURL := "http://localhost:8080/entities"
	queryURL := "http://localhost:8081/v1/entities"

	// Create API client
	client := api.NewClient(updateURL, queryURL)

	// Test connection and wait for API to be ready
	fmt.Println("Testing API connection...")
	err := testAPIConnection(client)
	if err != nil {
		log.Fatalf("Failed to connect to API: %v", err)
	}
	fmt.Println("API connection successful!")

	// Get the directory containing the CSV files
	csvDir := filepath.Join("docs_linking_data")

	// Check if directory exists
	if _, err := os.Stat(csvDir); os.IsNotExist(err) {
		log.Fatalf("Directory %s does not exist", csvDir)
	}

	// Find all CSV files in the directory
	csvFiles, err := filepath.Glob(filepath.Join(csvDir, "*.csv"))
	if err != nil {
		log.Fatalf("Failed to find CSV files: %v", err)
	}

	if len(csvFiles) == 0 {
		log.Fatalf("No CSV files found in directory %s", csvDir)
	}

	fmt.Printf("Found %d CSV files to process\n", len(csvFiles))

	// Process each CSV file
	totalSuccess := 0
	totalErrors := 0

	for _, csvFile := range csvFiles {
		fmt.Printf("\n=== Processing %s ===\n", csvFile)

		err := LinkDocuments(client, csvFile)
		if err != nil {
			log.Printf("Failed to process %s: %v", csvFile, err)
			totalErrors++
		} else {
			totalSuccess++
		}
	}

	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Successfully processed: %d files\n", totalSuccess)
	fmt.Printf("Failed to process: %d files\n", totalErrors)
	fmt.Printf("Total files: %d\n", len(csvFiles))
}
