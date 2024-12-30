package input_processor

import (
	"context"
	"encoding/json"
	"falconhound/internal"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

type MSGraphConfig struct {
}

type MSGraphProcessor struct {
	*InputProcessor
	Config MSGraphConfig
}

type MSGraphSession struct {
	initialized bool
	token       string
}

type MSGraphResults struct {
	Schema []struct {
		Context string `json:"@odata.context"`
	} `json:"Schema"`
	Results internal.QueryResults `json:"Value"`
}

var _MSGraphSession MSGraphSession

func (m *MSGraphProcessor) ExecuteQuery() (internal.QueryResults, error) {
	if m.Credentials.GraphAppSecret == "" && (m.Credentials.GraphManagedIdentity == "false" || m.Credentials.GraphManagedIdentity == "") {
		return internal.QueryResults{}, fmt.Errorf("GraphAppSecret is empty, skipping..")
	}

	if !_MSGraphSession.initialized {
		_MSGraphSession.token = graphToken(m.Credentials)
		_MSGraphSession.initialized = true
	}

	// Build request based on query in YAML
	baseURL := "https://graph.microsoft.com/"
	reqUrl := fmt.Sprintf("%s%s", baseURL, m.Query)
	url := strings.TrimSpace(reqUrl)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", _MSGraphSession.token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to run query %q: %w", m.Name, err)
	}

	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read HTTP response body: %v", err)
	}

	// Get rows
	var MSGraphResults MSGraphResults

	err = json.Unmarshal([]byte(result), &MSGraphResults)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON data received from MSGraph: %w", err)
	}

	// add GraphTenantID to each row. BH requires this and it is not returned by MSGraph
	for i := range MSGraphResults.Results {
		MSGraphResults.Results[i]["GraphTenantID"] = m.Credentials.GraphTenantID
	}

	return MSGraphResults.Results, nil
}

func graphToken(creds internal.Credentials) string {
	var cred azcore.TokenCredential
	var err error

	if creds.GraphManagedIdentity == "true" {
		log.Printf("Using Managed Identity for Graph API")
		cred, err = azidentity.NewManagedIdentityCredential(nil)
		if err != nil {
			fmt.Println("Error creating ManagedIdentityCredential:", err)
			panic(err)
		}
	} else {
		cred, err = azidentity.NewClientSecretCredential(creds.GraphTenantID, creds.GraphAppID, creds.GraphAppSecret, nil)
		if err != nil {
			fmt.Println("Error creating ClientSecretCredential:", err)
			panic(err)
		}
	}

	var ctx = context.Background()
	policy := policy.TokenRequestOptions{Scopes: []string{"https://graph.microsoft.com/.default"}}
	token, err := cred.GetToken(ctx, policy)
	if err != nil {
		fmt.Println("Error getting token:", err)
		panic(err)
	}
	return token.Token
}
