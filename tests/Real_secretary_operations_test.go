package tests

import (
	"fmt"
	"testing"
	"time"

	"orgchart_nexoan/api"
	"orgchart_nexoan/models"
)

// Test configuration
const (
	neoURL  = "bolt://localhost:7687"
	neoUser = "neo4j"
	neoPass = "neo4j123"
)

// Real-world test cases verifying secretary cascade behavior during ministry renames.

// TestSecretaryCascade_DifferentPersonReplacement_Aryasinghe verifies cascade logic
// when a different person replaces the secretary after ministry rename.
// Expected: Original secretary's cascade ends, new secretary starts fresh.
func TestSecretaryCascade_DifferentPersonReplacement_Aryasinghe(t *testing.T) {
	client := setupTestClient(t)

	// Expected relationships for R.P. Aryasinghe
	expectedAryasinghe := []SecretaryRelationship{
		{
			Ministry:  "Minister of Foreign Relations",
			StartDate: parseDate("2019-12-16"),
			EndDate:   parseDate("2020-08-09"),
		},
		{
			Ministry:  "Foreign Minister",
			StartDate: parseDate("2020-08-09"),
			EndDate:   parseDate("2020-08-13"),
		},
	}

	// Expected relationships for Jayanath Colombage
	expectedColombage := []SecretaryRelationship{
		{
			Ministry:  "Foreign Minister",
			StartDate: parseDate("2020-08-13"),
			EndDate:   parseDate("2022-02-23"),
		},
		{
			Ministry:  "Minister of Foreign Affairs",
			StartDate: parseDate("2022-02-23"),
			EndDate:   parseDate("2022-05-30"),
		},
	}
	_ = expectedColombage // Used in verification below

	// Verify Aryasinghe's relationships
	actualAryasinghe := getSecretaryRelationships(t, client, "R. P. Aryasinghe")
	assertRelationshipsMatch(t, "R. P. Aryasinghe", expectedAryasinghe, actualAryasinghe)

	// Verify Colombage's relationships
	actualColombage := getSecretaryRelationships(t, client, "Jayanath Colombage")
	assertRelationshipsMatch(t, "Jayanath Colombage", expectedColombage, actualColombage)
}

// TestSecretaryCascade_DifferentPersonReplacement_Ranjith verifies cascade logic
// when original secretary is replaced by different person after rename.
// Expected: Clean handover between secretaries.
func TestSecretaryCascade_DifferentPersonReplacement_Ranjith(t *testing.T) {
	client := setupTestClient(t)

	// Expected relationships for J.A. Ranjith
	expectedRanjith := []SecretaryRelationship{
		{
			Ministry:  "Minister of Industries and Supply Chain Management",
			StartDate: parseDate("2019-12-16"),
			EndDate:   parseDate("2020-08-09"),
		},
		{
			Ministry:  "Minister of Industries",
			StartDate: parseDate("2020-08-09"),
			EndDate:   parseDate("2020-08-13"),
		},
	}

	// Expected relationships for W.A. Chulananda Perera in Industries
	expectedPerera := []SecretaryRelationship{
		{
			Ministry:  "Minister of Industries",
			StartDate: parseDate("2020-08-13"),
			EndDate:   parseDate("2021-07-07"),
		},
	}

	// Verify Ranjith's relationships
	actualRanjith := getSecretaryRelationships(t, client, "J. A. Ranjith")
	assertRelationshipsMatch(t, "J. A. Ranjith", expectedRanjith, actualRanjith)

	// Verify Perera's relationships in Industries ministry
	actualPereraAll := getSecretaryRelationships(t, client, "W. A. Chulananda Perera")
	actualPereraIndustries := filterByMinistry(actualPereraAll, "Minister of Industries")
	assertRelationshipsMatch(t, "W. A. Chulananda Perera (Industries)", expectedPerera, actualPereraIndustries)
}

