package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
	"github.com/spf13/viper"
)

func GetSecretFromAzureKeyVault(keyVaultName string, secretName string, managedIdentity string) (string, error) {
	// Create a new DefaultAzureCredential
	var cred azcore.TokenCredential
	var err error
	if managedIdentity == "true" {
		cred, err = azidentity.NewManagedIdentityCredential(nil)
	} else {
		cred, err = azidentity.NewClientSecretCredential(viper.GetString("keyvault.tenantID"), viper.GetString("keyvault.appID"), viper.GetString("keyvault.appSecret"), nil)
	}
	// cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatalf("Failed to create the credentials: %v", err)
	}

	// Create a new client using the DefaultAzureCredential.
	client, err := azsecrets.NewClient(keyVaultName, cred, nil)
	if err != nil {
		log.Fatalf("Failed to create the client: %v", err)
	}

	// Get the secret
	secretResponse, err := client.GetSecret(context.Background(), secretName, "", &azsecrets.GetSecretOptions{})
	if err != nil {
		return "", fmt.Errorf("Failed to get the secret: %v", err)
	}

	return *secretResponse.Value, nil
}
