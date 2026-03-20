package output_processor

import (
	"context"
	"encoding/json"
	"falconhound/internal"
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs/v2"
)

type EventHubOutputConfig struct {
	BHQuery          string
	QueryName        string
	QueryDescription string
	QueryEventID     string
	EventHubName     string
	BatchSize        int
}

type EventHubSession struct {
	initialized    bool
	EventHubWriter *azeventhubs.ProducerClient
}

// Used to persist EventHub token across requests
var _EventHubSession EventHubSession

type EventHubOutputProcessor struct {
	*OutputProcessor
	Config EventHubOutputConfig
}

func (m *EventHubOutputProcessor) BatchSize() int {
	if m.Config.BatchSize > 0 {
		return m.Config.BatchSize
	}
	return 1
}

func (m *EventHubOutputProcessor) ProduceOutput(QueryResults internal.QueryResults) error {
	if m.Credentials.EventHubAppID == "" && (m.Credentials.EventHubAppSecret == "" || m.Credentials.EventHubFederatedWorkloadIdentity == "false" || m.Credentials.EventHubManagedIdentity == "false") && m.Credentials.EventHubConnectionString == "" {
		return fmt.Errorf("EventHub credentials not set, skipping EventHub output")
	}
	if !_EventHubSession.initialized {
		_EventHubSession.EventHubWriter = EventHubWriter(m, m.Credentials)
		_EventHubSession.initialized = true
	}

	runTime := time.Now().UTC().Format("2006-01-02T15:04:05.0000000Z07:00")

	eventDataItems := make([]map[string]interface{}, 0, len(QueryResults))
	for _, result := range QueryResults {
		resultJSON, err := json.Marshal(result)
		if err != nil {
			log.Fatalf("failed to marshal query result: %s", err)
			return err
		}
		eventDataItems = append(eventDataItems, map[string]interface{}{
			"Name":        m.Config.QueryName,
			"Description": m.Config.QueryDescription,
			"EventID":     m.Config.QueryEventID,
			"BHQuery":     m.Config.BHQuery,
			"Timestamp":   runTime,
			"EventData":   string(resultJSON),
		})
	}

	newBatchOptions := &azeventhubs.EventDataBatchOptions{}
	data, err := json.Marshal(map[string]interface{}{"data": eventDataItems})

	EventHubBatch, err := _EventHubSession.EventHubWriter.NewEventDataBatch(context.TODO(), newBatchOptions)
	if err != nil {
		log.Fatalf("failed to create EventHub batch: %s", err)
		return err
	}

	if err := EventHubBatch.AddEventData(&azeventhubs.EventData{Body: data}, nil); err != nil {
		log.Fatalf("failed to add event to batch: %s", err)
		return err
	}

	if err := _EventHubSession.EventHubWriter.SendEventDataBatch(context.TODO(), EventHubBatch, nil); err != nil {
		log.Fatalf("failed to send EventHub batch: %s", err)
		return err
	}

	return nil
}

func EventHubCredential(creds internal.Credentials) azcore.TokenCredential {
	var cred azcore.TokenCredential
	var assertionCredentials azcore.TokenCredential
	var err error

	if creds.EventHubManagedIdentity == "true" {
		log.Printf("Using Managed Identity for EventHub")
		cred, err = azidentity.NewManagedIdentityCredential(nil)
		if err != nil {
			fmt.Println("Error creating ManagedIdentityCredential:", err)
		}
	} else if creds.EventHubFederatedWorkloadIdentity == "true" {
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
		cred, err = azidentity.NewClientAssertionCredential(creds.EventHubTenantID, creds.EventHubAppID, getAssertion, nil)
		if err != nil {
			fmt.Println("Error creating ClientAssertionCredential:", err)
			panic(err)
		}
	} else {
		cred, err = azidentity.NewClientSecretCredential(creds.EventHubTenantID, creds.EventHubAppID, creds.EventHubAppSecret, nil)
		if err != nil {
			fmt.Println("Error creating ClientSecretCredential:", err)
		}
	}

	// var ctx = context.Background()
	// policy := policy.TokenRequestOptions{Scopes: []string{creds.EventHubClusterURL + "/.default"}}
	// token, err := cred.GetToken(ctx, policy)
	// if err != nil {
	// 	fmt.Println("Error getting token:", err)
	// }
	return cred
}

func EventHubWriter(m *EventHubOutputProcessor, creds internal.Credentials) *azeventhubs.ProducerClient {

	if creds.EventHubConnectionString != "" {
		log.Printf("Using Shared Key for EventHub")
		client, err := azeventhubs.NewProducerClientFromConnectionString(creds.EventHubConnectionString, m.Config.EventHubName, nil)
		if err != nil {
			log.Fatalf("failed to create EventHub producer client with shared key: %s", err)
			return nil
		}
		return client
	} else {
		azureIdentity := EventHubCredential(creds)
		client, err := azeventhubs.NewProducerClient(creds.EventHubHostname, m.Config.EventHubName, azureIdentity, nil)
		if err != nil {
			log.Fatalf("failed to create EventHub producer client: %s", err)
			return nil
		}
		return client
	}

}
