// link_citizen_roles.go links citizen nodes to the Minister org-structure node
// that lives inside each minister's Organisation node, via an AS_ROLE relationship.
//
// For each minister (cabinetMinister or stateMinister) the script:
//  1. Fetches all AS_APPOINTED relationships on the minister entity.
//  2. For each appointed citizen, adds an AS_ROLE relationship from the citizen
//     to the minister's org-structure node (<ministerID>_minister), using the
//     same start and end times as the AS_APPOINTED relationship.
//
// Usage:
//
//	go run scripts/link_citizen_roles/main.go \
//	  [-update_endpoint http://localhost:8080/entities] \
//	  [-query_endpoint  http://localhost:8081/v1/entities]
package main

import (
	"flag"
	"fmt"
	"log"

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
	totalLinked := 0
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

			linked, errs, err := processAppointments(client, minister)
			if err != nil {
				log.Printf("  ERROR: %v", err)
				totalErrors++
			} else {
				fmt.Printf("  Linked %d citizen(s) to minister node\n", linked)
				totalLinked += linked
				totalErrors += errs
			}
		}
	}

	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Ministers processed : %d\n", totalProcessed)
	fmt.Printf("AS_ROLE links created: %d\n", totalLinked)
	fmt.Printf("Errors              : %d\n", totalErrors)
}

// processAppointments fetches all AS_APPOINTED relationships on the given
// minister and creates an AS_ROLE relationship from each citizen to the
// minister's org-structure node. Returns the number of relationships created.
func processAppointments(client *api.Client, minister models.SearchResult) (linked int, errors int, err error) {
	appointments, err := client.GetRelatedEntities(minister.ID, &models.Relationship{
		Name: "AS_APPOINTED",
	})
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get AS_APPOINTED relationships for %s: %w", minister.ID, err)
	}

	if len(appointments) == 0 {
		fmt.Printf("  No AS_APPOINTED relationships found — skipping\n")
		return 0, 0, nil
	}

	ministerNodeID := fmt.Sprintf("%s_minister", minister.ID)
	linked = 0
	errors = 0

	for _, rel := range appointments {
		citizenID := rel.RelatedEntityID
		if citizenID == "" {
			log.Printf("  WARNING: AS_APPOINTED relationship has empty RelatedEntityID — skipping")
			continue
		}

		err := linkCitizenToMinisterNode(client, citizenID, ministerNodeID, rel.StartTime, rel.EndTime)
		if err != nil {
			log.Printf("  ERROR linking citizen %s to %s: %v", citizenID, ministerNodeID, err)
			errors++
			continue
		}

		fmt.Printf("  Linked citizen %s → %s (AS_ROLE, %s – %s)\n",
			citizenID, ministerNodeID, rel.StartTime, rel.EndTime)
		linked++
	}

	return linked, errors, nil
}

// linkCitizenToMinisterNode adds an AS_ROLE relationship from the citizen
// to the given minister org-structure node.
func linkCitizenToMinisterNode(client *api.Client, citizenID, ministerNodeID, start, end string) error {
	relID := fmt.Sprintf("%s_%s_%s",
		citizenID,
		ministerNodeID,
		uuid.New().String(),
	)

	_, err := client.UpdateEntity(citizenID, &models.Entity{
		ID:         citizenID,
		Metadata:   []models.MetadataEntry{},
		Attributes: []models.AttributeEntry{},
		Relationships: []models.RelationshipEntry{
			{
				Key: relID,
				Value: models.Relationship{
					RelatedEntityID: ministerNodeID,
					Name:            "AS_ROLE",
					StartTime:       start,
					EndTime:         end,
					ID:              relID,
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to add AS_ROLE relationship: %w", err)
	}

	return nil
}
