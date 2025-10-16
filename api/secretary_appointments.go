package api

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"orgchart_nexoan/models"
)

// ProcessSecretaryAppointments processes secretary appointment relationships from CSV files
func (c *Client) ProcessSecretaryAppointments(dataDir string) error {
	fmt.Println("Processing secretary appointments...")

	// Find all secretary CSV files
	csvFiles, err := findSecretaryCSVFiles(dataDir)
	if err != nil {
		return fmt.Errorf("failed to find secretary CSV files: %w", err)
	}

	fmt.Printf("Found %d secretary CSV files\n", len(csvFiles))

	totalAppointments := 0
	processedAppointments := 0
	failedAppointments := 0

	for _, csvFile := range csvFiles {
		fmt.Printf("\nProcessing: %s\n", filepath.Base(filepath.Dir(csvFile)))

		// Load transactions from CSV
		transactions, err := loadSecretaryTransactions(csvFile)
		if err != nil {
			fmt.Printf("Warning: Failed to load transactions from %s: %v\n", csvFile, err)
			continue
		}

		for _, transaction := range transactions {
			totalAppointments++

			ministerName := transaction["parent"].(string)
			secretaryName := transaction["child"].(string)
			dateStr := transaction["date"].(string)
			transactionID := transaction["transaction_id"].(string)

			// Process this secretary appointment
			err := c.processSecretaryAppointment(ministerName, secretaryName, dateStr, transactionID)
			if err != nil {
				fmt.Printf("  ✗ Failed: %s → %s: %v\n", ministerName, secretaryName, err)
				failedAppointments++
			} else {
				fmt.Printf("  ✓ Created: %s appoints %s\n", ministerName, secretaryName)
				processedAppointments++
			}
		}
	}

	// Print summary
	fmt.Printf("\n" + strings.Repeat("=", 70) + "\n")
	fmt.Printf("Secretary Appointment Processing Summary:\n")
	fmt.Printf("- Total appointments found:   %d\n", totalAppointments)
	fmt.Printf("- Successfully processed:     %d ✓\n", processedAppointments)
	fmt.Printf("- Failed:                     %d ✗\n", failedAppointments)
	fmt.Printf(strings.Repeat("=", 70) + "\n")

	return nil
}

// processSecretaryAppointment processes a single secretary appointment transaction
func (c *Client) processSecretaryAppointment(
	ministerName string,
	secretaryName string,
	dateStr string,
	transactionID string,
) error {
	// Parse the appointment date
	appointmentDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("invalid date format: %w", err)
	}

	// Step 1: Search for minister entity
	ministerResults, err := c.SearchEntities(&models.SearchCriteria{
		Kind: &models.Kind{
			Major: "Organisation",
			Minor: "minister",
		},
		Name: ministerName,
	})
	if err != nil {
		return fmt.Errorf("failed to search for minister: %w", err)
	}

	if len(ministerResults) == 0 {
		return fmt.Errorf("minister '%s' not found", ministerName)
	}

	ministerResult := ministerResults[0]

	// Step 2: Minister found - proceed to create secretary appointment
	// Note: We accept the appointment even if dates don't perfectly align
	// because government restructuring causes date mismatches in historical data

	// Step 3: Search for secretary entity
	secretaryResults, err := c.SearchEntities(&models.SearchCriteria{
		Kind: &models.Kind{
			Major: "Person",
			Minor: "citizen",
		},
		Name: secretaryName,
	})
	if err != nil {
		return fmt.Errorf("failed to search for secretary: %w", err)
	}

	var secretaryID string

	// Step 4: Create secretary if not exists
	if len(secretaryResults) == 0 {
		secretaryEntity := &models.Entity{
			ID: fmt.Sprintf("sec_%d", time.Now().UnixNano()),
			Kind: models.Kind{
				Major: "Person",
				Minor: "citizen",
			},
			Created: appointmentDate.Format(time.RFC3339),
			Name: models.TimeBasedValue{
				StartTime: appointmentDate.Format(time.RFC3339),
				Value:     secretaryName,
			},
		}

		createdEntity, err := c.CreateEntity(secretaryEntity)
		if err != nil {
			return fmt.Errorf("failed to create secretary entity: %w", err)
		}
		secretaryID = createdEntity.ID
	} else {
		secretaryID = secretaryResults[0].ID
	}

	// Step 5: Generate relationship ID
	relationshipID := fmt.Sprintf("%s_%s_%s", transactionID, ministerResult.ID, secretaryID)

	// Step 6: Create relationship
	relationship := &models.Relationship{
		Name:            "SECRETARY_APPOINTED",
		RelatedEntityID: secretaryID,
		StartTime:       appointmentDate.Format(time.RFC3339),
		ID:              relationshipID,
	}

	// Step 7: Update minister entity with the new relationship
	ministerEntity := &models.Entity{
		ID:   ministerResult.ID,
		Kind: ministerResult.Kind,
		Relationships: []models.RelationshipEntry{
			{
				Key:   relationshipID,
				Value: *relationship,
			},
		},
	}

	_, err = c.UpdateEntity(ministerResult.ID, ministerEntity)
	if err != nil {
		return fmt.Errorf("failed to create secretary relationship: %w", err)
	}

	return nil
}

// findSecretaryCSVFiles finds all ADD.csv files in Secretary subdirectories
func findSecretaryCSVFiles(rootDir string) ([]string, error) {
	var csvFiles []string

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && info.Name() == "ADD.csv" && strings.Contains(path, "/Secretary/") {
			csvFiles = append(csvFiles, path)
		}

		return nil
	})

	return csvFiles, err
}

// loadSecretaryTransactions loads secretary appointment transactions from CSV file
func loadSecretaryTransactions(filePath string) ([]map[string]interface{}, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read records: %w", err)
	}

	var transactions []map[string]interface{}

	for _, record := range records {
		transaction := make(map[string]interface{})

		// Map CSV columns to transaction
		for i, value := range record {
			if i < len(header) {
				transaction[header[i]] = value
			}
		}

		// Validate required fields
		if _, ok := transaction["parent"]; !ok {
			continue
		}
		if _, ok := transaction["child"]; !ok {
			continue
		}
		if _, ok := transaction["date"]; !ok {
			continue
		}
		if _, ok := transaction["rel_type"]; !ok {
			continue
		}

		// Only process SECRETARY_APPOINTED relationships
		if transaction["rel_type"] != "SECRETARY_APPOINTED" {
			continue
		}

		transactions = append(transactions, transaction)
	}

	return transactions, nil
}