// TestSecretaryCascade_SamePersonWithFutureReplacement_Rathnayake verifies cascade logic
// when same person is reappointed and later replaced by different person.
// Expected: Termination chain correctly terminates when future secretary starts.
func TestSecretaryCascade_SamePersonWithFutureReplacement_Rathnayake(t *testing.T) {
	client := setupTestClient(t)

	// Expected relationships for R.M.I. Rathnayake
	expectedRathnayake := []SecretaryRelationship{
		{
			Ministry:  "Minister of Fisheries & Aquatic Resources",
			StartDate: parseDate("2019-12-16"),
			EndDate:   parseDate("2020-08-09"),
		},
		{
			Ministry:  "Minister of Fisheries",
			StartDate: parseDate("2020-08-09"),
			EndDate:   parseDate("2020-08-13"),
		},
		{
			Ministry:  "Minister of Fisheries",
			StartDate: parseDate("2020-08-13"),
			EndDate:   parseDate("2024-01-11"),
		},
	}

	// Verify Rathnayake's relationships
	actualRathnayake := getSecretaryRelationships(t, client, "R. M. I. Rathnayake")
	assertRelationshipsMatch(t, "R. M. I. Rathnayake", expectedRathnayake, actualRathnayake)
}

// TestSecretaryCascade_DoubleRename_ChulanandaPerera verifies cascade logic
// when ministry undergoes multiple sequential renames.
// Expected: Each rename creates proper cascade, all terminated correctly.
func TestSecretaryCascade_DoubleRename_ChulanandaPerera(t *testing.T) {
	client := setupTestClient(t)

	// Expected relationships for W.A. Chulananda Perera
	expectedPerera := []SecretaryRelationship{
		{
			Ministry:  "Minister of Information and Communication Technology",
			StartDate: parseDate("2019-12-16"),
			EndDate:   parseDate("2020-01-22"),
		},
		{
			Ministry:  "Minister of Information and Mass Media",
			StartDate: parseDate("2020-01-22"),
			EndDate:   parseDate("2020-08-09"),
		},
		{
			Ministry:  "Minister of Mass Media",
			StartDate: parseDate("2020-08-09"),
			EndDate:   parseDate("2020-08-13"),
		},
		{
			Ministry:  "Minister of Industries",
			StartDate: parseDate("2020-08-13"),
			EndDate:   parseDate("2021-07-07"),
		},
	}

	// Verify Perera's relationships
	actualPerera := getSecretaryRelationships(t, client, "W. A. Chulananda Perera")
	assertRelationshipsMatch(t, "W. A. Chulananda Perera", expectedPerera, actualPerera)

	// Verify cascade continuity (each end date matches next start date for first 3 relationships)
	for i := 0; i < len(actualPerera)-1 && i < 3; i++ {
		if !actualPerera[i].EndDate.Equal(actualPerera[i+1].StartDate) {
			t.Errorf("W. A. Chulananda Perera cascade break at index %d: rel[%d] ends %s but rel[%d] starts %s",
				i, i, formatDate(actualPerera[i].EndDate), i+1, formatDate(actualPerera[i+1].StartDate))
		}
	}
}

// TestSecretaryCascade_SamePersonNoFutureReplacement_Pemasiri verifies correct behavior
// when same person is reappointed with NO future secretary (Pattern D).
func TestSecretaryCascade_SamePersonNoFutureReplacement_Pemasiri(t *testing.T) {
	client := setupTestClient(t)

	// Expected relationships for R.W.R. Pemasiri
	expectedPemasiri := []SecretaryRelationship{
		{
			Ministry:  "Minister of Roads and Highways",
			StartDate: parseDate("2019-12-16"),
			EndDate:   parseDate("2020-08-09"),
		},
		{
			Ministry:  "Minister of Highways",
			StartDate: parseDate("2020-08-09"),
			EndDate:   parseDate("2020-08-13"),
		},
		{
			Ministry:  "Minister of Highways",
			StartDate: parseDate("2020-08-13"),
			EndDate:   parseDate("2020-08-13"),
		},
	}

	// Get actual relationships
	actualPemasiri := getSecretaryRelationships(t, client, "R. W. R. Pemasiri")

	// Verify relationships match expected
	assertRelationshipsMatch(t, "R. W. R. Pemasiri", expectedPemasiri, actualPemasiri)
}

// TestSecretaryCascade_SamePersonNoFutureReplacement_Ranawaka verifies correct behavior
// when same person is reappointed with NO future secretary (Pattern D).
func TestSecretaryCascade_SamePersonNoFutureReplacement_Ranawaka(t *testing.T) {
	client := setupTestClient(t)

	// Expected relationships for R.A.A.K. Ranawaka
	expectedRanawaka := []SecretaryRelationship{
		{
			Ministry:  "Minister of Lands & Land Development",
			StartDate: parseDate("2019-12-16"),
			EndDate:   parseDate("2020-08-09"),
		},
		{
			Ministry:  "Minister of Lands",
			StartDate: parseDate("2020-08-09"),
			EndDate:   parseDate("2020-08-13"),
		},
		{
			Ministry:  "Minister of Lands",
			StartDate: parseDate("2020-08-13"),
			EndDate:   parseDate("2020-08-13"),
		},
	}
	// Get actual relationships
	actualRanawaka := getSecretaryRelationships(t, client, "R. A. A. K. Ranawaka")

	// Verify relationships match expected
	assertRelationshipsMatch(t, "R. A. A. K. Ranawaka", expectedRanawaka, actualRanawaka)
}

