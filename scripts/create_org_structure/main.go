// link_minister_orgs.go creates an "Organisation" node for every minister entity
// (cabinetMinister or stateMinister) and links the minister to it via AS_ORGANISATION.
//
// For each minister the script:
//  1. Checks whether the minister already has an AS_ORGANISATION relationship → skip if yes.
//  2. Fetches all AS_MINISTER relationships to derive the time window:
//     - earliestStartTime across all relationships
//     - latestEndTime across all relationships (empty string if any is still open)
//  3. Creates a new Organisation/orgStructure entity named "<minister_name> Organisation".
//  4. Adds an AS_ORGANISATION relationship from the minister to the new org node.
//  5. Creates a Minister/orgStructure node named "Minister" and links it from the org node via OVERSEES.
//  6. Creates a Secretary/orgStructure node named "Secretary" and links it from the Minister node via OVERSEES.
//
// Usage:
//
//	go run scripts/link_minister_orgs/main.go \
//	  [-update_endpoint http://localhost:8080/entities] \
//	  [-query_endpoint  http://localhost:8081/v1/entities]
package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"orgchart_nexoan/api"
	"orgchart_nexoan/models"

	"github.com/google/uuid"
)

// ministerMinorKinds lists the minorKind values that identify a minister.
var ministerMinorKinds = []string{"cabinetMinister", "stateMinister"}

func main() {
	updateEndpoint := flag.String("update_endpoint", "http://localhost:8080/entities", "Endpoint for the Update API")
	queryEndpoint := flag.String("query_endpoint", "http://localhost:8081/v1/entities", "Endpoint for the Query API")
	flag.Parse()

	client := api.NewClient(*updateEndpoint, *queryEndpoint)

	totalProcessed := 0
	totalSkipped := 0
	totalCreated := 0
	totalErrors := 0

	for _, minorKind := range ministerMinorKinds {
		fmt.Printf("\n=== Processing ministers with minorKind=%s ===\n", minorKind)

		ministers, err := client.SearchEntities(&models.SearchCriteria{
			Kind: &models.Kind{
				Major: "Organisation",
				Minor: minorKind,
			},
		})
		if err != nil {
			log.Printf("Failed to search for %s entities: %v", minorKind, err)
			totalErrors++
			continue
		}

		fmt.Printf("Found %d %s entities\n", len(ministers), minorKind)

		for i, minister := range ministers {
			fmt.Printf("\n[%d/%d] Processing minister: %s (ID: %s)\n",
				i+1, len(ministers), minister.Name, minister.ID)
			totalProcessed++

			created, skipped, err := processMinister(client, minister)
			if err != nil {
				log.Printf("  ERROR: %v", err)
				totalErrors++
			} else if skipped {
				fmt.Printf("  SKIPPED: already has AS_ORGANISATION relationship\n")
				totalSkipped++
			} else if created {
				fmt.Printf("  CREATED: AS_ORGANISATION relationship established\n")
				totalCreated++
			}
		}
	}

	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Total processed : %d\n", totalProcessed)
	fmt.Printf("Skipped (exists): %d\n", totalSkipped)
	fmt.Printf("Created         : %d\n", totalCreated)
	fmt.Printf("Errors          : %d\n", totalErrors)
}

// processMinister handles the full flow for a single minister entity.
// Returns (created bool, skipped bool, err error).
func processMinister(client *api.Client, minister models.SearchResult) (created bool, skipped bool, err error) {
	// Step 1: Check whether the minister already has an AS_ORGANISATION relationship.
	existingOrgRels, err := client.GetRelatedEntities(minister.ID, &models.Relationship{
		Name: "AS_ORGANISATION",
	})
	if err != nil {
		return false, false, fmt.Errorf("failed to check existing AS_ORGANISATION: %w", err)
	}
	if len(existingOrgRels) > 0 {
		return false, true, nil
	}

	// Step 2: Fetch all AS_MINISTER relationships and derive the time window.
	asMinisters, err := client.GetRelatedEntities(minister.ID, &models.Relationship{
		Name: "AS_MINISTER",
	})
	if err != nil {
		return false, false, fmt.Errorf("failed to get AS_MINISTER relationships: %w", err)
	}

	earliestStart, latestEnd := deriveTimeWindow(asMinisters)

	if earliestStart == "" {
		return false, false, fmt.Errorf("no AS_MINISTER relationships found — cannot determine start time")
	}

	// Step 3: Create the Organisation node.
	orgName := "Organisation"
	orgID := fmt.Sprintf("%s_org", minister.ID)

	orgEntity := &models.Entity{
		ID: orgID,
		Kind: models.Kind{
			Major: "Organisation",
			Minor: "orgStructure",
		},
		Created:    earliestStart,
		Terminated: latestEnd,
		Name: models.TimeBasedValue{
			StartTime: earliestStart,
			Value:     orgName,
		},
		Metadata:      []models.MetadataEntry{},
		Attributes:    []models.AttributeEntry{},
		Relationships: []models.RelationshipEntry{},
	}

	createdOrg, err := client.CreateEntity(orgEntity)
	if err != nil {
		return false, false, fmt.Errorf("failed to create organisation node '%s': %w", orgName, err)
	}
	fmt.Printf("  Created org node: %s (ID: %s)\n", orgName, createdOrg.ID)

	// Step 4: Add the AS_ORGANISATION relationship from the minister to the org node.
	uniqueRelID := fmt.Sprintf("%s_%s_%s",
		minister.ID,
		createdOrg.ID,
		uuid.New().String(),
	)

	ministerUpdate := &models.Entity{
		ID:         minister.ID,
		Metadata:   []models.MetadataEntry{},
		Attributes: []models.AttributeEntry{},
		Relationships: []models.RelationshipEntry{
			{
				Key: uniqueRelID,
				Value: models.Relationship{
					RelatedEntityID: createdOrg.ID,
					StartTime:       earliestStart,
					EndTime:         latestEnd,
					ID:              uniqueRelID,
					Name:            "AS_ORGANISATION",
				},
			},
		},
	}

	_, err = client.UpdateEntity(minister.ID, ministerUpdate)
	if err != nil {
		return false, false, fmt.Errorf("failed to add AS_ORGANISATION relationship: %w", err)
	}

	// Step 5: Create the Minister node and link it from the org node via OVERSEES.
	createdMinisterNode, err := createMinisterNode(client, minister.ID, createdOrg.ID, earliestStart, latestEnd)
	if err != nil {
		return false, false, fmt.Errorf("failed to create minister node: %w", err)
	}
	fmt.Printf("  Created minister node: Minister (ID: %s)\n", createdMinisterNode.ID)

	// Step 6: Create the Secretary node and link it from the Minister node via OVERSEES.
	if err := createSecretaryNode(client, minister.ID, createdMinisterNode.ID, earliestStart, latestEnd); err != nil {
		return false, false, fmt.Errorf("failed to create secretary node: %w", err)
	}
	fmt.Printf("  Created secretary node: Secretary (ID: %s_secretary)\n", minister.ID)

	return true, false, nil
}

