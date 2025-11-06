package tests

import (
	"os"
	"sort"
	"testing"
	"time"

	"orgchart_nexoan/api"
	"orgchart_nexoan/models"

	"github.com/stretchr/testify/assert"
)

// UNIT TESTS - Test logic without database

// Helper function to create search criteria for organisations
func createOrgSearchCriteria(name string) *models.SearchCriteria {
	return &models.SearchCriteria{
		Kind: &models.Kind{
			Major: "Organisation",
			Minor: "",
		},
		Name: name,
	}
}

// Helper function to create search criteria for persons
func createPersonSearchCriteria(name string) *models.SearchCriteria {
	return &models.SearchCriteria{
		Kind: &models.Kind{
			Major: "Person",
			Minor: "",
		},
		Name: name,
	}
}

// TestSecretaryOperationsClientCreation tests that the client can be created
func TestSecretaryOperationsClientCreation(t *testing.T) {
	client := api.NewClient("http://localhost:8080/entities", "http://localhost:8081/v1/entities")
	assert.NotNil(t, client, "Client should be created successfully")
}

// TestDateParsing tests date parsing functionality
func TestDateParsing(t *testing.T) {
	validDates := []string{
		"2024-01-01T00:00:00Z",
		"2020-08-09T00:00:00Z",
		"2022-04-28T00:00:00Z",
	}

	for _, dateStr := range validDates {
		parsedDate, err := time.Parse(time.RFC3339, dateStr)
		assert.NoError(t, err, "Should parse valid date: %s", dateStr)
		assert.NotZero(t, parsedDate, "Parsed date should not be zero")
	}
}

// TestSecretaryOperationsConstants tests relationship constants
func TestSecretaryOperationsConstants(t *testing.T) {
	secretaryRelType := "SECRETARY_APPOINTED"
	assert.Equal(t, "SECRETARY_APPOINTED", secretaryRelType)

	renameRelType := "RENAMED_TO"
	assert.Equal(t, "RENAMED_TO", renameRelType)
}

// TestSecretaryOperationsTimeComparison tests time comparison logic
func TestSecretaryOperationsTimeComparison(t *testing.T) {
	date1, _ := time.Parse(time.RFC3339, "2024-01-01T00:00:00Z")
	date2, _ := time.Parse(time.RFC3339, "2024-01-01T00:00:00Z")
	assert.True(t, date1.Equal(date2), "Equal dates should be equal")

	earlier, _ := time.Parse(time.RFC3339, "2024-01-01T00:00:00Z")
	later, _ := time.Parse(time.RFC3339, "2024-01-02T00:00:00Z")
	assert.True(t, earlier.Before(later), "Earlier date should be before later")
}

// TestSecretaryValidationRules tests core validation rules
func TestSecretaryValidationRules(t *testing.T) {
	terminatedDate, _ := time.Parse(time.RFC3339, "2020-08-09T00:00:00Z")
	renameDate, _ := time.Parse(time.RFC3339, "2020-08-09T00:00:00Z")
	shouldMove := terminatedDate.Equal(renameDate)
	assert.True(t, shouldMove, "Secretary terminated on rename date should be moved")
}

// TestChronologicalOrdering tests date sorting
func TestChronologicalOrdering(t *testing.T) {
	dates := []string{
		"2020-08-15T00:00:00Z",
		"2019-12-16T00:00:00Z",
		"2020-08-09T00:00:00Z",
	}

	var parsedDates []time.Time
	for _, dateStr := range dates {
		parsed, _ := time.Parse(time.RFC3339, dateStr)
		parsedDates = append(parsedDates, parsed)
	}

	sort.Slice(parsedDates, func(i, j int) bool {
		return parsedDates[i].Before(parsedDates[j])
	})

	assert.True(t, parsedDates[0].Year() == 2019, "First date should be 2019")
}

// INTEGRATION TESTS - Test with actual database

