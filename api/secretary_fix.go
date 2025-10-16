package api

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"
)

// SecretaryAppointmentRecord holds secretary appointment data for processing
type SecretaryAppointmentRecord struct {
	SecretaryID   string
	SecretaryName string
	MinistryID    string
	MinistryName  string
	Created       time.Time
	RelID         string
}

// FixSecretaryMoveOnMinistryRename fixes StartTime and EndTime for all secretary appointments
// using the Timeline Compression algorithm:
// 1. Group all appointments by secretary
// 2. Sort each secretary's appointments by ministry creation date
// 3. Assign StartTime = ministry.Created, EndTime = next_ministry.Created
// 4. Handle duplicates by terminating extras
func (c *Client) FixSecretaryMoveOnMinistryRename() error {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("FIXING SECRETARY APPOINTMENT DATES")
	fmt.Println(strings.Repeat("=", 80))

	// Step 1: Fetch all secretary appointments
	fmt.Println("\n[Step 1] Fetching all secretary appointments from database...")
	appointments, err := c.getAllSecretaryAppointments()
	if err != nil {
		return fmt.Errorf("failed to fetch appointments: %w", err)
	}
	fmt.Printf("✓ Fetched %d secretary appointments\n", len(appointments))

	if len(appointments) == 0 {
		fmt.Println("No secretary appointments found to fix")
		return nil
	}

	// Step 2: Group appointments by secretary
	fmt.Println("\n[Step 2] Grouping appointments by secretary...")
	secretaryGroups := groupBySecretary(appointments)
	fmt.Printf("✓ Found %d secretaries with appointments\n", len(secretaryGroups))

	// Step 3: Process each secretary's timeline
	fmt.Println("\n[Step 3] Processing each secretary's timeline...")
	stats := &FixStats{
		TotalRelationships:  0,
		FixedWithDates:      0,
		TerminatedDuplicates: 0,
		Errors:              0,
	}

	for secretaryName, secretaryAppts := range secretaryGroups {
		// Sort appointments by creation date
		sort.Slice(secretaryAppts, func(i, j int) bool {
			return secretaryAppts[i].Created.Before(secretaryAppts[j].Created)
		})

		fmt.Printf("\n  Secretary: %s (%d appointments)\n", secretaryName, len(secretaryAppts))

		// Assign start/end times based on position in timeline
		for i, appt := range secretaryAppts {
			stats.TotalRelationships++

			startTime := appt.Created.Format("2006-01-02T15:04:05Z")
			endTime := ""

			// If not the last appointment, end time is when next appointment starts
			if i < len(secretaryAppts)-1 {
				endTime = secretaryAppts[i+1].Created.Format("2006-01-02T15:04:05Z")
			}

			// Update relationship with dates
			err := c.UpdateSecretaryAppointmentDates(appt.RelID, startTime, endTime)
			if err != nil {
				fmt.Printf("    ✗ Error updating %s: %v\n", appt.MinistryName, err)
				stats.Errors++
			} else {
				if endTime == "" {
					fmt.Printf("    ✓ %s: StartTime=%s (active)\n", appt.MinistryName, startTime)
				} else {
					fmt.Printf("    ✓ %s: %s → %s\n", appt.MinistryName, startTime, endTime)
				}
				stats.FixedWithDates++
			}
		}

		// Handle duplicates for this secretary
		duplicates := findDuplicateAppointments(secretaryAppts)
		if len(duplicates) > 0 {
			fmt.Printf("    [Duplicates] Found %d duplicate relationships for %s\n", len(duplicates), secretaryName)

			for _, dupRels := range duplicates {
				// Keep first, terminate others
				for i := 1; i < len(dupRels); i++ {
					dup := dupRels[i]
					startTime := dup.Created.Format("2006-01-02T15:04:05Z")

					// Set EndTime = StartTime to mark as immediately terminated
					err := c.UpdateSecretaryAppointmentDates(dup.RelID, startTime, startTime)
					if err != nil {
						fmt.Printf("      ✗ Error terminating duplicate %s: %v\n", dup.MinistryName, err)
						stats.Errors++
					} else {
						fmt.Printf("      ✓ Terminated duplicate: %s (rel_id: %s)\n", dup.MinistryName, dup.RelID)
						stats.TerminatedDuplicates++
					}
				}
			}
		}
	}

	// Print summary
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("FIX SUMMARY")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("Total relationships processed:     %d\n", stats.TotalRelationships)
	fmt.Printf("Fixed with start/end times:       %d ✓\n", stats.FixedWithDates)
	fmt.Printf("Terminated duplicates:            %d\n", stats.TerminatedDuplicates)
	fmt.Printf("Errors encountered:               %d ✗\n", stats.Errors)
	fmt.Println(strings.Repeat("=", 80))

	if stats.Errors > 0 {
		return fmt.Errorf("completed with %d errors", stats.Errors)
	}

	fmt.Println("✓ Secretary appointment dates fixed successfully!")
	return nil
}