// deriveTimeWindow scans a slice of relationships and returns the earliest
// non-empty StartTime and the latest non-empty EndTime. If any relationship
// has an empty EndTime (still open), the returned latestEnd is also "".
func deriveTimeWindow(rels []models.Relationship) (earliestStart, latestEnd string) {
	hasOpenEnd := false

	for _, rel := range rels {
		// Track earliest start
		if earliestStart == "" || rel.StartTime < earliestStart {
			earliestStart = rel.StartTime
		}

		// Track latest end
		if rel.EndTime == "" {
			hasOpenEnd = true
		} else if !hasOpenEnd && (latestEnd == "" || rel.EndTime > latestEnd) {
			latestEnd = rel.EndTime
		}
	}

	if hasOpenEnd {
		latestEnd = "" // still active overall
	}

	return earliestStart, latestEnd
}

// createMinisterNode creates a Minister/orgStructure node and adds an OVERSEES
// relationship from the given org node to the new minister node.
func createMinisterNode(client *api.Client, ministerID, orgID, start, end string) (*models.Entity, error) {
	nodeID := fmt.Sprintf("%s_minister", ministerID)

	ministerNodeEntity := &models.Entity{
		ID: nodeID,
		Kind: models.Kind{
			Major: "Organisation",
			Minor: "orgStructure",
		},
		Created:    start,
		Terminated: end,
		Name: models.TimeBasedValue{
			StartTime: start,
			Value:     "Minister",
		},
		Metadata:      []models.MetadataEntry{},
		Attributes:    []models.AttributeEntry{},
		Relationships: []models.RelationshipEntry{},
	}

	created, err := client.CreateEntity(ministerNodeEntity)
	if err != nil {
		return nil, fmt.Errorf("failed to create Minister node: %w", err)
	}

	relID := fmt.Sprintf("%s_%s_%s",
		orgID,
		nodeID,
		strings.ReplaceAll(time.Now().Format(time.RFC3339), ":", "-"),
	)

	_, err = client.UpdateEntity(orgID, &models.Entity{
		ID:         orgID,
		Metadata:   []models.MetadataEntry{},
		Attributes: []models.AttributeEntry{},
		Relationships: []models.RelationshipEntry{
			{
				Key: relID,
				Value: models.Relationship{
					RelatedEntityID: nodeID,
					Name:            "OVERSEES",
					StartTime:       start,
					EndTime:         end,
					ID:              relID,
				},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add OVERSEES relationship from org to minister node: %w", err)
	}

	return created, nil
}

// createSecretaryNode creates a Secretary/orgStructure node and adds an OVERSEES
// relationship from the given minister node to the new secretary node.
func createSecretaryNode(client *api.Client, ministerID, ministerNodeID, start, end string) error {
	nodeID := fmt.Sprintf("%s_secretary", ministerID)

	secretaryNodeEntity := &models.Entity{
		ID: nodeID,
		Kind: models.Kind{
			Major: "Organisation",
			Minor: "orgStructure",
		},
		Created:    start,
		Terminated: end,
		Name: models.TimeBasedValue{
			StartTime: start,
			Value:     "Secretary",
		},
		Metadata:      []models.MetadataEntry{},
		Attributes:    []models.AttributeEntry{},
		Relationships: []models.RelationshipEntry{},
	}

	_, err := client.CreateEntity(secretaryNodeEntity)
	if err != nil {
		return fmt.Errorf("failed to create Secretary node: %w", err)
	}

	relID := fmt.Sprintf("%s_%s_%s",
		ministerNodeID,
		nodeID,
		strings.ReplaceAll(time.Now().Format(time.RFC3339), ":", "-"),
	)

	_, err = client.UpdateEntity(ministerNodeID, &models.Entity{
		ID:         ministerNodeID,
		Metadata:   []models.MetadataEntry{},
		Attributes: []models.AttributeEntry{},
		Relationships: []models.RelationshipEntry{
			{
				Key: relID,
				Value: models.Relationship{
					RelatedEntityID: nodeID,
					Name:            "OVERSEES",
					StartTime:       start,
					EndTime:         end,
					ID:              relID,
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to add OVERSEES relationship from minister to secretary node: %w", err)
	}

	return nil
}