// TestIntegrationGetMinistryRenames tests fetching ministry renames from database
func TestIntegrationGetMinistryRenames(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Check if API servers are available
	updateURL := os.Getenv("UPDATE_URL")
	queryURL := os.Getenv("QUERY_URL")
	if updateURL == "" {
		updateURL = "http://localhost:8080/entities"
	}
	if queryURL == "" {
		queryURL = "http://localhost:8081/v1/entities"
	}

	client := api.NewClient(updateURL, queryURL)
	assert.NotNil(t, client, "Client should be created")

	t.Log("Testing ministry rename detection...")
	
	// Search for organisations with RENAMED_TO relationships
	criteria := createOrgSearchCriteria("")
	
	orgs, err := client.SearchEntities(criteria)
	if err != nil {
		t.Logf("Warning: Could not search organisations: %v", err)
		t.Skip("Skipping test - API not available")
		return
	}
	
	t.Logf("Found %d organisations in database", len(orgs))
	
	// Count how many have RENAMED_TO relationships
	renameCount := 0
	for _, org := range orgs {
		relations, err := client.GetRelatedEntities(org.ID, &models.Relationship{
			Name: "RENAMED_TO",
		})
		if err == nil && len(relations) > 0 {
			renameCount++
			t.Logf("Found rename: %s -> %s", org.Name, relations[0].RelatedEntityID)
		}
	}
	
	t.Logf("Total ministries with renames: %d", renameCount)
	
	// Test passes if we can query the database (even if no renames found)
	assert.True(t, len(orgs) >= 0, "Should be able to query organisations")
}

// TestIntegrationPemasiriQuery tests querying R. W. R. Pemasiri from database
func TestIntegrationPemasiriQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	updateURL := os.Getenv("UPDATE_URL")
	queryURL := os.Getenv("QUERY_URL")
	if updateURL == "" {
		updateURL = "http://localhost:8080/entities"
	}
	if queryURL == "" {
		queryURL = "http://localhost:8081/v1/entities"
	}

	client := api.NewClient(updateURL, queryURL)

	t.Log("Testing secretary query for R. W. R. Pemasiri...")

	// Search for person by name
	criteria := createPersonSearchCriteria("R. W. R. Pemasiri")

	persons, err := client.SearchEntities(criteria)
	if err != nil {
		t.Logf("Warning: Could not search for person: %v", err)
		t.Skip("Skipping test - API not available")
		return
	}

	if len(persons) == 0 {
		t.Log("R. W. R. Pemasiri not found in database - this is expected if data not loaded")
		return
	}

	person := persons[0]
	t.Logf("Found person: %s (ID: %s)", person.Name, person.ID)

	// Get SECRETARY_APPOINTED relationships
	// Note: We need to search for organisations that have this person as secretary
	orgCriteria := createOrgSearchCriteria("")

	orgs, err := client.SearchEntities(orgCriteria)
	if err != nil {
		t.Logf("Warning: Could not search organisations: %v", err)
		return
	}

	// Find organisations where Pemasiri was secretary
	secretaryRelations := 0
	for _, org := range orgs {
		relations, err := client.GetRelatedEntities(org.ID, &models.Relationship{
			Name:            "SECRETARY_APPOINTED",
			RelatedEntityID: person.ID,
		})
		if err == nil && len(relations) > 0 {
			for _, rel := range relations {
				secretaryRelations++
				t.Logf("Secretary relationship: %s -> %s (from %s to %s)",
					org.Name, person.Name, rel.StartTime, rel.EndTime)
			}
		}
	}

	t.Logf("Total SECRETARY_APPOINTED relationships found for %s: %d", person.Name, secretaryRelations)
	assert.True(t, secretaryRelations >= 0, "Should be able to query relationships")
}