// TestSecretaryCascade_SamePersonNoFutureReplacement_Hettiarachchi verifies correct behavior
// when same person is reappointed with NO future secretary (Pattern D).
func TestSecretaryCascade_SamePersonNoFutureReplacement_Hettiarachchi(t *testing.T) {
	client := setupTestClient(t)

	// Expected relationships for S. Hettiarachchi
	expectedHettiarachchi := []SecretaryRelationship{
		{
			Ministry:  "Minister of Public Administration, Home Affairs, Provincial Councils & Local Government",
			StartDate: parseDate("2019-12-16"),
			EndDate:   parseDate("2020-05-20"),
		},
		{
			Ministry:  "Minister of Tourism and Civil Aviation",
			StartDate: parseDate("2020-05-20"),
			EndDate:   parseDate("2020-08-09"),
		},
		{
			Ministry:  "Minister of Tourism",
			StartDate: parseDate("2020-08-09"),
			EndDate:   parseDate("2020-08-13"),
		},
		{
			Ministry:  "Minister of Tourism",
			StartDate: parseDate("2020-08-13"),
			EndDate:   parseDate("2020-08-13"),
		},
		{
			Ministry:  "Minister of Public Security",
			StartDate: parseDate("2022-05-30"),
			EndDate:   time.Time{}, // NULL
		},
	}

	// Get actual relationships
	actualHettiarachchi := getSecretaryRelationships(t, client, "S. Hettiarachchi")

	// Verify relationships match expected
	assertRelationshipsMatch(t, "S. Hettiarachchi", expectedHettiarachchi, actualHettiarachchi)
}



// SecretaryRelationship represents a secretary appointment relationship
type SecretaryRelationship struct {
	Ministry  string
	StartDate time.Time
	EndDate   time.Time // Zero value means NULL (open relationship)
}

// setupTestClient creates and returns a configured API client
func setupTestClient(t *testing.T) *api.Client {
	updateURL := "http://localhost:8080/entities"
	queryURL := "http://localhost:8081/v1/entities"
	
	client := api.NewClient(updateURL, queryURL)
	return client
}

// Secretary ID mapping - obtained from database queries (2020-08-13 cabinet reshuffle)
var secretaryIDs = map[string]string{
	"R. W. R. Pemasiri":           "sec_1762421601075842317",
	"R. M. I. Rathnayake":         "sec_1762421606144619918",
	"W. A. Chulananda Perera":     "sec_1762421604233982343",
	"R. P. Aryasinghe":            "sec_1762421604941360119",
	"J. A. Ranjith":               "sec_1762421603596901607",
	"R. A. A. K. Ranawaka":        "sec_1762421605464609501",
	"S. Hettiarachchi":            "sec_1762421604734563942",
	"Jayanath Colombage":          "sec_1762421611839002204",
}

