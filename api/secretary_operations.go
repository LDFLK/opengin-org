package api

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"orgchart_nexoan/models"
)

// ProcessAllSecretaryOperations processes secretary appointments after ministry  renames
// This implements the COMPLETE CASCADE LOGIC from SECRETARY_CASCADE_LOGIC.md:
// 1. Detect Ministry Renames
// 2. Process Secretaries in Old Ministry (2a: Get All, 2b: Termination Chain, 2c: Terminate Active)
//sorting the secretaries by appointment date
// 3. Move Secretary to New Ministry  
// 4. Process Secretaries in New Ministry (Apply Termination Chain)
//sorting the secretaries by appointment date
func (c *Client) ProcessAllSecretaryOperations() error {
	log.Println("Processing secretary operations (renames and cascade terminations)...")

	// Detect Ministry Renames
	renames, err := c.getAllMinistryRenames()
	if err != nil {
		return fmt.Errorf("error getting ministry renames: %v", err)
	}

	log.Printf("Found %d ministry renames to process\n", len(renames))

	//Sort renames by date to process 
	sort.Slice(renames, func(i, j int) bool {
		return renames[i].RenameDate.Before(renames[j].RenameDate)
	})
	log.Printf("Processing renames chronologically from %s to %s", 
		renames[0].RenameDate.Format("2006-01-02"), 
		renames[len(renames)-1].RenameDate.Format("2006-01-02"))

	processedCount := 0
	errorCount := 0

	for _, rename := range renames {
		log.Printf("\nProcessing rename: %s -> %s (Date: %v)", rename.OldMinistryName, rename.NewMinistryName, rename.RenameDate)

		// STEP 2:The Process Secretaries in Old Ministry 
		err := c.processOldMinistrySecretaries(rename)
		if err != nil {
			log.Printf("Error processing old ministry secretaries: %v", err)
			errorCount++
			continue
		}

		// STEP 3: Move Secretary to New Ministry  
		movedSecretaries, err := c.moveSecretariesToNewMinistry(rename)
		if err != nil {
			log.Printf("Error moving secretaries to new ministry: %v", err)
			errorCount++
			continue

			
		}

		// STEP 4: Process Secretaries in New Ministry
		if len(movedSecretaries) > 0 {
			err = c.processNewMinistrySecretaries(rename)
			if err != nil {
				log.Printf("Error processing new ministry secretaries: %v", err)
				errorCount++
			} else {
				processedCount += len(movedSecretaries)
				log.Printf("Successfully processed %d secretaries for rename", len(movedSecretaries))
			}
		}
	}
	log.Printf("\n== secretary operation complete ===")
	log.Printf("\n=== COMPLETE CASCADE PROCESSING SUMMARY ===")
	log.Printf("  Ministry renames found: %d", len(renames))
	log.Printf("  Secretary operations completed: %d", processedCount)
	log.Printf("  Errors encountered: %d", errorCount)
	log.Printf("=== Cascade Logic: Terminate from OLD + Appoint to NEW ===")

	//Validate results after processing
	log.Printf("\n=== VALIDATION PHASE ===")
	err = c.validateSecretaryCascadeResults()
	if err != nil {
		log.Printf("Validation warnings: %v", err)
	}

	return nil
}

// MinistryRename represents a ministry rename operation
type MinistryRename struct {
	OldMinistryID   string
	OldMinistryName string
	NewMinistryID   string
	NewMinistryName string
	RenameDate      time.Time
}

