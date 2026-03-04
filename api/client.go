package api

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"orgchart_nexoan/models"

	"github.com/hashicorp/go-retryablehttp"
)

// Client represents the API client
type Client struct {
	updateURL  string
	queryURL   string
	httpClient *retryablehttp.Client
}

// NewClient creates a new API client with automatic retry logic.
//
// Retry policy:
//   - 400 / 404 responses are not retried
//   - 500 / 502 / 503 / 504 and network errors are retried
func NewClient(updateURL, queryURL string) *Client {
	rc := retryablehttp.NewClient()
	rc.RetryMax = 10
	rc.RetryWaitMin = 1 * time.Second
	rc.RetryWaitMax = 6 * time.Second
	rc.CheckRetry = customCheckRetry
	rc.Logger = nil // silence the default per-request log lines

	return &Client{
		updateURL:  updateURL,
		queryURL:   queryURL,
		httpClient: rc,
	}
}

// CreateEntity creates a new entity
func (c *Client) CreateEntity(entity *models.Entity) (*models.Entity, error) {
	jsonData, err := json.Marshal(entity)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal entity: %w", err)
	}

	resp, err := c.httpClient.Post(c.updateURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create entity: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, httpErrorFromStatus(resp.StatusCode, string(body))
	}

	var createdEntity models.Entity
	if err := json.Unmarshal(body, &createdEntity); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &createdEntity, nil
}

// UpdateEntity updates an existing entity
func (c *Client) UpdateEntity(id string, entity *models.Entity) (*models.Entity, error) {
	jsonData, err := json.Marshal(entity)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal entity: %w", err)
	}

	// URL encode the entity ID to handle special characters like slashes
	encodedID := url.QueryEscape(id)

	req, err := retryablehttp.NewRequest(
		http.MethodPut,
		fmt.Sprintf("%s/%s", c.updateURL, encodedID),
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to update entity: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, httpErrorFromStatus(resp.StatusCode, string(body))
	}

	var updatedEntity models.Entity
	if err := json.Unmarshal(body, &updatedEntity); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &updatedEntity, nil
}

// DeleteEntity deletes an entity
func (c *Client) DeleteEntity(id string) error {
	req, err := retryablehttp.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("%s/%s", c.updateURL, id),
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete entity: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return httpErrorFromStatus(resp.StatusCode, string(body))
	}

	return nil
}

// GetRootEntities gets root entity IDs of a given kind
func (c *Client) GetRootEntities(kind string) ([]string, error) {
	params := url.Values{}
	params.Add("kind", kind)

	resp, err := c.httpClient.Get(fmt.Sprintf("%s/root?%s", c.queryURL, params.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to get root entities: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, httpErrorFromStatus(resp.StatusCode, string(body))
	}

	var response models.RootEntitiesResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Body, nil
}

// SearchEntities searches for entities based on criteria
func (c *Client) SearchEntities(criteria *models.SearchCriteria) ([]models.SearchResult, error) {
	jsonData, err := json.Marshal(criteria)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search criteria: %w", err)
	}

	resp, err := c.httpClient.Post(
		fmt.Sprintf("%s/search", c.queryURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search entities: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, httpErrorFromStatus(resp.StatusCode, string(body))
	}

	var response models.SearchResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Decode the name field for each search result
	for i := range response.Body {
		// The name is already a JSON string containing a protobuf object
		var protobufName struct {
			TypeURL string `json:"typeUrl"`
			Value   string `json:"value"`
		}
		if err := json.Unmarshal([]byte(response.Body[i].Name), &protobufName); err != nil {
			return nil, fmt.Errorf("failed to unmarshal protobuf name: %w", err)
		}

		// Convert hex to string
		decoded, err := hex.DecodeString(protobufName.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to decode hex value: %w", err)
		}
		response.Body[i].Name = string(decoded)
	}

	return response.Body, nil
}

// GetEntityMetadata gets metadata of an entity
func (c *Client) GetEntityMetadata(entityID string) (map[string]interface{}, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/%s/metadata", c.queryURL, entityID))
	if err != nil {
		return nil, fmt.Errorf("failed to get entity metadata: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, httpErrorFromStatus(resp.StatusCode, string(body))
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(body, &metadata); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return metadata, nil
}

// GetEntityAttribute retrieves a specific attribute of an entity
func (c *Client) GetEntityAttribute(entityID, attributeName string, startTime, endTime string) (interface{}, error) {
	reqURL := fmt.Sprintf("%s/%s/attributes/%s", c.queryURL, entityID, attributeName)
	if startTime != "" {
		reqURL += fmt.Sprintf("?startTime=%s", startTime)
		if endTime != "" {
			reqURL += fmt.Sprintf("&endTime=%s", endTime)
		}
	}

	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get entity attribute: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, httpErrorFromStatus(resp.StatusCode, string(body))
	}

	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// GetRelatedEntities gets related entity IDs based on query parameters
func (c *Client) GetRelatedEntities(entityID string, query *models.Relationship) ([]models.Relationship, error) {
	jsonData, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	encodedID := url.QueryEscape(entityID)

	resp, err := c.httpClient.Post(
		fmt.Sprintf("%s/%s/relations", c.queryURL, encodedID),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get related entities: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, httpErrorFromStatus(resp.StatusCode, string(body))
	}

	var relations []models.Relationship
	if err := json.Unmarshal(body, &relations); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return relations, nil
}
