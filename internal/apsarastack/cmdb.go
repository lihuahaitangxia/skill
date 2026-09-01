package apsarastack

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/zhe-xing/alert-impact-assessment/internal/models"
)

// CMDBClient performs read-only CMDB lineage lookups.
type CMDBClient struct {
	BaseURL    string
	Token      string
	Available  bool
	httpClient *http.Client
}

func NewCMDBClient() *CMDBClient {
	baseURL := strings.TrimRight(os.Getenv("CMDB_API_URL"), "/")
	token := os.Getenv("CMDB_API_TOKEN")
	return &CMDBClient{
		BaseURL:    baseURL,
		Token:      token,
		Available:  baseURL != "" && token != "",
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *CMDBClient) GetResourceLineage(resourceType, resourceID string) models.Lineage {
	if !c.Available {
		return models.Lineage{Available: false, Source: "none"}
	}

	url := fmt.Sprintf("%s/resources/%s/%s/lineage", c.BaseURL, resourceType, resourceID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return models.Lineage{Available: false, Source: "cmdb_error", Error: err.Error()}
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return models.Lineage{Available: false, Source: "cmdb_error", Error: err.Error()}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return models.Lineage{Available: false, Source: "cmdb_error", Error: string(body)}
	}

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return models.Lineage{Available: false, Source: "cmdb_error", Error: err.Error()}
	}

	return models.Lineage{
		Available:      true,
		Source:         "cmdb",
		Instance:       strVal(data["instance"], resourceID),
		Application:    strVal(data["application"], ""),
		BusinessSystem: strVal(data["businessSystem"], ""),
		Customer:       strVal(data["customer"], ""),
		Owner:          strVal(data["owner"], ""),
		OwnerContact:   strVal(data["ownerContact"], ""),
	}
}

func LineageFromTags(tags map[string]string, resourceID string) models.Lineage {
	app := tags["Application"]
	if app == "" {
		app = tags["app"]
	}
	bs := tags["BusinessSystem"]
	if bs == "" {
		bs = tags["business"]
	}
	customer := tags["Customer"]
	if customer == "" {
		customer = tags["customer"]
	}
	owner := tags["Owner"]
	if owner == "" {
		owner = tags["owner"]
	}

	return models.Lineage{
		Available:      true,
		Source:         "tags",
		Instance:       resourceID,
		Application:    app,
		BusinessSystem: bs,
		Customer:       customer,
		Owner:          owner,
		OwnerContact:   tags["OwnerContact"],
	}
}