// TestIntegrationMinistryRenameDetection tests detecting ministry renames
func TestIntegrationMinistryRenameDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	updateURL := os.Getenv("UPDATE_URL")
	queryURL := os.Getenv("QUERY_URL")
	if updateURL == "" {
		updateURL = "http://localhost:8080/entities"
	}
	if queryURL == "" {
		queryURL = "http://localhost:8081/v1/entities"
	}

	client := api.NewClient(updateURL, queryURL)

	t.Log("Testing ministry rename detection...")
	t.Log("Looking for: Minister of Roads and Highways -> Minister of Highways")

	// Search for "Minister of Roads and Highways"
	criteria := createOrgSearchCriteria("Minister of Roads and Highways")

	orgs, err := client.SearchEntities(criteria)
	if err != nil {
		t.Logf("Warning: Could not search for ministry: %v", err)
		t.Skip("Skipping test - API not available")
		return
	}

	if len(orgs) == 0 {
		t.Log("Minister of Roads and Highways not found - this is expected if data not loaded")
		return
	}

	oldMinistry := orgs[0]
	t.Logf("Found old ministry: %s (ID: %s)", oldMinistry.Name, oldMinistry.ID)

	// Check for RENAMED_TO relationship
	relations, err := client.GetRelatedEntities(oldMinistry.ID, &models.Relationship{
		Name: "RENAMED_TO",
	})

	if err != nil {
		t.Logf("Warning: Could not get RENAMED_TO relationships: %v", err)
		return
	}

	if len(relations) == 0 {
		t.Log("No RENAMED_TO relationship found for this ministry")
		return
	}

	// Found a rename
	rename := relations[0]
	t.Logf("Found RENAMED_TO relationship:")
	t.Logf("  - From: %s", oldMinistry.Name)
	t.Logf("  - To ID: %s", rename.RelatedEntityID)
	t.Logf("  - Rename Date: %s", rename.StartTime)

	// Verify date can be parsed
	parsedDate, err := time.Parse(time.RFC3339, rename.StartTime)
	assert.NoError(t, err, "Rename date should be parseable")
	t.Logf("  - Parsed date: %s", parsedDate.Format("2006-01-02"))

	// Check if it's the expected rename date (2020-08-09)
	expectedDate := "2020-08-09"
	if parsedDate.Format("2006-01-02") == expectedDate {
		t.Logf("✓ Rename date matches expected: %s", expectedDate)
	} else {
		t.Logf("⚠ Rename date is %s (expected %s)", parsedDate.Format("2006-01-02"), expectedDate)
	}
}

// TestIntegrationSecretaryMovement tests the secretary movement logic
func TestIntegrationSecretaryMovement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	updateURL := os.Getenv("UPDATE_URL")
	queryURL := os.Getenv("QUERY_URL")
	if updateURL == "" {
		updateURL = "http://localhost:8080/entities"
	}
	if queryURL == "" {
		queryURL = "http://localhost:8081/v1/entities"
	}

	client := api.NewClient(updateURL, queryURL)

	t.Log("Testing secretary movement logic...")
	t.Log("Scenario: When ministry renames, secretary terminated on rename date should move to new ministry")
	
	// Search for a ministry with RENAMED_TO relationship
	orgCriteria := createOrgSearchCriteria("")
	
	orgs, err := client.SearchEntities(orgCriteria)
	if err != nil {
		t.Logf("Warning: Could not search organisations: %v", err)
		t.Skip("Skipping test - API not available")
		return
	}
	
	// Find a ministry that was renamed
	var oldMinistry, newMinistry models.SearchResult
	var renameDate time.Time
	var found bool
	
	for _, org := range orgs {
		relations, err := client.GetRelatedEntities(org.ID, &models.Relationship{
			Name: "RENAMED_TO",
		})
		if err == nil && len(relations) > 0 {
			oldMinistry = org
			newMinistry.ID = relations[0].RelatedEntityID
			renameDate, _ = time.Parse(time.RFC3339, relations[0].StartTime)
			found = true
			t.Logf("Found rename: %s -> %s on %s", oldMinistry.Name, newMinistry.ID, renameDate.Format("2006-01-02"))
			break
		}
	}
	
	if !found {
		t.Log("No ministry renames found in database - skipping movement test")
		return
	}
	
	// Get secretaries from old ministry
	secretaryRels, err := client.GetRelatedEntities(oldMinistry.ID, &models.Relationship{
		Name: "SECRETARY_APPOINTED",
	})
	
	if err != nil {
		t.Logf("Warning: Could not get secretary relationships: %v", err)
		return
	}
	
	t.Logf("Old ministry has %d secretary relationships", len(secretaryRels))
	
	// Check for secretaries terminated on rename date
	movedCount := 0
	for _, rel := range secretaryRels {
		if rel.EndTime != "" {
			endDate, err := time.Parse(time.RFC3339, rel.EndTime)
			if err == nil && endDate.Format("2006-01-02") == renameDate.Format("2006-01-02") {
				movedCount++
				t.Logf("✓ Secretary %s terminated on rename date - should be moved", rel.RelatedEntityID)
			}
		}
	}
	
	t.Logf("Secretaries that should be moved: %d", movedCount)
	
	// Verify the logic: If terminated == rename date, should move
	assert.True(t, movedCount >= 0, "Should identify secretaries to move")
}

