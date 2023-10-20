# Index

- [Index](#index)
- [Credential requirements for each action](#credential-requirements-for-each-action)
  - [App registration(s)](#app-registrations)
  - [Azure Keyvault](#azure-keyvault)
  - [Sentinel Querying](#sentinel-querying)
  - [Sentinel Log Writing](#sentinel-log-writing)
    - [Workspace ID](#workspace-id)
    - [Shared Key](#shared-key)
  - [Sentinel Watchlist Writing](#sentinel-watchlist-writing)
  - [Defender for Endpoint](#defender-for-endpoint)
  - [MS Graph API](#ms-graph-api)
  - [Neo4j](#neo4j)
  - [Splunk](#splunk)
  
# Credential requirements for each action

## App registration(s)
Most of the action processors require an app registration to be able to authenticate to the various services. 
Depending on your internal policies, you may need to create a new app registration for each action processor, or you may be able to use the same app registration for all of them.

Creating a new app registration is well [documented online](https://learn.microsoft.com/en-us/azure/active-directory/develop/quickstart-register-app), and can be done via the Azure Portal, or via the Azure CLI. 

## Azure Keyvault
This step is optional. It depends on where you want to store and maintain your credentials.

Obviously, if you decide to go the Keyvault route you will need to create one.
On top of that you'll need to assign the app registration the following role:

`Key Vault Secrets User`

You can do this on the Access Control (IAM) tab of the Keyvault you want to use.

## Sentinel Querying
To be able to query Sentinel, you need to have the following roles assigned to your app registration:

`Log Analytics Reader` 
`Microsoft Sentinel Reader`

Both need to be added on the resource group containing the workspace(s) you want to query.

Also, the following Application API permissions are required (set on the app registration):

`Log Analytics API => Data.Read`

## Sentinel Log Writing
To write to Sentinel, you need to have the following details:

### Workspace ID
This is the ID of the workspace you want to write to. 
You can find this in the Azure Portal by navigating to your Microsoft Sentinel workspace, and then clicking on `Settings` and then `Workspace Settings`.
The workspace ID is the `Workspace ID` value on the top right.

### Shared Key
This is the shared key for the workspace you want to write to. 
You can find this in the Azure Portal by navigating to your Microsoft Sentinel workspace, and then clicking on `Settings` and then `Workspace Settings`.
Next, go to the `Agents`  section and unfold the `Log Analytics agent instructions`. Copy the `Primary Key` value.

## Sentinel Watchlist Writing
To write to a Sentinel Watchlist, you need to have the following roles assigned to your app registration:
`Microsoft Sentinel Contributor`

This role has more permissions than the app currently needs, so you can either assign this role to the app registration, or create a custom role with the following permissions:
`Microsoft.SecurityInsights/watchlists/delete`
`Microsoft.SecurityInsights/watchlists/write`

## Defender for Endpoint
To query Sentinel, you need to have the following roles assigned to your app registration:

`WindowsDefenderATP => AdvancedQuery.Read.All`

This will require admin consent.

## MS Graph API 
Currently, the MS Graph API action processor only collects PIM information, so this is optional to enable.
To query the MS Graph API, you need to have the following Application API permissions (set on the App registration):

`Microsoft Graph => PrivilegedAccess.Read.AzureAD`
`Microsoft Graph => PrivilegedAccess.Read.AzureADGroup`
`Microsoft Graph => PrivilegedAccess.Read.AzureResources`
`Microsoft Graph => PrivilegedEligibilitySchedule.Read.AzureADGroup`
`Microsoft Graph => RoleAssignmentSchedule.Read.Directory`
`Microsoft Graph => RoleEligibilitySchedule.Read.Directory`
`Microsoft Graph => RoleManagement.Read.All`
`Microsoft Graph => User.Read`

## Neo4j
This is the only action processor that requires a username and password to be set in the config file.
Assuming you're using the community edition this is the Neo4j user and password you set during the installation of the database.

## Splunk
Enable HTTP Event Collector:

- Log into Splunk Web on your Splunk platform instance.
- Click on "Settings".
- Under "DATA", click on "Data inputs".
- Click on "HTTP Event Collector".
- If HTTP Event Collector is not enabled, click on "Global Settings", a new window will appear.
- In the new window, check "Enable HTTP Event Collector" and if you want to change the HTTP Port Number, you can do so here.
- Click on "Save".

Create a new HEC token:

- After enabling HEC, click on "New Token".
- Fill in the name, and optionally give a description and source name override, if required.
- Under "Select Source", you can set the source type and application context, but these are optional.
- Under "Input Settings", select the allowed indexes. You can also set the default index where events will be stored.
- Click on "Review", review your settings and then click on "Submit".
- After the token is created, you will be given a token value. Make sure to save this token value, as it will be needed to send data to this HEC endpoint. For security reasons, Splunk doesn't display the token value after you leave the page.