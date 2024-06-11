package input_processor

import (
	"encoding/json"
	"falconhound/internal"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Role struct {
	RoleId           string `json:"RoleId"`
	RoleName         string `json:"RoleName"`
	IsPrivilegedRole bool   `json:"isPrivileged"`
	Classification   struct {
		EAMTierLevelTagValue string `json:"EAMTierLevelTagValue,omitempty"`
		EAMTierLevelName     string `json:"EAMTierLevelName,omitempty"`
	} `json:"Classification"`
	RolePermissions []RolePermissions `json:"RolePermissions"`
}

type RolePermissions struct {
	AuthorizedResourceAction string `json:"AuthorizedResourceAction,omitempty"`
	//Category                 []string `json:"Category,omitempty"`
	EAMTierLevelTagValue string `json:"EAMTierLevelTagValue,omitempty"`
	EAMTierLevelName     string `json:"EAMTierLevelName,omitempty"`
}

type RoleMap struct {
	RoleId               string `json:"RoleId"`
	EAMTierLevelTagValue string `json:"EAMTierLevelTagValue"`
	AdminTierLevel       string `json:"AdminTierLevel"`
	TenantId             string `json:"TenantId"`
}

type HTTPConfig struct {
}

type HTTPProcessor struct {
	*InputProcessor
	Config HTTPConfig
}

type HTTPResults struct {
	Results internal.QueryResults `json:"Results"`
}

func (m *HTTPProcessor) ExecuteQuery() (internal.QueryResults, error) {
	//
	//func GetRoleJson() {
	resp, err := http.Get("https://github.com/Cloud-Architekt/AzurePrivilegedIAM/raw/main/Classification/Classification_EntraIdDirectoryRoles.json")
	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("failed to get role json")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("failed to read role json")
	}

	var roles []Role
	err = json.Unmarshal(body, &roles)
	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("failed to unmarshal role json")
	}

	var EAMTierLevelTagValueAlias = map[string]string{
		"0": "admin_tier_0",
		"1": "admin_tier_1",
		"2": "admin_tier_2",
	}

	roleMaps := make([]RoleMap, 0)

	for _, role := range roles {
		roleMaps = append(roleMaps, RoleMap{
			RoleId:               role.RoleId,
			EAMTierLevelTagValue: role.Classification.EAMTierLevelTagValue,
			AdminTierLevel:       EAMTierLevelTagValueAlias[role.Classification.EAMTierLevelTagValue],
			TenantId:             m.Credentials.SentinelTenantID,
		})
	}

	results := internal.QueryResults{}

	for _, roleMap := range roleMaps {
		roleMapJson, err := json.Marshal(roleMap)
		if err != nil {
			fmt.Println(err)
			return nil, fmt.Errorf("failed to marshal role map")
		}

		var roleMapInterface map[string]interface{}
		err = json.Unmarshal(roleMapJson, &roleMapInterface)
		if err != nil {
			fmt.Println(err)
			return nil, fmt.Errorf("failed to unmarshal role map json")
		}

		results = append(results, roleMapInterface)
	}

	return results, nil
}

func (r *Role) UnmarshalJSON(data []byte) error {
	type Alias Role
	aux := &struct {
		RolePermissions json.RawMessage `json:"RolePermissions"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	var single RolePermissions
	if err := json.Unmarshal(aux.RolePermissions, &single); err == nil {
		r.RolePermissions = []RolePermissions{single}
		return nil
	}

	var multiple []RolePermissions
	if err := json.Unmarshal(aux.RolePermissions, &multiple); err != nil {
		return fmt.Errorf("RolePermissions could not be unmarshalled as an array: %v", err)
	}

	r.RolePermissions = multiple
	return nil
}
