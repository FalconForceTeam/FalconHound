package input_cmd

import (
	"context"
	"encoding/json"
	"fmt"
	msgraphsdk "github.com/microsoftgraph/msgraph-beta-sdk-go"
	"github.com/microsoftgraph/msgraph-beta-sdk-go/models"
	msgraphcore "github.com/microsoftgraph/msgraph-sdk-go-core"
	"time"
)

type OauthConsent struct {
	ClientId    string
	ConsentType string
	StartTime   time.Time
	ExpiryTime  time.Time
	ResourceId  string
	Scope       string
}

// requires Directory.Read.All
func GetOAuthConsent(graphClient msgraphsdk.GraphServiceClient) ([]byte, error) {
	userRiskDetection, err := graphClient.Oauth2PermissionGrants().Get(context.Background(), nil)
	if err != nil {
		fmt.Println("Error getting consented apps:", err)
		return nil, err
	}
	oauthconsentperAccount := make([]OauthConsent, 0)

	pageIterator, err := msgraphcore.NewPageIterator[*models.OAuth2PermissionGrant](userRiskDetection, graphClient.GetAdapter(), models.CreateOAuth2PermissionGrantCollectionResponseFromDiscriminatorValue)
	err = pageIterator.Iterate(context.Background(), func(pageItem *models.OAuth2PermissionGrant) bool {
		settings := OauthConsent{}
		settings.ClientId = *pageItem.GetClientId()
		settings.ConsentType = *pageItem.GetConsentType()
		settings.StartTime = *pageItem.GetStartTime()
		settings.ExpiryTime = *pageItem.GetExpiryTime()
		settings.ResourceId = *pageItem.GetResourceId()
		settings.Scope = *pageItem.GetScope()
		oauthconsentperAccount = append(oauthconsentperAccount, settings)
		return true
	})
	if err != nil {
		fmt.Println("Error iterating role app consents:", err)
		return nil, err
	}
	json, err := json.MarshalIndent(oauthconsentperAccount, "", "  ")
	if err != nil {
		fmt.Println("Error converting to JSON:", err)
		return nil, err
	}
	return json, nil
}