// getSecretaryRelationships queries Neo4j for all relationships of a given secretary
func getSecretaryRelationships(t *testing.T, client *api.Client, secretaryName string) []SecretaryRelationship {
	t.Helper() // Mark as helper function for better error reporting
	
	// Get the secretary ID from our known mapping
	secretaryID, exists := secretaryIDs[secretaryName]
	if !exists {
		t.Fatalf("Secretary '%s' not found in ID mapping. Available: %v", secretaryName, getMapKeys(secretaryIDs))
	}
	
	t.Logf("Looking up secretary: %s (ID: %s)", secretaryName, secretaryID)
	
	// Get all incoming SECRETARY_APPOINTED relationships (Organisation->Person)
	// Note: In the database, ministries point TO secretaries, not the reverse
	relations, err := client.GetRelatedEntities(secretaryID, &models.Relationship{
		Name:      "SECRETARY_APPOINTED",
		Direction: "incoming",
	})
	
	if err != nil {
		t.Fatalf("Failed to get relationships for '%s': %v", secretaryName, err)
	}
	
	if len(relations) == 0 {
		t.Logf("WARNING: No SECRETARY_APPOINTED relationships found for '%s'", secretaryName)
		return []SecretaryRelationship{}
	}
	
	t.Logf("Found %d relationships for '%s'", len(relations), secretaryName)
	
	// Convert API relationships to test data structures
	var results []SecretaryRelationship
	for i, rel := range relations {
		// Parse start date
		startDate, err := time.Parse(time.RFC3339, rel.StartTime)
		if err != nil {
			// Try alternative format
			startDate, err = time.Parse("2006-01-02", rel.StartTime)
			if err != nil {
				t.Errorf("Failed to parse start date for relationship %d: %s", i, rel.StartTime)
				continue
			}
		}
		
		// Parse end date (may be empty/null)
		var endDate time.Time
		if rel.EndTime != "" {
			endDate, err = time.Parse(time.RFC3339, rel.EndTime)
			if err != nil {
				// Try alternative format
				endDate, err = time.Parse("2006-01-02", rel.EndTime)
				if err != nil {
					t.Errorf("Failed to parse end date for relationship %d: %s", i, rel.EndTime)
				}
			}
		}
		
		// Get ministry name from the RelatedEntityID
		// The relationship connects person to ministry
		ministryID := rel.RelatedEntityID
		ministryName := getMinistryName(t, client, ministryID)
		
		results = append(results, SecretaryRelationship{
			Ministry:  ministryName,
			StartDate: startDate,
			EndDate:   endDate,
		})
		
		t.Logf("  [%d] %s: %s → %s", i+1, ministryName, 
			formatDate(startDate), formatDateOrNull(endDate))
	}
	
	// Sort by start date for consistent ordering
	sortRelationshipsByStartDate(results)
	
	return results
}

// getMinistryName retrieves ministry name by ID
func getMinistryName(t *testing.T, client *api.Client, ministryID string) string {
	t.Helper()
	
	ministries, err := client.SearchEntities(&models.SearchCriteria{
		ID: ministryID,
		Kind: &models.Kind{
			Major: "organisation",
			Minor: "minister",
		},
	})
	
	if err != nil {
		t.Logf("WARNING: Failed to get ministry details for ID %s: %v", ministryID, err)
		return fmt.Sprintf("Unknown Ministry (ID: %s)", ministryID)
	}
	
	if len(ministries) == 0 {
		t.Logf("WARNING: Ministry not found for ID %s", ministryID)
		return fmt.Sprintf("Unknown Ministry (ID: %s)", ministryID)
	}
	
	return ministries[0].Name
}

// sortRelationshipsByStartDate sorts relationships chronologically
func sortRelationshipsByStartDate(rels []SecretaryRelationship) {
	for i := 0; i < len(rels)-1; i++ {
		for j := i + 1; j < len(rels); j++ {
			if rels[j].StartDate.Before(rels[i].StartDate) {
				rels[i], rels[j] = rels[j], rels[i]
			}
		}
	}
}

// filterByMinistry filters relationships to only include a specific ministry
func filterByMinistry(rels []SecretaryRelationship, ministryName string) []SecretaryRelationship {
	var filtered []SecretaryRelationship
	for _, rel := range rels {
		if rel.Ministry == ministryName {
			filtered = append(filtered, rel)
		}
	}
	return filtered
}

