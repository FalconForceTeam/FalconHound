package input_processor

import (
	"bytes"
	"context"
	"encoding/json"
	"falconhound/internal"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

type MDEResults struct {
	Schema []struct {
		Name string `json:"Name"`
		Type string `json:"Type"`
	} `json:"Schema"`
	Results internal.QueryResults `json:"Results"`
}

type MDEConfig struct {
}

type MDEProcessor struct {
	*InputProcessor
	Config MDEConfig
}

type MDESession struct {
	initialized bool
	token       string
}

// Used to persist MDE token across requests
var _MDESession MDESession

func (m *MDEProcessor) ExecuteQuery() (internal.QueryResults, error) {
	if m.Credentials.MDEAppSecret == "" {
		return internal.QueryResults{}, fmt.Errorf("MDEAppSecret is empty, skipping..")
	}

	if !_MDESession.initialized {
		_MDESession.token = MDEToken(m.Credentials)
		_MDESession.initialized = true
	}

	url := "https://api.securitycenter.microsoft.com/api/advancedqueries/run"

	body := map[string]string{
		"Query": m.Query,
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
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", _MDESession.token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("status code: %v. (StatusCode 400 = most likely there is a syntax error in the query)", resp.StatusCode)
		return nil, fmt.Errorf("failed to run query %q: %w", m.Name, err)
	}

	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read HTTP response body: %v", err)
	}
	// Get rows
	var MDEResults MDEResults

	err = json.Unmarshal([]byte(result), &MDEResults)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON data received from MDE: %w", err)
	}
	return MDEResults.Results, nil
}

func MDEToken(creds internal.Credentials) string {
	cred, err := azidentity.NewClientSecretCredential(creds.MDETenantID, creds.MDEAppID, creds.MDEAppSecret, nil)
	if err != nil {
		fmt.Println("err")
	}

	var ctx = context.Background()
	policy := policy.TokenRequestOptions{Scopes: []string{"https://api.securitycenter.microsoft.com/.default"}}
	token, err := cred.GetToken(ctx, policy)
	if err != nil {
		fmt.Println("Error getting token:", err)
	}
	return token.Token
}
