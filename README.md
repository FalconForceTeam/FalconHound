![Maintenance](https://img.shields.io/maintenance/yes/2023.svg?style=flat-square)
[![Twitter](https://img.shields.io/twitter/follow/falconforceteam.svg?style=social&label=Follow)](https://twitter.com/falconforceteam)
[![Discord Shield](https://discordapp.com/api/guilds/715302469751668787/widget.png?style=shield)](https://discord.gg/CBTvTkb4)

# FalconHound

![logo](docs/falconhound-logo.png)

FalconHound is a blue team multi-tool. It allows you to utilize and enhance the power of BloodHound in a more automated fashion. It is designed to be used in conjunction with a SIEM or other log aggregation tool. 

One of the challenging aspects of BloodHound is that it is a snapshot in time. FalconHound includes functionality that can be used to keep a graph of your environment up-to-date. This allows you to see your environment as it is NOW. This is especially useful for environments that are constantly changing.

One of the hardest releationships to gather for BloodHound is the local group memberships and the session information. As blue teamers we have this information readily available in our logs. FalconHound can be used to gather this information and add it to the graph, allowing it to be used by BloodHound. 

This is just an example of how FalconHound can be used. It can be used to gather any information that you have in your logs or security tools and add it to the BloodHound graph.

Additionally, the graph can be used to trigger alerts or generate enrichment lists. 
For example, if a user is added to a certain group, FalconHound can be used to query the graph database for the shortest path to a sensitive or high-privilege group. If there is a path, this can be logged to the SIEM or used to trigger an alert.

Other examples where FalconHound can be used:
- Adding, removing or timing out sessions in the graph, based on logon and logoff events.
- Marking users and computers as compromised in the graph when they have an incident in Sentinel or MDE.
- Adding CVE information and whether there is a public exploit available to the graph.
- All kinds of Azure activities.
- Recalculating the shortest path to sensitive groups when a user is added to a group or has a new role.
- Adding new users, groups and computers to the graph.
- Generating enrichment lists for Sentinel and Splunk of, for example, Kerberoastable users or users with ownerships of certain entities.

The possibilities are endless here. Please add more ideas to the issue tracker or submit a PR.

A blog detailing more on why we developed it and some use case examples can be found [here](https://medium.com/falconforce/falconhound-attack-path-management-for-blue-teams-42adedc9cae5?source=friends_link&sk=9f64b6b3028c5a2a6087d63b4fd2c82f)

Index:
- [Supported data sources and targets](#supported-data-sources-and-targets)
- [Installation](#installation)
- [Usage](#usage)
- [Actions](#actions)
- [Extensions to the graph](#extensions-to-the-graph)
- [Credential Management](#credential-management)
- [Deployment](#deployment)
- [License](#license)

## Supported data sources and targets

FalconHound is designed to be used with BloodHound. It is not a replacement for BloodHound. It is designed to leverage the power of BloodHound and all other data platforms it supports in an automated fashion. 

Currently, FalconHound supports the following data sources and or targets:
- Azure Sentinel
- Azure Sentinel Watchlists
- Splunk
- Microsoft Defender for Endpoint
- Neo4j
- MS Graph API (early stage)
- CSV files

Additional data sources and targets are planned for the future.

At this moment, FalconHound only supports the Neo4j database for BloodHound. Support for the API of BH CE and BHE is under active development.

---

## Installation

Since FalconHound is written in Go, there is no installation required. Just download the binary from the release section and run it.
There are compiled binaries available for Windows, Linux and MacOS. You can find them in the [releases](https://github.com/FalconForceTeam/FalconHound/releases) section.

Before you can run it, you need to create a config file. You can find an example config file in the root folder. Instructions on how to creat all crededentials can be found [here](docs/required_permissions.md).

The recommened way to run FalconHound is to run it as a scheduled task or cron job. This will allow you to run it on a regular basis and keep your graph, alerts and enrichments up-to-date.

### Requirements

- BloodHound, or at least the Neo4j database for now.
- A SIEM or other log aggregation tool. Currently, Azure Sentinel and Splunk are supported.
- Credentials for each endpoint you want to talk to, with the [required permissions](docs/required_permissions.md).

### Configuration

FalconHound is configured using a YAML file. You can find an example config file in the root folder.
Each section of the config file is explained below.

--- 

## Usage

#### Default run

To run FalconHound, just run the binary and add the `-go` parameter to have it run all queries in the actions folder.
```bash
./falconhound -go
```

#### List all enabled actions
To list all enabled actions, use the `-actionlist` parameter. This will list all actions that are enabled in the config files in the actions folder. This should be used in combination with the `-go` parameter.
```bash
./falconhound -actionlist -go
```

### Run with a select set of actions
To run a select set of actions, use the `-ids` parameter, followed by one or a list of comma-separated action IDs. This will run the actions that are specified in the parameter, which can be very handy when testing, troubleshooting or when you require specific, more frequent updates. This should be used in combination with the `-go` parameter.

```bash
./falconhound -ids action1,action2,action3 -go
```

#### Run with a different config file
By default, FalconHound will look for a config file in the current directory. You can also specify a config file using the `-config` flag. This can allow you to run multiple instances of FalconHound with different configurations, against different environments.

```bash
./falconhound -go -config /path/to/config.yml
```

#### Run with a different actions folder
By default, FalconHound will look for the actions folder in the current directory. You can also specify a different folder using the `-actions-dir` flag. This makes testing and troubleshooting easier, but also allows you to run multiple instances of FalconHound with different configurations, against different environments, or at different time intervals.

```bash
./falconhound -go -actions-dir /path/to/actions
```

#### Run with credentials from a keyvault
By default, FalconHound will use the credentials in the config.yml (or a custom loaded one). By setting the `-keyvault` flag FalconHound will get the keyvault from the config and retrieve all secrets from there. Should there be items missing in the keyvault it will fall back to the config file.

```bash
./falconhound -go -keyvault
```
---

## Actions

Actions are the core of FalconHound. They are the queries that FalconHound will run. They are written in the native language of the source and target and are stored in the actions folder. Each action is a separate file and is stored in the directory of the source of the information, the query target. The filename is used as the name of the action. 

### Action folder structure

The action folder is divided into sub-directories per query source. All folders will be processed recursively and all YAML files will be executed in alphabetical order.

The Neo4j actions **should** be processed last, since their output relies on other data sources to have updated the graph database first, to get the most up-to-date results.

### Action files

All files are YAML files. The YAML file contains the query, some metadata and the target(s) of the queried information. 

There is a template file available in the root folder. You can use this to create your own actions. Have a look at the actions in the actions folder for more examples.

While most items will be fairly self explanatory,there are some important things to note about actions:

#### Enabled

As the name implies, this is used to enable or disable an action. If this is set to false, the action will not be run.
```yaml
Enabled: true
```

#### Debug

This is used to enable or disable debug mode for an action. If this is set to true, the action will be run in debug mode. This will output the results of the query to the console. This is useful for testing and troubleshooting, but is not recommended to be used in production. It will slow down the processing of the action depending on the number of results.

```yaml
Debug: false
```

#### Query

The `Query` field is the query that will be run against the source. This can be a KQL query, a SPL query or a Cypher query depending on your `SourcePlatform`.
IMPORTANT: Try to keep the query as exact as possible and only return the fields that you need. This will make the processing of the results faster and more efficient.

Additionally, when running Cypher queries, make sure to RETURN a JSON object as the result, otherwise processing will fail.
For example, this will return the Name, Count, Role and Owners of the Azure Subscriptions:

```cypher
MATCH p = (n)-[r:AZOwns|AZUserAccessAdministrator]->(g:AZSubscription) 
  RETURN {Name:g.name , Count:COUNT(g.name), Role:type(r), Owners:COLLECT(n.name)}
``` 

#### Targets

Each target has several options that can be configured. Depending on the target, some might require more configuration than others.
All targets have the `Name` and  `Enabled` fields. The `Name` field is used to identify the target. The `Enabled` field is used to enable or disable the target. If this is set to false, the target will be ignored.

#### CSV
```yaml
  - Name: CSV
    Enabled: true
    Path: path/to/filename.csv
```

#### Neo4j

The Neo4j target will write the results of the query to a Neo4j database. This output is per line and therefore it requires some additional configuration.
Since we can transfer all sorts of data in all directions, FalconHound needs to understand what to do with the data. This is done by using replacement variables in the first line of your Cypher queries. These are passed to Neo4j as parameters and can be used in the query.
The `ReplacementFields` fields are configured below.

```yaml
  - Name: Neo4j
    Enabled: true
    Query: |
      MATCH (x:Computer {name:$Computer}) MATCH (y:User {objectid:$TargetUserSid}) MERGE (x)-[r:HasSession]->(y) SET r.since=$Timestamp SET r.source='falconhound'
    Parameters:
      Computer: Computer
      TargetUserSid: TargetUserSid
      Timestamp: Timestamp
```

The Parameters section defines a set of parameters that will be replaced by the values from the query results. These can be referenced as Neo4j parameters using the `$parameter_name` syntax.

#### Sentinel

The Sentinel target will write the results of the query to a Sentinel table. The table will be created if it does not exist. The table will be created in the workspace that is specified in the config file. The data from the query will be added to the EventData field. The EventID will be the action ID and the Description will be the action name.

This is why also query output needs to be controlled, you might otherwise flood your target.

```yaml
  - Name: Sentinel
    Enabled: true
```

#### Sentinel Watchlists

The Sentinel Watchlists target will write the results of the query to a Sentinel watchlist. The watchlist will be created if it does not exist. The watchlist will be created in the workspace that is specified in the config file. All columns returned by the query will be added to the watchlist.

```yaml
 - Name: Watchlist
    Enabled: true
    WatchlistName: FH_MDE_Exploitable_Machines
    DisplayName: MDE Exploitable Machines
    SearchKey: DeviceName
    Overwrite: true       
```

The `WatchlistName` field is the name of the watchlist. The `DisplayName` field is the display name of the watchlist. 

The `SearchKey` field is the column that will be used as the search key. 

The `Overwrite` field is used to determine if the watchlist should be overwritten or appended to. If this is set to false, the results of the query will be appended to the watchlist. If this is set to true, the watchlist will be deleted and recreated with the results of the query.

#### Splunk

Like Sentinel, Splunk will write the results of the query to a Splunk index. The index will need to be created and tied to a HEC endpoint. The data from the query will be added to the EventData field.  The EventID will be the action ID and the Description will be the action name.

```yaml
  - Name: Splunk
    Enabled: true
```

### Extensions to the graph

#### Relationship: HadSession

Once a session has ended, it had to be removed from the graph, but this felt like a waste of information. So instead of removing the session,it will be added as a relationship between the computer and the user. The relationship will be called `HadSession`. The relationship will have the following properties:

```json
{
  "till": "2021-08-31T14:00:00Z",
  "source": "falconhound",
  "reason": "logoff",
}
```

This allows for additional path discoveries where we can investigate whether the user ever logged on to a certain system, even if the session has ended.

#### Properties

FalconHound will add the following properties to nodes in the graph:

Computer:
    - 'exploitable': true/false
    - 'exploits': list of CVEs
    - 'alertids': list of alert ids

## Credential management

The currently supported ways of providing FalconHound with credentials are:

- Via the config.yml file on disk.
- Keyvault secrets. This still requires a ServicePrincipal with secrets in the yaml.
- Mixed mode.

#### Config.yml

The config file holds all details required by each platform. All items in the config file are **case-sensitive**.
Best practise is to separate the apps on a per service level but you *can* use 1 AppID/AppSecret for all Azure based actions.

The required permissions for your AppID/AppSecret are listed [here](docs/required_permissions.md).

#### Keyvault

A more secure way of storing the credentials would be to use an Azure KeyVault. Be aware that there is a small [cost aspect](https://azure.microsoft.com/en-us/pricing/details/key-vault/) to using Keyvaults. 
Access to KeyVaults currently only supports authentication based on a AppID/AppSecret which needs to be configured in the config.yml file.

The recommended way to set this up is to use a ServicePrincipal that only has the `Key Vault Secrets User` role to this Keyvault. This role only allows access to the secrets, not even list them. Do *NOT* reuse the ServicePrincipal which has access to Sentinel and/or MDE, since this almost completely negates the use of a Keyvault.  

The items to configure in the Keyvault are listed below. Please note Keyvault secrets are **not** case-sensitive.

```
SentinelAppSecret
SentinelAppID
SentinelTenantID
SentinelTargetTable
SentinelResourceGroup
SentinelSharedKey
SentinelSubscriptionID
SentinelWorkspaceID
SentinelWorkspaceName
MDETenantID
MDEAppID
MDEAppSecret
Neo4jUri
Neo4jUsername
Neo4jPassword
GraphTenantID
GraphAppID
GraphAppSecret
SplunkUri
SplunkToken
```

Once configured you can add the `-keyvault` parameter while starting FalconHound.

#### Mixed mode / fallback

When the `-keyvault` parameter is set on the command-line, this will be the primary source for all required secrets. Should FalconHound fail to retrieve items, it will fall back to the equivalent item in the `config.yml`. If both fail and there are actions enabled for that source or target, it will throw errors on attempts to authenticate.

## Deployment

FalconHound is designed to be run as a scheduled task or cron job. This will allow you to run it on a regular basis and keep your graph, alerts and enrichments up-to-date.
Depending on the amount of actions you have enabled, the amount of data you are processing and the amount of data you are writing to the graph, this can take a while.

All log based queries are built to run every 15 minutes. Should processing take too long you might need to tweak this a little.
If this is the case it might be recommended to disable certain actions.

Also there might be some overlap with for instance the session actions. If you have a lot of sessions you might want to disable the session actions for Sentinel and rely on the one from MDE. This is assuming you have MDE and Sentinel connected and most machines are onboarded into MDE.

### Sharphound / Azurehound

While FalconHound is designed to be used with BloodHound, it is not a replacement for Sharphound and Azurehound. It is designed to compliment the collection and remove the moment-in-time problem of the peroiodic collection. Both Sharphound and Azurehound are still required to collect the data, since not all similar data is available in logs.

It is recommended to run Sharphound and Azurehound on a regular basis, for example once a day/week or month, and FalconHound every 15 minutes.

## License

This project is licensed under the BSD3 License - see the [LICENSE](LICENSE) file for details.

This means you can use this software for free, even in commercial products, as long as you credit us for it.
You cannot hold us liable for any damages caused by this software.