// assertRelationshipsMatch verifies expected relationships match actual
func assertRelationshipsMatch(t *testing.T, secretaryName string, expected, actual []SecretaryRelationship) {
	t.Helper() // Mark as helper for better error reporting
	
	// Check count mismatch first
	if len(expected) != len(actual) {
		t.Errorf("%s: Relationship count mismatch", secretaryName)
		t.Errorf("   Expected: %d relationships", len(expected))
		t.Errorf("   Actual:   %d relationships", len(actual))
		
		// Show detailed comparison table
		t.Logf("║ %-74s ║", "EXPECTED RELATIONSHIPS:")
		t.Log("║ #  ║ Ministry                                      ║ Start     ║ End       ║")
		for i, exp := range expected {
			t.Logf("║ %-2d ║ %-45s ║ %-9s ║ %-9s ║", 
				i+1, truncate(exp.Ministry, 45), 
				formatDate(exp.StartDate), formatDateOrNull(exp.EndDate))
		}		
		t.Logf("║ %-74s ║", "ACTUAL RELATIONSHIPS:")
		t.Log("║ #  ║ Ministry                                      ║ Start     ║ End       ║")
		for i, act := range actual {
			marker := ""
			if i >= len(expected) {
				marker = "  EXTRA"
			}
			t.Logf("║ %-2d ║ %-45s ║ %-9s ║ %-9s ║%s", 
				i+1, truncate(act.Ministry, 45),
				formatDate(act.StartDate), formatDateOrNull(act.EndDate), marker)
		}
		
		return
	}
	
	// Check each relationship in detail
	hasErrors := false
	for i := 0; i < len(expected); i++ {
		exp := expected[i]
		act := actual[i]
		
		// Check ministry name
		if exp.Ministry != act.Ministry {
			if !hasErrors {
				t.Errorf("%s: Relationship mismatches found:", secretaryName)
				hasErrors = true
			}
			t.Errorf("   [%d] Ministry: expected '%s', got '%s'", 
				i+1, exp.Ministry, act.Ministry)
		}
		
		// Check start date
		if !datesEqual(exp.StartDate, act.StartDate) {
			if !hasErrors {
				t.Errorf("%s: Relationship mismatches found:", secretaryName)
				hasErrors = true
			}
			t.Errorf("   [%d] Start date: expected %s, got %s",
				i+1, formatDate(exp.StartDate), formatDate(act.StartDate))
		}
		
		// Check end date
		if !datesEqual(exp.EndDate, act.EndDate) {
			if !hasErrors {
				t.Errorf("%s: Relationship mismatches found:", secretaryName)
				hasErrors = true
			}
			t.Errorf("   [%d] End date: expected %s, got %s",
				i+1, formatDateOrNull(exp.EndDate), formatDateOrNull(act.EndDate))
		}
	}
	
	if !hasErrors {
		t.Logf(" %s: All %d relationships match expected values", secretaryName, len(expected))
	}
}

// truncate truncates a string to max length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// parseDate parses date string in YYYY-MM-DD format
func parseDate(dateStr string) time.Time {
	t, _ := time.Parse("2006-01-02", dateStr)
	return t
}

// datesEqual compares two dates (ignoring time component)
func datesEqual(d1, d2 time.Time) bool {
	// Both zero = equal (both NULL)
	if d1.IsZero() && d2.IsZero() {
		return true
	}
	// One zero, one not = not equal
	if d1.IsZero() || d2.IsZero() {
		return false
	}
	// Compare year, month, day only (ignore time)
	return d1.Year() == d2.Year() && 
	       d1.Month() == d2.Month() && 
	       d1.Day() == d2.Day()
}

// formatDate formats date as YYYY-MM-DD
func formatDate(t time.Time) string {
	if t.IsZero() {
		return "NULL"
	}
	return t.Format("2006-01-02")
}

// formatDateOrNull formats date or returns "NULL" for zero time
func formatDateOrNull(t time.Time) string {
	if t.IsZero() {
		return "NULL"
	}
	return t.Format("2006-01-02")
}

// TestSecretaryCascade_ValidationSummary provides an overall validation summary
// Run this to get a quick overview of all secretary cascade patterns
func TestSecretaryCascade_ValidationSummary(t *testing.T) {
	client := setupTestClient(t)

	
	testCases := []struct {
		name           string
		pattern        string
		expectedRelCount int
	}{
		{"R. P. Aryasinghe", "A", 2},
		{"J. A. Ranjith", "A", 2},
		{"R. M. I. Rathnayake", "B", 3},
		{"W. A. Chulananda Perera", "C", 4},
		{"R. W. R. Pemasiri", "D", 3}, 
		{"R. A. A. K. Ranawaka", "D", 3}, 
		{"S. Hettiarachchi", "D", 5},
	
	passCount := 0
	failCount := 0
	
	for i, tc := range testCases {
		rels := getSecretaryRelationships(t, client, tc.name)
		
		status := "PASS"
		if len(rels) != tc.expectedRelCount {
			status = fmt.Sprintf("FAIL (expected %d rels, got %d)", tc.expectedRelCount, len(rels))
			failCount++
		} else {
			passCount++
		}
		
		t.Logf("[%d] Pattern %s: %-30s  %s", i+1, tc.pattern, tc.name, status)
	}
	
	t.Log("")
	t.Logf("║ Total Test Cases:     %-47d ║", len(testCases))
	t.Logf("║ Passing:              %-47d ║", passCount)
	t.Logf("║ Failing:              %-47d ║", failCount)
	
	if failCount > 0 {
		t.Log("")
		t.Errorf("%d test case(s) failed", failCount)
	} else {
		t.Log("All secretary cascade operations validated successfully")
	}
}

// getMapKeys returns the keys of a string map
func getMapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
