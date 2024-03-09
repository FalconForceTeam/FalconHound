package input_processor

import (
	"bytes"
	"encoding/json"
	"falconhound/internal"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	auth "github.com/gamepat/azure-oauth2-token"
)

type SentinelConfig struct {
}

type SentinelProcessor struct {
	*InputProcessor
	Config SentinelConfig
}

type SentinelResults struct {
	Tables []struct {
		Name    string `json:"name"`
		Columns []struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"columns"`
		Rows [][]interface{} `json:"rows"`
	} `json:"tables"`
}

func (m *SentinelProcessor) ExecuteQuery() (internal.QueryResults, error) {
	if m.Credentials.SentinelAppSecret == "" {
		return internal.QueryResults{}, fmt.Errorf("SentinelAppSecret is empty, skipping..")
	}

	results, err := LArunQuery(m.Query, m.Credentials)
	if err != nil {
		if strings.Contains(err.Error(), "unexpected HTTP status code: 400") {
			return internal.QueryResults{}, fmt.Errorf("failed to run query %q. Most likely there is a syntax error in the query", m.Query)
		} else {
			return internal.QueryResults{}, fmt.Errorf("failed to run query %q: %w", m.Query, err)
		}
	}

	// Get rows
	var sentinelResults SentinelResults

	err = json.Unmarshal([]byte(results), &sentinelResults)
	if err != nil {
		return internal.QueryResults{}, fmt.Errorf("failed to unmarshal JSON from Sentinel: %v", err)
	}

	// Sentinel data parsing
	rows := sentinelResults.Tables[0].Rows
	columns := sentinelResults.Tables[0].Columns

	queryResults := make(internal.QueryResults, len(rows))
	for i, row := range rows {
		rowMap := make(map[string]interface{})
		for j, column := range columns {
			columnName := column.Name
			rowValue := row[j]
			rowMap[columnName] = rowValue
		}
		queryResults[i] = rowMap
	}

	return queryResults, nil
}

func LArunQuery(query string, creds internal.Credentials) ([]byte, error) {
	url := fmt.Sprintf("https://api.loganalytics.io/v1/workspaces/%s/query/", creds.SentinelWorkspaceID)
	body := map[string]string{
		"query": query,
	}
	jsonBody, err := json.Marshal(body)

	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON request body: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	token, err := getToken(creds)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %v", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("results: %v", resp)
		return nil, fmt.Errorf("unexpected HTTP status code: %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read HTTP response body: %v", err)
	}

	return respBody, nil
}

func getToken(creds internal.Credentials) (string, error) {
	cfg := auth.AuthConfig{
		ClientID:     creds.SentinelAppID,
		ClientSecret: creds.SentinelAppSecret,
		ClientScope:  "https://api.loganalytics.io/.default",
	}

	token, err := auth.RequestAccessToken(creds.SentinelTenantID, cfg)
	if err != nil {
		return "", err
	}
	return token, nil
}
