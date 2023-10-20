package internal

// The `config:"name"` tags are used to identify under which key the value is stored
// in the config.
type Credentials struct {
	SentinelAppSecret      string `config:"sentinel.appSecret"`
	SentinelAppID          string `config:"sentinel.appID"`
	SentinelTenantID       string `config:"sentinel.tenantID"`
	SentinelTargetTable    string `config:"sentinel.targetTable"`
	SentinelResourceGroup  string `config:"sentinel.resourceGroup"`
	SentinelSharedKey      string `config:"sentinel.sharedKey"`
	SentinelSubscriptionID string `config:"sentinel.subscriptionID"`
	SentinelWorkspaceID    string `config:"sentinel.workspaceID"`
	SentinelWorkspaceName  string `config:"sentinel.workspaceName"`
	MDETenantID            string `config:"mde.tenantID"`
	MDEAppID               string `config:"mde.appID"`
	MDEAppSecret           string `config:"mde.appSecret"`
	GraphTenantID          string `config:"graph.tenantID"`
	GraphAppID             string `config:"graph.appID"`
	GraphAppSecret         string `config:"graph.appSecret"`
	Neo4jUri               string `config:"neo4j.uri"`
	Neo4jUsername          string `config:"neo4j.username"`
	Neo4jPassword          string `config:"neo4j.password"`
	SplunkUri              string `config:"splunk.url"` // TODO inconsistent one is called uri the other url
	SplunkToken            string `config:"splunk.token"`
	BHUrl                  string `config:"bloodhound.url"`
	BHTokenID              string `config:"bloodhound.tokenID"`
	BHTokenKey             string `config:"bloodhound.tokenKey"`
}