// getAllSecretaryAppointments fetches all secretary appointments from Neo4j using cypher-shell
func (c *Client) getAllSecretaryAppointments() ([]SecretaryAppointmentRecord, error) {
	cypherQuery := `
		MATCH (m:Organisation)-[r:SECRETARY_APPOINTED]->(s:Person)
		RETURN 
			s.ID as secretary_id,
			s.Name as secretary_name,
			m.Id as ministry_id,
			m.Name as ministry_name,
			m.Created as created,
			r.ID as rel_id
		ORDER BY s.Name, m.Created
	`

	// Execute cypher-shell command
	cmd := exec.Command(
		"docker", "exec", "neo4j", "cypher-shell",
		"-u", "neo4j", "-p", "neo4j123",
		cypherQuery,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute cypher query: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 2 {
		return nil, nil // No results
	}

	var appointments []SecretaryAppointmentRecord

	// Skip header line and process data lines
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		// Parse CSV-like output from cypher-shell
		// Format: "secretary_id", "secretary_name", "ministry_id", "ministry_name", created_date, "rel_id"
		parts := parseCSVLine(line)
		if len(parts) < 6 {
			continue
		}

		secretaryID := strings.Trim(parts[0], "\"")
		secretaryName := strings.Trim(parts[1], "\"")
		ministryID := strings.Trim(parts[2], "\"")
		ministryName := strings.Trim(parts[3], "\"")
		createdStr := strings.Trim(parts[4], "\"")
		relID := strings.Trim(parts[5], "\"")

		// Parse date
		created, err := time.Parse("2006-01-02T15:04:05Z", createdStr)
		if err != nil {
			// Try alternate format
			created, err = time.Parse("2006-01-02", createdStr)
			if err != nil {
				fmt.Printf("Warning: Could not parse date %s: %v\n", createdStr, err)
				continue
			}
		}

		appt := SecretaryAppointmentRecord{
			SecretaryID:   secretaryID,
			SecretaryName: secretaryName,
			MinistryID:    ministryID,
			MinistryName:  ministryName,
			Created:       created,
			RelID:         relID,
		}

		appointments = append(appointments, appt)
	}

	return appointments, nil
}

// groupBySecretary groups appointments by secretary name
func groupBySecretary(appointments []SecretaryAppointmentRecord) map[string][]SecretaryAppointmentRecord {
	groups := make(map[string][]SecretaryAppointmentRecord)

	for _, appt := range appointments {
		groups[appt.SecretaryName] = append(groups[appt.SecretaryName], appt)
	}

	return groups
}

// findDuplicateAppointments finds relationships pointing to the same ministry
// Returns a slice of slices, where each inner slice contains duplicate relationships
func findDuplicateAppointments(appointments []SecretaryAppointmentRecord) [][]SecretaryAppointmentRecord {
	// Group by ministry ID
	ministryGroups := make(map[string][]SecretaryAppointmentRecord)

	for _, appt := range appointments {
		ministryGroups[appt.MinistryID] = append(ministryGroups[appt.MinistryID], appt)
	}

	// Find groups with more than one relationship
	var duplicates [][]SecretaryAppointmentRecord

	for _, rels := range ministryGroups {
		if len(rels) > 1 {
			duplicates = append(duplicates, rels)
		}
	}

	return duplicates
}

// UpdateSecretaryAppointmentDates updates the StartTime and EndTime for a secretary appointment relationship
func (c *Client) UpdateSecretaryAppointmentDates(relationshipID, startTime, endTime string) error {
	query := fmt.Sprintf(`
		MATCH (m:Organisation)-[r:SECRETARY_APPOINTED]->(s:Person)
		WHERE r.ID = "%s"
		SET r.StartTime = "%s"
	`, relationshipID, startTime)

	if endTime != "" {
		query += fmt.Sprintf(`, r.EndTime = "%s"`, endTime)
	}

	// Execute cypher-shell command
	cmd := exec.Command(
		"docker", "exec", "neo4j", "cypher-shell",
		"-u", "neo4j", "-p", "neo4j123",
		query,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to update relationship %s: %v (output: %s)", relationshipID, err, string(output))
	}

	return nil
}

// parseCSVLine parses a CSV line from cypher-shell output
func parseCSVLine(line string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false

	for _, ch := range line {
		switch ch {
		case '"':
			inQuotes = !inQuotes
			current.WriteRune(ch)
		case ',':
			if !inQuotes {
				parts = append(parts, current.String())
				current.Reset()
			} else {
				current.WriteRune(ch)
			}
		default:
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// FixStats holds statistics about the fix operation
type FixStats struct {
	TotalRelationships   int
	FixedWithDates       int
	TerminatedDuplicates int
	Errors               int
}
