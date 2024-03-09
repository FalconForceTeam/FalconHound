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
	AdxTenantID            string `config:"adx.tenantID"`
	AdxAppID               string `config:"adx.appID"`
	AdxAppSecret           string `config:"adx.appSecret"`
	AdxClusterURL          string `config:"adx.clusterUrl"`
	AdxDatabase            string `config:"adx.database"`
	Neo4jUri               string `config:"neo4j.uri"`
	Neo4jUsername          string `config:"neo4j.username"`
	Neo4jPassword          string `config:"neo4j.password"`
	SplunkUrl              string `config:"splunk.url"`
	SplunkIndex            string `config:"splunk.index"`
	SplunkApiPort          string `config:"splunk.apiport"`
	SplunkApiToken         string `config:"splunk.apitoken"`
	SplunkHecPort          string `config:"splunk.hecport"`
	SplunkHecToken         string `config:"splunk.hectoken"`
	BHUrl                  string `config:"bloodhound.url"`
	BHTokenID              string `config:"bloodhound.tokenID"`
	BHTokenKey             string `config:"bloodhound.tokenKey"`
	LogScaleUrl            string `config:"logscale.url"`
	LogScaleToken          string `config:"logscale.token"`
	LogScaleRepository     string `config:"logscale.repository"`
	LimaCharlieAPIUrl      string `config:"limacharlie.apiurl"`
	LimaCharlieOrgId       string `config:"limacharlie.orgid"`
	LimaCharlieIngestKey   string `config:"limacharlie.ingestkey"`
	ElasticCloudID         string `config:"elastic.cloudid"`
	ElasticApiKey          string `config:"elastic.apikey"`
}
