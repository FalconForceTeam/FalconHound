package input_processor

import (
	"encoding/json"
	"falconhound/input_processor/input_cmd"
	"falconhound/internal"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	msgraphsdk "github.com/microsoftgraph/msgraph-beta-sdk-go"
	"strings"
)

type MsGraphApiConfig struct {
}

type MsGraphApiProcessor struct {
	*InputProcessor
	Config MsGraphApiConfig
}

type MSGraphApiSession struct {
	initialized bool
	client      msgraphsdk.GraphServiceClient
}

var _MSGraphApiSession MSGraphApiSession

type MsGraphApiResults struct {
	Schema []struct {
		Context string `json:"@odata.context"`
	} `json:"Schema"`
	Results internal.QueryResults `json:"Value"`
}

func (m *MsGraphApiProcessor) ExecuteQuery() (internal.QueryResults, error) {
	if m.Credentials.GraphAppSecret == "" {
		return internal.QueryResults{}, fmt.Errorf("GraphAppSecret is empty, skipping..")
	}

	if !_MSGraphApiSession.initialized {
		_MSGraphApiSession.client = graphClient(m.Credentials)
		_MSGraphApiSession.initialized = true
	}

	var MsGraphApiResults internal.QueryResults
	m.Query = strings.TrimSpace(m.Query)
	switch m.Query {
	case "GetMFA":
		results, err := input_cmd.GetMFA(_MSGraphApiSession.client)
		if err != nil {
			return internal.QueryResults{}, fmt.Errorf("GetMFA failed: %v", err)
		}
		err = json.Unmarshal([]byte(results), &MsGraphApiResults)
		if err != nil {
			return internal.QueryResults{}, fmt.Errorf("failed to unmarshal JSON: %v", err)
		}
		return MsGraphApiResults, nil
	//case "GetUserRisk":
	//	results, err := input_cmd.GetUserRisk(_MSGraphApiSession.client)
	//	if err != nil {
	//		return internal.QueryResults{}, fmt.Errorf("GetUserRisk failed: %v", err)
	//	}
	//	err = json.Unmarshal([]byte(results), &MsGraphApiResults)
	//	if err != nil {
	//		return internal.QueryResults{}, fmt.Errorf("failed to unmarshal JSON: %v", err)
	//	}
	//	return MsGraphApiResults, nil
	case "GetOAuthConsent":
		results, err := input_cmd.GetOAuthConsent(_MSGraphApiSession.client)
		if err != nil {
			return internal.QueryResults{}, fmt.Errorf("GetOAuthConsent failed: %v", err)
		}
		err = json.Unmarshal([]byte(results), &MsGraphApiResults)
		if err != nil {
			return internal.QueryResults{}, fmt.Errorf("failed to unmarshal JSON: %v", err)
		}
		return MsGraphApiResults, nil
	case "GetDynamicGroups":
		results, err := input_cmd.GetDynamicGroups(_MSGraphApiSession.client)
		if err != nil {
			return internal.QueryResults{}, fmt.Errorf("GetDynamicGroups failed: %v", err)
		}
		err = json.Unmarshal([]byte(results), &MsGraphApiResults)
		if err != nil {
			return internal.QueryResults{}, fmt.Errorf("failed to unmarshal JSON: %v", err)

		}
		return MsGraphApiResults, nil
	}

	return nil, fmt.Errorf("Query not found: %s", m.Query)
}

func graphClient(creds internal.Credentials) msgraphsdk.GraphServiceClient {
	cred, err := azidentity.NewClientSecretCredential(creds.GraphTenantID, creds.GraphAppID, creds.GraphAppSecret, nil)
	if err != nil {
		fmt.Println("err")
		panic(err)
	}

	graphClient, err := msgraphsdk.NewGraphServiceClientWithCredentials(cred, []string{"https://graph.microsoft.com/.default"})
	if err != nil {
		fmt.Println("Error creating the client:", err)
		panic(err)
	}

	return *graphClient
}
