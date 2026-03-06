package input_processor

import (
	"context"
	"encoding/json"
	"falconhound/input_processor/input_cmd"
	"falconhound/internal"
	"fmt"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	msgraphsdk "github.com/microsoftgraph/msgraph-beta-sdk-go"
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
	if m.Credentials.GraphAppSecret == "" && (m.Credentials.GraphManagedIdentity == "false" || m.Credentials.GraphManagedIdentity == "") && (m.Credentials.GraphFederatedWorkloadIdentity == "false" || m.Credentials.GraphFederatedWorkloadIdentity == "") {
		return internal.QueryResults{}, fmt.Errorf("GraphAppSecret is empty and no Managed Identity or Federated Workload Identity set, skipping..")
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
	var cred azcore.TokenCredential
	var assertionCredentials azcore.TokenCredential
	var err error

	if creds.GraphManagedIdentity == "true" {
		log.Printf("Using Managed Identity for Graph API")
		cred, err = azidentity.NewManagedIdentityCredential(nil)
		if err != nil {
			fmt.Println("Error creating ManagedIdentityCredential:", err)
			panic(err)
		}
	} else if creds.GraphFederatedWorkloadIdentity == "true" {
		log.Printf("Using Managed Identity to retrieve Federated Workload Identity Assertion Token")
		assertionCredentials, err = azidentity.NewManagedIdentityCredential(nil)
		if err != nil {
			fmt.Println("Error creating ManagedIdentityCredential:", err)
			panic(err)
		}
		getAssertion := func(ctx context.Context) (string, error) {
			tk, err := assertionCredentials.GetToken(ctx, policy.TokenRequestOptions{Scopes: []string{"api://AzureADTokenExchange/.default"}})
			if err != nil {
				return "", err
			}
			return tk.Token, nil
		}
		cred, err = azidentity.NewClientAssertionCredential(creds.GraphTenantID, creds.GraphAppID, getAssertion, nil)
		if err != nil {
			fmt.Println("Error creating ClientAssertionCredential:", err)
			panic(err)
		}
	} else {
		cred, err = azidentity.NewClientSecretCredential(creds.GraphTenantID, creds.GraphAppID, creds.GraphAppSecret, nil)
		if err != nil {
			fmt.Println("Error creating ClientSecretCredential:", err)
			panic(err)
		}
	}

	graphClient, err := msgraphsdk.NewGraphServiceClientWithCredentials(cred, []string{"https://graph.microsoft.com/.default"})
	if err != nil {
		fmt.Println("Error creating the client:", err)
		panic(err)
	}

	return *graphClient
}
