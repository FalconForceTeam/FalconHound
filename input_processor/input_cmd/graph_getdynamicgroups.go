package input_cmd

import (
	"context"
	"encoding/json"
	"fmt"
	msgraphsdk "github.com/microsoftgraph/msgraph-beta-sdk-go"
	graphgroups "github.com/microsoftgraph/msgraph-beta-sdk-go/groups"
	"github.com/microsoftgraph/msgraph-beta-sdk-go/models"
	msgraphcore "github.com/microsoftgraph/msgraph-sdk-go-core"
)

type DynamicGroups struct {
	GroupType                     string
	MembershipRule                string
	ObjectId                      string
	MembershipRuleProcessingState string
	DisplayName                   string
	TenantId                      string
}

// requires Directory.Read.All
func GetDynamicGroups(graphClient msgraphsdk.GraphServiceClient) ([]byte, error) {

	requestFilter := "mailEnabled eq false and securityEnabled eq true and membershipRuleProcessingState eq 'On'"
	requestCount := true

	requestParameters := &graphgroups.GroupsRequestBuilderGetQueryParameters{
		Filter: &requestFilter,
		Count:  &requestCount,
		Select: []string{"id", "membershipRule", "membershipRuleProcessingState", "groupTypes", "displayName", "organizationId"},
	}
	configuration := &graphgroups.GroupsRequestBuilderGetRequestConfiguration{
		QueryParameters: requestParameters,
	}

	dynamicGroups, err := graphClient.Groups().Get(context.Background(), configuration)
	if err != nil {
		fmt.Println("Error dynamic groups:", err)
		return nil, err
	}
	settingsperGroup := make([]DynamicGroups, 0)

	pageIterator, err := msgraphcore.NewPageIterator[*models.Group](dynamicGroups, graphClient.GetAdapter(), models.CreateGroupCollectionResponseFromDiscriminatorValue)
	err = pageIterator.Iterate(context.Background(), func(pageItem *models.Group) bool {
		settings := DynamicGroups{}
		// no need for nil check as we are filtering on these fields, also dynamic groups can only have one group type
		settings.GroupType = pageItem.GetGroupTypes()[0]
		settings.MembershipRule = *pageItem.GetMembershipRule()
		settings.ObjectId = *pageItem.GetId()
		settings.MembershipRuleProcessingState = *pageItem.GetMembershipRuleProcessingState()
		settings.DisplayName = *pageItem.GetDisplayName()
		settings.TenantId = *pageItem.GetOrganizationId()
		settingsperGroup = append(settingsperGroup, settings)
		return true
	})
	if err != nil {
		fmt.Println("Error iterating dynamic group requests:", err)
		return nil, err
	}
	json, err := json.MarshalIndent(settingsperGroup, "", "  ")
	if err != nil {
		fmt.Println("Error converting to JSON:", err)
		return nil, err
	}
	return json, nil
}
