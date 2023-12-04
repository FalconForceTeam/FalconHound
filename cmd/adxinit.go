package cmd

import (
	"context"
	"falconhound/internal"
	"github.com/Azure/azure-kusto-go/kusto"
	"log"
)

func AdxInitTable(creds internal.Credentials) error {

	kustoConnectionStringBuilder := kusto.NewConnectionStringBuilder(creds.AdxClusterURL)
	kustoConnectionString := kustoConnectionStringBuilder.WithAadAppKey(creds.AdxAppID, creds.AdxAppSecret, creds.AdxTenantID)

	client, err := kusto.New(kustoConnectionString)
	if err != nil {
		log.Fatalf("failed to create Kusto client: %s", err)
		return err
	}

	// Create a context
	ctx := context.Background()
	const command = (".create table FalconHound (Name: string, Description: string, EventID: string, BHQuery: string, EventData: dynamic, Timestamp: datetime)")

	// Execute the control command
	_, err = client.Mgmt(ctx, creds.AdxDatabase, kusto.NewStmt(command))
	//_, err = client.Mgmt(ctx, creds.AdxDatabase, command)
	if err != nil {
		log.Fatalf("Failed to execute the control command: %s", err)
		return err
	}

	LogInfo("[+] Table FalconHound created successfully, ready for ingestion.")
	return nil
}
