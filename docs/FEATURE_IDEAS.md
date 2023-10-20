# Feature ideas and ambitions for the project

## Actions ideas

- [ ] Add new role to Azure user
- [ ] Group added
- [ ] Role assignment
- [ ] API access changes
- [ ] Added to local admins group > AdminTo
- [ ] Owned to sensitive resource with more than 1 owned
- [ ] Next step on path from owned prediction
- [ ] Old last set passwords
- [ ] User owns list
- [ ] Get public groups and dynamic groups
- [ ] Public groups with path
- [ ] Dynamic groups with path
- [ ] Machine onboarded in EDR property to Neo4j
- [ ] Query for not onboarded machines with a patch to sensitive

- [ ] Azure risk score to users / devices
- [ ] Get VMs and IPs from graph
- [ ] Get conditional access policies

## In/out processors

Sensitive resource list processor
- [ ] Read from CSV
- [ ] Read from watchlist

BH(E) API - under development
- [ ] Read from new BH(E) API > TODO: wait for query over API to be fixed/improved
- [-] Query new BH(E) API, parse results
- [ ] Write to new BH(E) API  > TODO: solve the objectid issue

Generic output processors
- [ ] Write BH compatible JSON outputs
- [ ] Write markdown outputs
- [ ] Write to storage account
- [ ] Write to ADX

GraphAPI
- [ ] Look into Defender, AAD, Intune, CA policies, ?
- [ ] Write GraphAPI (User properties)
- [ ] Read signinlog
- [ ] Read auditlog / azureactivity

- [ ] Read watchlists
- [x] Save watchlists

- [ ] Read from Splunk

## Operational

- [ ] Token needs to be used more efficiently
- [ ] Add global debug mode
- [ ] Add check that if a credential is empty in the config the in/out processor is not used

- [ ] Add more logging
- [ ] Add more error handling

## Future releases

- [ ] Add env variable to creds options
- [ ] Add managed identity option
- [ ] Excel sheet reports
- [ ] Configurable time window for actions
- [ ] Write to Teams
- [ ] Write to Slack