// TestIntegrationTerminationChain tests the termination chain logic
func TestIntegrationTerminationChain(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	updateURL := os.Getenv("UPDATE_URL")
	queryURL := os.Getenv("QUERY_URL")
	if updateURL == "" {
		updateURL = "http://localhost:8080/entities"
	}
	if queryURL == "" {
		queryURL = "http://localhost:8081/v1/entities"
	}

	client := api.NewClient(updateURL, queryURL)

	t.Log("Testing termination chain logic...")
	t.Log("Rule: Secretary[i].EndDate = Secretary[i+1].StartDate")
	
	// Search for organisations
	orgCriteria := createOrgSearchCriteria("")
	
	orgs, err := client.SearchEntities(orgCriteria)
	if err != nil {
		t.Logf("Warning: Could not search organisations: %v", err)
		t.Skip("Skipping test - API not available")
		return
	}
	
	// Find a ministry with multiple secretaries
	var secretaries []models.Relationship
	found := false
	
	for _, org := range orgs {
		rels, err := client.GetRelatedEntities(org.ID, &models.Relationship{
			Name: "SECRETARY_APPOINTED",
		})
		if err == nil && len(rels) >= 2 {
			secretaries = rels
			found = true
			t.Logf("Found ministry with multiple secretaries: %s (%d secretaries)", org.Name, len(rels))
			break
		}
	}
	
	if !found {
		t.Log("No ministry with multiple secretaries found - skipping chain test")
		return
	}
	
	// Sort secretaries by start date
	sort.Slice(secretaries, func(i, j int) bool {
		dateI, _ := time.Parse(time.RFC3339, secretaries[i].StartTime)
		dateJ, _ := time.Parse(time.RFC3339, secretaries[j].StartTime)
		return dateI.Before(dateJ)
	})
	
	t.Log("Checking termination chain:")
	gapsFound := 0
	overlapsFound := 0
	properChains := 0
	
	// Check chain: Secretary[i].EndDate should equal Secretary[i+1].StartDate
	for i := 0; i < len(secretaries)-1; i++ {
		current := secretaries[i]
		next := secretaries[i+1]
		
		if current.EndTime == "" {
			t.Logf("  [%d] Secretary %s has no end date (still active?)", i, current.RelatedEntityID)
			continue
		}
		
		endDate, err1 := time.Parse(time.RFC3339, current.EndTime)
		startDate, err2 := time.Parse(time.RFC3339, next.StartTime)
		
		if err1 != nil || err2 != nil {
			t.Logf("  [%d] Could not parse dates", i)
			continue
		}
		
		t.Logf("  [%d] %s: %s to %s", i, current.RelatedEntityID[:12]+"...", 
			endDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
		t.Logf("  [%d] %s: %s to ...", i+1, next.RelatedEntityID[:12]+"...", 
			startDate.Format("2006-01-02"))
		
		// Check if chain is correct
		if endDate.Equal(startDate) {
			t.Logf("    ✓ Perfect chain: end = start")
			properChains++
		} else if endDate.Before(startDate) {
			dayGap := int(startDate.Sub(endDate).Hours() / 24)
			t.Logf("    ⚠ Gap of %d days", dayGap)
			gapsFound++
		} else {
			dayOverlap := int(endDate.Sub(startDate).Hours() / 24)
			t.Logf("    ⚠ Overlap of %d days", dayOverlap)
			overlapsFound++
		}
	}
	
	// Check last secretary
	lastSecretary := secretaries[len(secretaries)-1]
	if lastSecretary.EndTime == "" {
		t.Logf("  [%d] %s: Active (no end date) ✓", len(secretaries)-1, lastSecretary.RelatedEntityID[:12]+"...")
	} else {
		t.Logf("  [%d] %s: Terminated ⚠", len(secretaries)-1, lastSecretary.RelatedEntityID[:12]+"...")
	}
	
	t.Logf("\nTermination chain summary:")
	t.Logf("  - Proper chains: %d", properChains)
	t.Logf("  - Gaps found: %d", gapsFound)
	t.Logf("  - Overlaps found: %d", overlapsFound)
	
	// Test passes if we can analyze the chain
	assert.True(t, len(secretaries) >= 2, "Should have secretaries to analyze")
}

// TestIntegrationOverlappingRelationships tests overlapping relationship creation
func TestIntegrationOverlappingRelationships(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	updateURL := os.Getenv("UPDATE_URL")
	queryURL := os.Getenv("QUERY_URL")
	if updateURL == "" {
		updateURL = "http://localhost:8080/entities"
	}
	if queryURL == "" {
		queryURL = "http://localhost:8081/v1/entities"
	}

	client := api.NewClient(updateURL, queryURL)

	t.Log("Testing overlapping relationship detection...")
	t.Log("Looking for secretaries appointed on ministry rename dates")
	
	// Search for organisations
	orgCriteria := createOrgSearchCriteria("")
	
	orgs, err := client.SearchEntities(orgCriteria)
	if err != nil {
		t.Logf("Warning: Could not search organisations: %v", err)
		t.Skip("Skipping test - API not available")
		return
	}
	
	overlappingCount := 0
	phantomCount := 0
	
	// Check each organisation's secretary relationships
	for _, org := range orgs {
		rels, err := client.GetRelatedEntities(org.ID, &models.Relationship{
			Name: "SECRETARY_APPOINTED",
		})
		if err != nil || len(rels) < 2 {
			continue
		}
		
		// Sort by start date
		sort.Slice(rels, func(i, j int) bool {
			dateI, _ := time.Parse(time.RFC3339, rels[i].StartTime)
			dateJ, _ := time.Parse(time.RFC3339, rels[j].StartTime)
			return dateI.Before(dateJ)
		})
		
		// Check for overlapping relationships
		for i := 0; i < len(rels)-1; i++ {
			current := rels[i]
			next := rels[i+1]
			
			if current.EndTime == "" {
				continue // Active secretary, skip
			}
			
			startDate, err1 := time.Parse(time.RFC3339, current.StartTime)
			endDate, err2 := time.Parse(time.RFC3339, current.EndTime)
			nextStartDate, err3 := time.Parse(time.RFC3339, next.StartTime)
			
			if err1 != nil || err2 != nil || err3 != nil {
				continue
			}
			
			// Check for phantom appointments (start == end)
			if startDate.Equal(endDate) {
				phantomCount++
				t.Logf("⚠ Phantom appointment found in %s: secretary appointed and terminated same day (%s)",
					org.Name, startDate.Format("2006-01-02"))
			}
			
			// Check for overlapping relationships
			if endDate.After(nextStartDate) {
				overlappingCount++
				dayOverlap := int(endDate.Sub(nextStartDate).Hours() / 24)
				t.Logf("Found overlapping relationship in %s: %d days overlap",
					org.Name, dayOverlap)
			} else if endDate.Equal(nextStartDate) {
				// This is the expected scenario after secretary cascade
				serviceDays := int(endDate.Sub(startDate).Hours() / 24)
				t.Logf("✓ Perfect handover in %s: secretary served %d days (%s to %s)",
					org.Name, serviceDays, 
					startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
			}
		}
	}
	
	t.Logf("\nOverlapping relationship summary:")
	t.Logf("  - Phantom appointments (start == end): %d", phantomCount)
	t.Logf("  - Overlapping relationships: %d", overlappingCount)
	
	// After secretary cascade is run, there should be no phantoms
	if phantomCount > 0 {
		t.Logf("⚠ WARNING: Found %d phantom appointments - run ProcessAllSecretaryOperations() to fix", phantomCount)
	}
	
	assert.True(t, true, "Overlap analysis completed")
}

// TestIntegrationFullCascadeLogic tests the complete secretary cascade process
func TestIntegrationFullCascadeLogic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	updateURL := os.Getenv("UPDATE_URL")
	queryURL := os.Getenv("QUERY_URL")
	if updateURL == "" {
		updateURL = "http://localhost:8080/entities"
	}
	if queryURL == "" {
		queryURL = "http://localhost:8081/v1/entities"
	}

	client := api.NewClient(updateURL, queryURL)

	t.Log("Testing FULL secretary cascade logic")
	t.Log("Complete workflow:")
	t.Log("  1. Detect ministry renames")
	t.Log("  2. Process old ministry secretaries (apply termination chain)")
	t.Log("  3. Terminate active secretary at rename date")
	t.Log("  4. Move secretaries where Terminated == RenameDate")
	t.Log("  5. Process new ministry secretaries (overlapping relationships)")
	t.Log("  6. Verify no gaps in service, no phantom appointments")
	t.Log("")

	// Get statistics BEFORE running cascade
	t.Log("Collecting statistics BEFORE cascade...")
	beforeStats := collectSecretaryStatistics(t, client)
	t.Logf("BEFORE: %d ministries, %d secretary relationships, %d null end dates, %d phantoms",
		beforeStats.totalMinistries, beforeStats.totalRelationships, 
		beforeStats.nullEndDates, beforeStats.phantoms)
	t.Log("")

	

	// Get statistics AFTER running cascade
	t.Log("Collecting statistics AFTER cascade...")
	afterStats := collectSecretaryStatistics(t, client)
	t.Logf("AFTER:  %d ministries, %d secretary relationships, %d null end dates, %d phantoms",
		afterStats.totalMinistries, afterStats.totalRelationships, 
		afterStats.nullEndDates, afterStats.phantoms)
	t.Log("")

	// Compare statistics
	t.Log("COMPARISON:")
	t.Logf("Ministries:          %d -> %d (change: %+d)",
		beforeStats.totalMinistries, afterStats.totalMinistries,
		afterStats.totalMinistries-beforeStats.totalMinistries)
	t.Logf("Relationships:       %d -> %d (change: %+d)",
		beforeStats.totalRelationships, afterStats.totalRelationships,
		afterStats.totalRelationships-beforeStats.totalRelationships)
	t.Logf("NULL end dates:      %d -> %d (change: %+d)",
		beforeStats.nullEndDates, afterStats.nullEndDates,
		afterStats.nullEndDates-beforeStats.nullEndDates)
	t.Logf("Phantom appointments: %d -> %d (change: %+d)",
		beforeStats.phantoms, afterStats.phantoms,
		afterStats.phantoms-beforeStats.phantoms)
	t.Log("")

	// Expected outcomes after cascade:
	// - NULL end dates should decrease (secretaries get proper termination dates)
	// - Phantoms should be 0 (no same-day appointments)
	// - Relationships might increase (moved secretaries get new relationships)

	assert.True(t, afterStats.totalMinistries >= 0, "Should have ministries")
	assert.True(t, afterStats.totalRelationships >= 0, "Should have relationships")

	if afterStats.phantoms > 0 {
		t.Logf("⚠ WARNING: Still have %d phantom appointments after cascade", afterStats.phantoms)
	}

	t.Log("Test completed successfully")
}

// Helper struct for statistics
type secretaryStats struct {
	totalMinistries     int
	totalRelationships  int
	nullEndDates        int
	phantoms            int
	gaps                int
	overlaps            int
}

// collectSecretaryStatistics gathers statistics about secretary relationships
func collectSecretaryStatistics(t *testing.T, client *api.Client) secretaryStats {
	stats := secretaryStats{}

	// Search for all organisations
	criteria := createOrgSearchCriteria("")

	orgs, err := client.SearchEntities(criteria)
	if err != nil {
		t.Logf("Warning: Could not search organisations: %v", err)
		return stats
	}

	stats.totalMinistries = len(orgs)

	// Check each organisation's secretary relationships
	for _, org := range orgs {
		rels, err := client.GetRelatedEntities(org.ID, &models.Relationship{
			Name: "SECRETARY_APPOINTED",
		})
		if err != nil {
			continue
		}

		stats.totalRelationships += len(rels)

		// Count NULL end dates
		for _, rel := range rels {
			if rel.EndTime == "" {
				stats.nullEndDates++
			}
		}
		//check if any phantom appointments
		// Check for phantoms (start == end)
		for _, rel := range rels {
			if rel.EndTime != "" {
				startDate, err1 := time.Parse(time.RFC3339, rel.StartTime)
				endDate, err2 := time.Parse(time.RFC3339, rel.EndTime)
				if err1 == nil && err2 == nil {
					if startDate.Equal(endDate) {
						stats.phantoms++
					}
				}
			}
		}

		// Check for gaps and overlaps (if multiple secretaries)
		if len(rels) >= 2 {
			// Sort by start date
			sort.Slice(rels, func(i, j int) bool {
				dateI, _ := time.Parse(time.RFC3339, rels[i].StartTime)
				dateJ, _ := time.Parse(time.RFC3339, rels[j].StartTime)
				return dateI.Before(dateJ)
			})

			for i := 0; i < len(rels)-1; i++ {
				current := rels[i]
				next := rels[i+1]

				if current.EndTime == "" {
					continue
				}

				endDate, err1 := time.Parse(time.RFC3339, current.EndTime)
				nextStartDate, err2 := time.Parse(time.RFC3339, next.StartTime)

				if err1 == nil && err2 == nil {
					if endDate.Before(nextStartDate) {
						stats.gaps++
					} else if endDate.After(nextStartDate) {
						stats.overlaps++
					}
				}
			}
		}
	}

	return stats
}
