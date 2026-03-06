package output_processor

import (
	"bytes"
	"context"
	"encoding/json"
	"falconhound/internal"
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/ingest"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

type ADXOutputConfig struct {
	BHQuery          string
	QueryName        string
	QueryDescription string
	QueryEventID     string
	Table            string
	BatchSize        int
}

type ADXSession struct {
	initialized bool
	token       string
}

// Used to persist ADX token across requests
var _ADXSession ADXSession

type ADXOutputProcessor struct {
	*OutputProcessor
	Config ADXOutputConfig
}

func (m *ADXOutputProcessor) BatchSize() int {
	if m.Config.BatchSize > 0 {
		return m.Config.BatchSize
	}
	return 1
}

func (m *ADXOutputProcessor) ProduceOutput(QueryResults internal.QueryResults) error {
	if m.Credentials.AdxAppSecret == "" && (m.Credentials.AdxManagedIdentity == "" || m.Credentials.AdxManagedIdentity == "false") && (m.Credentials.AdxFederatedWorkloadIdentity == "" || m.Credentials.AdxFederatedWorkloadIdentity == "false") {
		return fmt.Errorf("ADXAppSecret is empty and no Managed Identity or Federated Workload Identity set, skipping..")
	}

	if !_ADXSession.initialized {
		_ADXSession.token = ADXToken(m.Credentials)
		_ADXSession.initialized = true
	}

	jsonData, err := json.Marshal(QueryResults)
	if err != nil {
		log.Fatalf("failed to marshal data: %s", err)
		return err
	}

	runTime := time.Now().UTC().Format("2006-01-02T15:04:05.0000000Z07:00")
	// Create a data object ADXdata with data from EventData or EnrichmentData, whichever is not nil
	ADXData := map[string]interface{}{
		"Name":        m.Config.QueryName,
		"Description": m.Config.QueryDescription,
		"EventID":     m.Config.QueryEventID,
		"BHQuery":     m.Config.BHQuery,
		"Timestamp":   runTime,
		"EventData":   string(jsonData),
	}

	kustoConnectionStringBuilder := kusto.NewConnectionStringBuilder(m.Credentials.AdxClusterURL)
	kustoConnectionString := kustoConnectionStringBuilder.WithApplicationToken(m.Credentials.AdxAppID, _ADXSession.token)

	client, err := kusto.New(kustoConnectionString)
	if err != nil {
		log.Fatalf("failed to create Kusto client: %s", err)
		return err
	}

	// Create an ingestion client
	ingestor, err := ingest.New(client, m.Credentials.AdxDatabase, m.Config.Table)
	if err != nil {
		log.Fatalf("failed to create ingestor: %s", err)
		return err
	}

	if m.Debug {
		log.Printf("clusterURL: %s", m.Credentials.AdxClusterURL)
		log.Println("Ingesting the following data: ", ADXData)
	}

	// Ingest the data
	data, err := json.Marshal(ADXData)
	ctx := context.Background()
	reader := bytes.NewReader(data)
	_, err = ingestor.FromReader(ctx, reader, ingest.FileFormat(ingest.MultiJSON))
	if err != nil {
		log.Fatalf("failed to create file: %s", err)
		return err
	}

	return client.Close()
}

func ADXToken(creds internal.Credentials) string {
	var cred azcore.TokenCredential
	var assertionCredentials azcore.TokenCredential
	var err error

	if creds.AdxManagedIdentity == "true" {
		log.Printf("Using Managed Identity for ADX")
		cred, err = azidentity.NewManagedIdentityCredential(nil)
		if err != nil {
			fmt.Println("Error creating ManagedIdentityCredential:", err)
		}
	} else if creds.AdxFederatedWorkloadIdentity == "true" {
		log.Printf("Using Managed Identity to retrieve Federated Workload Identity Assertion Token for ADX")
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
		cred, err = azidentity.NewClientAssertionCredential(creds.AdxTenantID, creds.AdxAppID, getAssertion, nil)
		if err != nil {
			fmt.Println("Error creating ClientAssertionCredential:", err)
			panic(err)
		}
	} else {
		log.Printf("Using Client Secret for ADX")
		cred, err = azidentity.NewClientSecretCredential(creds.AdxTenantID, creds.AdxAppID, creds.AdxAppSecret, nil)
		if err != nil {
			fmt.Println("Error creating ClientSecretCredential:", err)
		}
	}

	var ctx = context.Background()
	policy := policy.TokenRequestOptions{Scopes: []string{creds.AdxClusterURL + "/.default"}}
	token, err := cred.GetToken(ctx, policy)
	if err != nil {
		fmt.Println("Error getting token:", err)
	}
	return token.Token
}
