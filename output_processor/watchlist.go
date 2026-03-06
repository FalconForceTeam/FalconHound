package output_processor

import (
	"context"
	"falconhound/internal"
	"fmt"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/securityinsights/armsecurityinsights"
	"github.com/iancoleman/orderedmap"
)

type WatchlistOutputConfig struct {
	WatchlistName string
	DisplayName   string
	SearchKey     string
	Overwrite     bool
}

type WatchlistSession struct {
	initialized bool
	credential  azcore.TokenCredential
}

// Used to persist Watchlist TokenCredential across requests
var _WatchlistSession WatchlistSession

type WatchlistOutputProcessor struct {
	*OutputProcessor
	Config WatchlistOutputConfig
}

func (m *WatchlistOutputProcessor) ProduceOutput(QueryResults internal.QueryResults) error {
	// TODO check the CreateUpdateWatchlist function and properly return errors from there
	CreateUpdateWatchlist(QueryResults, m.Config.WatchlistName, m.Config.DisplayName, m.Config.SearchKey, m.Config.Overwrite, m.Credentials)
	return nil
}

// Watchlist does not require batching, will write all output in one go
func (m *WatchlistOutputProcessor) BatchSize() int {
	return 0
}

func CreateUpdateWatchlist(results internal.QueryResults, WatchlistName string, DisplayName string, SearchKey string, Overwrite bool, creds internal.Credentials) {
	if creds.SentinelAppSecret == "" && (creds.SentinelManagedIdentity == "" || creds.SentinelManagedIdentity == "false") && (creds.SentinelFederatedWorkloadIdentity == "" || creds.SentinelFederatedWorkloadIdentity == "false") {
		log.Fatalf("SentinelAppSecret is empty and no Managed Identity/Federated Workload Identity Enabled, skipping..")
	}

	if !_WatchlistSession.initialized {
		_WatchlistSession.credential = WatchlistTokenCredential(creds)
		_WatchlistSession.initialized = true
	}

	ctx := context.Background()
	clientFactory, err := armsecurityinsights.NewClientFactory(creds.SentinelSubscriptionID, _WatchlistSession.credential, nil)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	if len(results) == 0 {
		return
	}

	log.Println("[>] Creating watchlist", WatchlistName, "with", len(results), "items")

	var keys []string
	if len(results) > 0 {
		for k := range results[0] {
			keys = append(keys, k)
		}
	}
	// Write the header
	var rows [][]string
	rows = append(rows, keys)
	for _, record := range results {
		// Convert the record to a map using the ordered map
		m := make(map[string]interface{})
		for _, k := range keys {
			v, ok := record[k]
			if !ok {
				v = nil
			}
			m[k] = v
		}
		// Convert the map to an ordered map
		orderedMap := orderedmap.New()
		for _, k := range keys {
			v, ok := m[k]
			if !ok {
				v = nil
			}
			orderedMap.Set(k, v)
		}
		// Convert the ordered map to a slice of strings
		var row []string
		for _, k := range keys {
			v, _ := orderedMap.Get(k)
			if k == "Resources" {
				v = fmt.Sprintf("%v", v)
			}
			// Replace commas with semicolons
			vStr := fmt.Sprintf("%v", v)
			vStr = strings.ReplaceAll(vStr, ",", ";")
			row = append(row, vStr)
		}
		rows = append(rows, row)
	}

	var rowsStr string
	for _, row := range rows {
		rowsStr += strings.Join(row, ",") + "\n"
	}
	rowsStr = strings.ReplaceAll(rowsStr, "\n", "\n")
	rowsStr = strings.ReplaceAll(rowsStr, "\"", "'")

	// Check if the watchlist already exists
	skipDelete := false
	listRes, err := clientFactory.NewWatchlistsClient().Get(ctx, creds.SentinelResourceGroup, creds.SentinelWorkspaceName, WatchlistName, nil)

	if err != nil {
		if strings.Contains(err.Error(), "404") {
			log.Println("[*] The watchlist", WatchlistName, "does not exist. Creating it now.")
			skipDelete = true
		} else {
			log.Printf("failed to finish the request: %v", err)
		}
	}
	_ = listRes

	// Delete the watchlist if overwrite is true
	if Overwrite && !skipDelete {
		_, err = clientFactory.NewWatchlistsClient().Delete(ctx, creds.SentinelResourceGroup, creds.SentinelWorkspaceName, WatchlistName, nil)
		if err != nil {
			log.Printf("failed to finish the request: %v", err)
		}
	}
	res, err := clientFactory.NewWatchlistsClient().CreateOrUpdate(ctx, creds.SentinelResourceGroup, creds.SentinelWorkspaceName, WatchlistName, armsecurityinsights.Watchlist{
		Etag: to.Ptr("\"0300bf09-0000-0000-0000-5c37296e0000\""),
		Properties: &armsecurityinsights.WatchlistProperties{
			Description:         to.Ptr("Watchlist from FalconHound"),
			ContentType:         to.Ptr("text/csv"),
			DisplayName:         to.Ptr(DisplayName),
			ItemsSearchKey:      to.Ptr(SearchKey),
			NumberOfLinesToSkip: to.Ptr[int32](0),
			Provider:            to.Ptr("FalconForce"),
			RawContent:          to.Ptr(rowsStr),
			Source:              to.Ptr(armsecurityinsights.SourceLocalFile),
		},
	}, nil)
	if err != nil {
		log.Printf("failed to finish the request: %v", err)
	}
	_ = res
}

func WatchlistTokenCredential(creds internal.Credentials) azcore.TokenCredential {
	var cred azcore.TokenCredential
	var assertionCredentials azcore.TokenCredential
	var err error

	if creds.SentinelManagedIdentity == "true" {
		log.Printf("Using Managed Identity for Sentinel")
		cred, err = azidentity.NewManagedIdentityCredential(nil)
		if err != nil {
			fmt.Println("Error creating ManagedIdentityCredential:", err)
		}
	} else if creds.SentinelFederatedWorkloadIdentity == "true" {
		log.Printf("Using Managed Identity to retrieve Federated Workload Identity Assertion Token for Sentinel")
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
		cred, err = azidentity.NewClientAssertionCredential(creds.SentinelTenantID, creds.SentinelAppID, getAssertion, nil)
		if err != nil {
			fmt.Println("Error creating ClientAssertionCredential:", err)
			panic(err)
		}
	} else {
		log.Printf("Using Client Secret for Sentinel")
		cred, err = azidentity.NewClientSecretCredential(creds.SentinelTenantID, creds.SentinelAppID, creds.SentinelAppSecret, nil)
		if err != nil {
			fmt.Println("Error creating ClientSecretCredential:", err)
		}
	}

	return cred
}