// getAllMinistryRenames finds all ministry renames using RENAMED_TO relationships
func (c *Client) getAllMinistryRenames() ([]MinistryRename, error) {
	log.Println("Searching for ministry renames using RENAMED_TO relationships...")

	// Get all ministries to find RENAMED_TO relationships
	ministries, err := c.SearchEntities(&models.SearchCriteria{
		Kind: &models.Kind{
			Major: "Organisation",
			Minor: "minister",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error searching for ministries: %v", err)
	}

	var renames []MinistryRename

	// For each ministry, check if it has RENAMED_TO relationships
	for _, ministry := range ministries {
		// Get relationships for this ministry
		relations, err := c.GetRelatedEntities(ministry.ID, &models.Relationship{
			Name: "RENAMED_TO",
		})
		if err != nil {
			continue // Skip if can't get relationships
		}

		// Process each RENAMED_TO relationship
		for _, relation := range relations {
			// Parse rename date
			renameDate, err := time.Parse(time.RFC3339, relation.StartTime)
			if err != nil {
				if renameDate, err = time.Parse("2006-01-02", relation.StartTime); err != nil {
					continue // Skip if can't parse date
				}
			}

			// Get the new ministry details
			newMinistries, err := c.SearchEntities(&models.SearchCriteria{
				ID: relation.RelatedEntityID,
			})
			if err != nil || len(newMinistries) == 0 {
				continue // Skip if can't find new ministry
			}

			newMinistry := newMinistries[0]

		rename := MinistryRename{
			OldMinistryID:   ministry.ID,
			OldMinistryName: ministry.Name,
			NewMinistryID:   newMinistry.ID,
			NewMinistryName: newMinistry.Name,
			RenameDate:      renameDate,
		}

		renames = append(renames, rename)
		log.Printf("  Found rename: %s -> %s (Date: %s)", 
			rename.OldMinistryName, rename.NewMinistryName, renameDate.Format("2006-01-02"))
		}
	}

	log.Printf("Found %d ministry renames", len(renames))
	return renames, nil
}

// processOldMinistrySecretaries implements Step 2 of cascade logic
func (c *Client) processOldMinistrySecretaries(rename MinistryRename) error {
	log.Printf("  Step 2: Processing secretaries in old ministry: %s", rename.OldMinistryName)

	// Step 2a: Get All Secretaries
	secretaries, err := c.getMinistrySecretaries(rename.OldMinistryID)
	if err != nil {
		return fmt.Errorf("failed to get secretaries for old ministry: %w", err)
	}

	if len(secretaries) == 0 {
		log.Printf("No secretaries found in old ministry")
		return nil
	}

	log.Printf("    Found %d secretary(ies) in old ministry", len(secretaries))

	// Step 2b: Sort and Apply Termination Chain
	err = c.applyTerminationChain(secretaries, rename.OldMinistryID)
	if err != nil {
		return fmt.Errorf("failed to apply termination chain to old ministry: %w", err)
	}

	//Refresh secretaries data after termination chain
	// This ensures we have updated termination dates for Step 2c
	secretaries, err = c.getMinistrySecretaries(rename.OldMinistryID)
	if err != nil {
		return fmt.Errorf("failed to refresh secretaries after termination chain: %w", err)
	}
	log.Printf("    Refreshed secretary data after termination chain")

	// Step 2c: Terminate Active Secretary at Rename Date
	err = c.terminateActiveSecretaryAtDate(secretaries, rename.OldMinistryID, rename.RenameDate)
	if err != nil {
		return fmt.Errorf("failed to terminate active secretary at rename date: %w", err)
	}

	return nil
}

// moveSecretariesToNewMinistry implements Step 3 of cascade logic
//LOGIC: Move secretary to new ministry starting on rename date, then terminate existing secretaries
func (c *Client) moveSecretariesToNewMinistry(rename MinistryRename) ([]string, error) {
	log.Printf("  Step 3: Moving secretaries to new ministry: %s", rename.NewMinistryName)

	// Get secretaries from old ministry
	secretaries, err := c.getMinistrySecretaries(rename.OldMinistryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get secretaries from old ministry: %w", err)
	}

	var movedSecretaries []string

	// LOGIC: Find secretary who was terminated at rename date
	for _, secretary := range secretaries {
		// Check if secretary was terminated at rename date
		if secretary.Terminated == "" {
			log.Printf("Secretary %s has no termination date - skipping move", secretary.ID)
			continue
		}

		endDate, err := time.Parse(time.RFC3339, secretary.Terminated)
		if err != nil {
			if endDate, err = time.Parse("2006-01-02", secretary.Terminated); err != nil {
				log.Printf("Secretary %s: Failed to parse termination date %s - skipping", secretary.ID, secretary.Terminated)
				continue // Skip if can't parse date
			}
		}

		// Secretary termination date MUST equal ministry rename date
		if !endDate.Equal(rename.RenameDate) {
			// INVALID: Secretary was terminated at different date
			// They were NOT active at rename date
			// DO NOT move them
			log.Printf("Secretary %s was terminated at %s, not at rename date %s - skipping move", 
				secretary.ID, endDate.Format("2006-01-02"), rename.RenameDate.Format("2006-01-02"))
			continue
		}

		// VALID: Secretary was terminated at rename date
		// This means they were ACTIVE when ministry was renamed
		
		// ADDITIONAL CHECK: Verify secretary was appointed before rename date
		appointmentDate, err := time.Parse(time.RFC3339, secretary.Created)
		if err != nil {
			if appointmentDate, err = time.Parse("2006-01-02", secretary.Created); err != nil {
				log.Printf("Secretary %s: Failed to parse appointment date %s - skipping", secretary.ID, secretary.Created)
				continue
			}
		}
		
		if !appointmentDate.Before(rename.RenameDate) {
			log.Printf("Secretary %s was appointed on/after rename date (%s) - skipping move", 
				secretary.ID, appointmentDate.Format("2006-01-02"))
			continue
		}

		//  Check if secretary is already in new ministry to prevent duplicates
		existingInNew, err := c.getMinistrySecretaries(rename.NewMinistryID)
		if err == nil {
			alreadyMoved := false
			for _, existing := range existingInNew {
				if existing.ID == secretary.ID {
					// Check if this is the moved relationship (starts on rename date)
					existingApptDate, err := time.Parse(time.RFC3339, existing.Created)
					if err == nil && existingApptDate.Equal(rename.RenameDate) {
						alreadyMoved = true
						log.Printf("Secretary %s already moved to new ministry on %s - skipping duplicate", 
							secretary.ID, rename.RenameDate.Format("2006-01-02"))
						break
					}
				}
			}
			if alreadyMoved {
				continue
			}
		}
		
		// STEP 1: Create relationship in new ministry starting on rename date
		err = c.appointSecretaryToMinistry(secretary.ID, rename.NewMinistryID, rename.RenameDate)
		if err != nil {
			log.Printf("Failed to move secretary %s: %v", secretary.ID, err)
			continue
		}

		movedSecretaries = append(movedSecretaries, secretary.ID)
		log.Printf("Moved secretary %s to new ministry starting on rename date %s", 
			secretary.ID, rename.RenameDate.Format("2006-01-02"))
	}

	log.Printf("    Successfully moved %d secretaries", len(movedSecretaries))
	return movedSecretaries, nil
}

// processNewMinistrySecretaries implements Step 4 of cascade logic
//LOGIC: After moving secretary, create overlapping relationships for continuous service
func (c *Client) processNewMinistrySecretaries(rename MinistryRename) error {
	log.Printf("  Step 4: Processing secretaries in new ministry")

	// Step 4a: Get All Secretaries (including moved ones)
	secretaries, err := c.getMinistrySecretaries(rename.NewMinistryID)
	if err != nil {
		return fmt.Errorf("failed to get secretaries for new ministry: %w", err)
	}

	if len(secretaries) == 0 {
		log.Printf("No secretaries found in new ministry")
		return nil
	}

	log.Printf("    Found %d secretary(ies) in new ministry", len(secretaries))

	// CORRECT LOGIC: Create overlapping relationships for continuous service
	err = c.createOverlappingRelationships(secretaries, rename.NewMinistryID, rename.RenameDate)
	if err != nil {
		return fmt.Errorf("failed to create overlapping relationships: %w", err)
	}

	log.Printf("    Created overlapping relationships for continuous service")
	return nil
}

// createOverlappingRelationships implements the correct logic for continuous service
// Creates overlapping relationships where moved secretary serves from rename date to later date
func (c *Client) createOverlappingRelationships(secretaries []models.SearchResult, ministryID string, renameDate time.Time) error {
	if len(secretaries) == 0 {
		return nil
	}

	// Sort secretaries by appointment date
	sort.Slice(secretaries, func(i, j int) bool {
		dateI, _ := time.Parse(time.RFC3339, secretaries[i].Created)
		dateJ, _ := time.Parse(time.RFC3339, secretaries[j].Created)
		return dateI.Before(dateJ)
	})

	log.Printf("    Sorted %d secretaries by appointment date", len(secretaries))

	// Find the moved secretary (appointed on rename date)
	var movedSecretary *models.SearchResult
	
	for i, secretary := range secretaries {
		appointmentDate, err := time.Parse(time.RFC3339, secretary.Created)
		if err != nil {
			if appointmentDate, err = time.Parse("2006-01-02", secretary.Created); err != nil {
				continue
			}
		}
		
		// Check if this is a moved secretary (appointed on rename date)
		if appointmentDate.Equal(renameDate) {
			movedSecretary = &secretaries[i]
			log.Printf("    Found moved secretary: %s (appointed on rename date %s)", 
				secretary.ID, renameDate.Format("2006-01-02"))
			break
		}
	}

	if movedSecretary == nil {
		log.Printf("    No moved secretary found - applying standard termination chain")
		return c.applyTerminationChain(secretaries, ministryID)
	}

	// Moved secretary serves from rename date to next secretary's appointment date
	// Next secretary serves from their appointment date to NULL (active)
	
	// Find the next secretary after the moved one
	var nextSecretary *models.SearchResult
	for _, secretary := range secretaries {
		appointmentDate, err := time.Parse(time.RFC3339, secretary.Created)
		if err != nil {
			if appointmentDate, err = time.Parse("2006-01-02", secretary.Created); err != nil {
				continue
			}
		}
		
		if appointmentDate.After(renameDate) {
			nextSecretary = &secretary
			break
		}
	}

	if nextSecretary != nil {
		// Create overlapping relationship: moved secretary serves from rename date to next secretary's appointment date
		nextAppointmentDate, err := time.Parse(time.RFC3339, nextSecretary.Created)
		if err != nil {
			if nextAppointmentDate, err = time.Parse("2006-01-02", nextSecretary.Created); err != nil {
				return fmt.Errorf("failed to parse next secretary appointment date: %w", err)
			}
		}

		// Terminate the moved secretary at the next secretary's appointment date
		//Pass renameDate as startDate to target the specific relationship
		err = c.updateSecretaryEndDateWithStart(movedSecretary.ID, ministryID, renameDate, nextAppointmentDate)
		if err != nil {
			log.Printf("Failed to terminate moved secretary %s: %v", movedSecretary.ID, err)
		} else {
			log.Printf("Created overlapping relationship: %s serves from %s to %s", 
				movedSecretary.ID, renameDate.Format("2006-01-02"), nextAppointmentDate.Format("2006-01-02"))
		}

		// Apply termination chain to remaining secretaries (after the moved one)
		// This ensures Secretary 4 → 2021-01-01, Secretary 5 → NULL, etc.
		log.Printf("    Applying termination chain to remaining secretaries after moved one")
		
		// Build list of secretaries after the moved one
		var remainingSecretaries []models.SearchResult
		for _, secretary := range secretaries {
			apptDate, err := time.Parse(time.RFC3339, secretary.Created)
			if err != nil {
				if apptDate, err = time.Parse("2006-01-02", secretary.Created); err != nil {
					continue
				}
			}
			
			// Include secretaries appointed AFTER rename date
			if apptDate.After(renameDate) {
				remainingSecretaries = append(remainingSecretaries, secretary)
			}
		}
		
		// Apply termination chain to remaining secretaries
		if len(remainingSecretaries) > 0 {
			err = c.applyTerminationChain(remainingSecretaries, ministryID)
			if err != nil {
				log.Printf("Warning: Failed to apply termination chain to remaining secretaries: %v", err)
			}
		}
		
	} else {
		// No next secretary - moved secretary remains active
		log.Printf("Moved secretary %s remains active (no next secretary)", movedSecretary.ID)
	}

	return nil
}

// getMinistrySecretaries gets all secretaries for a ministry
func (c *Client) getMinistrySecretaries(ministryID string) ([]models.SearchResult, error) {
	// Get all SECRETARY_APPOINTED relationships for this ministry
	relations, err := c.GetRelatedEntities(ministryID, &models.Relationship{
		Name: "SECRETARY_APPOINTED",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get secretary relationships: %w", err)
	}

	var secretaries []models.SearchResult

	// Get secretary details for each relationship
	for _, relation := range relations {
		secretaryResults, err := c.SearchEntities(&models.SearchCriteria{
			ID: relation.RelatedEntityID,
		})
		if err != nil || len(secretaryResults) == 0 {
			continue // Skip if can't find secretary
		}

		secretary := secretaryResults[0]
		// Add relationship info to secretary
		secretary.Created = relation.StartTime
		secretary.Terminated = relation.EndTime

		secretaries = append(secretaries, secretary)
	}

	return secretaries, nil
}

// applyTerminationChain implements the core termination chain logic
func (c *Client) applyTerminationChain(secretaries []models.SearchResult, ministryID string) error {
	if len(secretaries) <= 1 {
		return nil // No chain needed for 0 or 1 secretary
	}

	// Sort secretaries by appointment date (earliest first)
	sortedSecretaries := c.sortSecretariesByDate(secretaries)

	log.Printf("    Applying termination chain to %d secretaries", len(sortedSecretaries))

	// Apply termination logic: each secretary terminates when next one starts
	for i := 0; i < len(sortedSecretaries)-1; i++ {
		currentSecretary := sortedSecretaries[i]
		nextSecretary := sortedSecretaries[i+1]

		// Parse next secretary's start date
		nextStartDate, err := time.Parse(time.RFC3339, nextSecretary.Created)
		if err != nil {
			if nextStartDate, err = time.Parse("2006-01-02", nextSecretary.Created); err != nil {
				continue // Skip if can't parse date
			}
		}

		// Terminate current secretary when next one starts
		err = c.updateSecretaryEndDate(currentSecretary.ID, ministryID, nextStartDate)
		if err != nil {
			log.Printf("Failed to terminate secretary %s: %v", currentSecretary.ID, err)
			continue
		}

		log.Printf("Terminated secretary %s at %s", currentSecretary.ID, nextStartDate.Format("2006-01-02"))
	}

	// Last secretary remains active (no termination)
	lastSecretary := sortedSecretaries[len(sortedSecretaries)-1]
	log.Printf("Secretary %s remains active", lastSecretary.ID)

	return nil
}

// terminateActiveSecretaryAtDate terminates active secretary at specific date
func (c *Client) terminateActiveSecretaryAtDate(secretaries []models.SearchResult, ministryID string, terminateDate time.Time) error {
	log.Printf("    Terminating active secretary at rename date: %s", terminateDate.Format("2006-01-02"))

	//  Only terminate secretaries who are currently active AND appointed before terminate date
	for _, secretary := range secretaries {
		// Check if secretary is active (no termination date)
		if secretary.Terminated != "" {
			log.Printf("Secretary %s already terminated at %s, skipping", secretary.ID, secretary.Terminated)
			continue // Already terminated by applyTerminationChain
		}

		// Check if appointed before terminate date
		appointmentDate, err := time.Parse(time.RFC3339, secretary.Created)
		if err != nil {
			if appointmentDate, err = time.Parse("2006-01-02", secretary.Created); err != nil {
				log.Printf("Secretary %s: Failed to parse appointment date %s - skipping", secretary.ID, secretary.Created)
				continue // Skip if can't parse date
			}
		}

		// terminate if appointed before rename date
		if appointmentDate.Before(terminateDate) {
			// Terminate this secretary at the rename date (NOT at appointment date)
			err = c.updateSecretaryEndDate(secretary.ID, ministryID, terminateDate)
			if err != nil {
				log.Printf("Failed to terminate active secretary %s: %v", secretary.ID, err)
				continue
			}

			log.Printf("Terminated active secretary %s at rename date %s (appointed %s)", 
				secretary.ID, terminateDate.Format("2006-01-02"), appointmentDate.Format("2006-01-02"))
		} else {
			log.Printf("Secretary %s appointed after rename date (%s), skipping termination", 
				secretary.ID, appointmentDate.Format("2006-01-02"))
		}
	}

	return nil
}

// Helper function to check if secretary was active at given date
func (c *Client) isSecretaryActiveAtDate(secretary models.SearchResult, checkDate time.Time) bool {
	// Parse secretary start time
	startTime, err := time.Parse(time.RFC3339, secretary.Created)
	if err != nil {
		startTime, err = time.Parse("2006-01-02", secretary.Created)
		if err != nil {
			return false
		}
	}

	// Check if secretary started on or before the check date
	if checkDate.Before(startTime) {
		return false
	}

	// If no termination date, secretary is still active
	if secretary.Terminated == "" {
		return true
	}

	// Parse secretary end time
	endTime, err := time.Parse(time.RFC3339, secretary.Terminated)
	if err != nil {
		endTime, err = time.Parse("2006-01-02", secretary.Terminated)
		if err != nil {
			return true // If we can't parse end time, assume still active
		}
	}

	// Secretary is active if check date is before the end time
	return checkDate.Before(endTime) || checkDate.Equal(endTime)
}

// updateSecretaryEndDate terminates a secretary appointment relationship
func (c *Client) updateSecretaryEndDate(secretaryID, ministryID string, endDate time.Time) error {
	// Get the current relationship to find the relationship ID
	relations, err := c.GetRelatedEntities(ministryID, &models.Relationship{
		Name:            "SECRETARY_APPOINTED",
		RelatedEntityID: secretaryID,
	})
	if err != nil {
		return fmt.Errorf("failed to get secretary relationship: %w", err)
	}

	// Find the active relationship (no end time)
	var activeRelation *models.Relationship
	for _, relation := range relations {
		if relation.EndTime == "" {
			activeRelation = &relation
			break
		}
	}

	if activeRelation == nil {
		// check if secretary is already termmin
		for _, relation := range relations {
			if relation.EndTime != "" {
				log.Printf("Secretary %s already terminated at %s, skipping", secretaryID, relation.EndTime)
				return nil // Secretary already terminated, no error
			}
		}
		return fmt.Errorf("no active relationship found for secretary %s in ministry %s", secretaryID, ministryID)
	}

	// Update the relationship to set the end date
	endDateStr := endDate.Format(time.RFC3339)
	updateEntity := &models.Entity{
		ID: ministryID,
		Relationships: []models.RelationshipEntry{
			{
				Key: activeRelation.ID,
				Value: models.Relationship{
					EndTime: endDateStr,
					ID:      activeRelation.ID,
				},
			},
		},
	}

	_, err = c.UpdateEntity(ministryID, updateEntity)
	if err != nil {
		return fmt.Errorf("failed to update secretary relationship end date: %w", err)
	}

	return nil
}

// updateSecretaryEndDateWithStart terminates a specific secretary appointment relationship by start date
// This is used when there are multiple relationships for the same secretary in the same ministry
func (c *Client) updateSecretaryEndDateWithStart(secretaryID, ministryID string, startDate, endDate time.Time) error {
	// Get all relationships for this secretary in this ministry
	relations, err := c.GetRelatedEntities(ministryID, &models.Relationship{
		Name:            "SECRETARY_APPOINTED",
		RelatedEntityID: secretaryID,
	})
	if err != nil {
		return fmt.Errorf("failed to get secretary relationship: %w", err)
	}

	// Find the specific relationship that starts on the given date
	var targetRelation *models.Relationship
	for _, relation := range relations {
		relationStartDate, err := time.Parse(time.RFC3339, relation.StartTime)
		if err != nil {
			if relationStartDate, err = time.Parse("2006-01-02", relation.StartTime); err != nil {
				continue
			}
		}

		// Check if this relationship starts on the target date
		if relationStartDate.Equal(startDate) {
			targetRelation = &relation
			log.Printf("Found target relationship starting on %s for secretary %s", 
				startDate.Format("2006-01-02"), secretaryID)
			break
		}
	}

	if targetRelation == nil {
		return fmt.Errorf("no relationship found for secretary %s in ministry %s starting on %s", 
			secretaryID, ministryID, startDate.Format("2006-01-02"))
	}

	// Update the relationship to set the end date
	endDateStr := endDate.Format(time.RFC3339)
	updateEntity := &models.Entity{
		ID: ministryID,
		Relationships: []models.RelationshipEntry{
			{
				Key: targetRelation.ID,
				Value: models.Relationship{
					EndTime: endDateStr,
					ID:      targetRelation.ID,
				},
			},
		},
	}

	_, err = c.UpdateEntity(ministryID, updateEntity)
	if err != nil {
		return fmt.Errorf("failed to update secretary relationship end date: %w", err)
	}

	log.Printf("Terminated secretary %s relationship (started %s) at %s", 
		secretaryID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	return nil
}


// appointSecretaryToMinistry creates a new secretary appointment relationship
func (c *Client) appointSecretaryToMinistry(secretaryID, ministryID string, startDate time.Time) error {
	// Check if relationship already exists with same start date
	relations, err := c.GetRelatedEntities(ministryID, &models.Relationship{
		Name:            "SECRETARY_APPOINTED",
		RelatedEntityID: secretaryID,
	})
	if err == nil {
		for _, relation := range relations {
			// Check if appointment on same date already exists
			if relation.StartTime != "" {
				existingStartDate, err := time.Parse(time.RFC3339, relation.StartTime)
				if err == nil && existingStartDate.Equal(startDate) {
					log.Printf("DUPLICATE PREVENTED: Secretary %s already has appointment in ministry %s on %s", 
						secretaryID, ministryID, startDate.Format("2006-01-02"))
					return nil // Skip creating duplicate
				}
				
				//Only skip  if secretary is already active BEFORE or ON the startDate
				// If the existing relationship starts AFTER startDate, we should still create the new one
				if relation.EndTime == "" && (existingStartDate.Before(startDate) || existingStartDate.Equal(startDate)) {
					log.Printf("ALREADY ACTIVE: Secretary %s is already active in ministry %s from %s - skipping appointment", 
						secretaryID, ministryID, existingStartDate.Format("2006-01-02"))
					return nil
				}
			}
		}
	}

	// Generate unique relationship ID
	currentTimestamp := strings.ReplaceAll(time.Now().Format(time.RFC3339), ":", "-")
	relationshipID := fmt.Sprintf("sec_%s_%s_%s", ministryID, secretaryID, currentTimestamp)

	// Format the start date as RFC3339
	startDateStr := startDate.Format(time.RFC3339)
	
	// Create the relationship
	relationship := &models.Relationship{
		Name:            "SECRETARY_APPOINTED",
		RelatedEntityID: secretaryID,
		StartTime:       startDateStr,
		ID:              relationshipID,
	}

	// Update ministry entity with the new relationship
	updateEntity := &models.Entity{
		ID: ministryID,
		Relationships: []models.RelationshipEntry{
			{
				Key:   relationshipID,
				Value: *relationship,
			},
		},
	}

	_, err = c.UpdateEntity(ministryID, updateEntity)
	if err != nil {
		return fmt.Errorf("failed to create secretary appointment relationship: %w", err)
	}

	log.Printf("Created secretary appointment: %s -> %s starting %s", secretaryID, ministryID, startDateStr)
	return nil
}

// sortSecretariesByDate sorts secretaries by their appointment date (earliest first)
func (c *Client) sortSecretariesByDate(secretaries []models.SearchResult) []models.SearchResult {
	// Create a copy to avoid modifying the original slice
	sorted := make([]models.SearchResult, len(secretaries))
	copy(sorted, secretaries)
	
	// Sort by Created date (appointment date)
	sort.Slice(sorted, func(i, j int) bool {
		dateI := c.parseSecretaryDate(sorted[i].Created)
		dateJ := c.parseSecretaryDate(sorted[j].Created)
		return dateI.Before(dateJ)
	})
	
	return sorted
}

// validateSecretaryCascadeResults validates the results of secretary cascade processing
func (c *Client) validateSecretaryCascadeResults() error {
	log.Printf("Validating secretary cascade results...")
	
	// Get all secretaries to check for appointments
	secretaries, err := c.SearchEntities(&models.SearchCriteria{
		Kind: &models.Kind{
			Major: "Person",
			Minor: "secretary",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to get secretaries for validation: %w", err)
	}
	
	phantomCount := 0
	overlapCount := 0
	
	for _, secretary := range secretaries {
		// Check for phantom appointments (same-day create/terminate)
		if secretary.Created != "" && secretary.Terminated != "" {
			created, err := time.Parse(time.RFC3339, secretary.Created)
			if err != nil {
				if created, err = time.Parse("2006-01-02", secretary.Created); err != nil {
					continue
				}
			}
			
			terminated, err := time.Parse(time.RFC3339, secretary.Terminated)
			if err != nil {
				if terminated, err = time.Parse("2006-01-02", secretary.Terminated); err != nil {
					continue
				}
			}
			
			if created.Equal(terminated) {
				phantomCount++
				log.Printf("PHANTOM APPOINTMENT DETECTED: Secretary %s (Created: %s, Terminated: %s)", 
					secretary.ID, created.Format("2006-01-02"), terminated.Format("2006-01-02"))
			}
		}
	}
	
	log.Printf("Validation complete: %d phantom appointments, %d overlaps detected", phantomCount, overlapCount)
	
	if phantomCount > 0 {
		return fmt.Errorf("validation failed: %d phantom appointments detected", phantomCount)
	}
	
	return nil
}

// parseSecretaryDate parses secretary appointment date
func (c *Client) parseSecretaryDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Time{} // Return zero time for empty dates
	}
	
	// Try RFC3339 format first
	if parsedTime, err := time.Parse(time.RFC3339, dateStr); err == nil {
		return parsedTime
	}
		if parsedTime, err := time.Parse("2006-01-02", dateStr); err == nil {
		return parsedTime
	}
	
	// If parsing fails, return zero time
	return time.Time{}
}