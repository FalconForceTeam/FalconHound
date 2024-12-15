package cmd

import (
	"falconhound/internal"
	"log"
	"path/filepath"
	"reflect"

	"github.com/spf13/viper"
)

func setCredValue(credentials *internal.Credentials, field string, value string) {
	r := reflect.ValueOf(credentials)
	f := reflect.Indirect(r).FieldByName(field)
	if f.Kind() != reflect.Invalid {
		f.SetString(value)
	}
}

func GetCreds(configFile string, keyvaultFlag bool) (theCreds internal.Credentials) {
	var err error
	//read config file
	dir := filepath.Dir(configFile)
	fileName := filepath.Base(configFile)

	viper.SetConfigName(fileName)
	if configFile == "config.yml" {
		viper.AddConfigPath(".")
	} else {
		viper.AddConfigPath(dir)
	}
	viper.SetConfigType("yml")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read configuration file: %v", err)
	}

	log.Printf("[+] Using config file: %s\n", viper.ConfigFileUsed())
	if keyvaultFlag {
		log.Printf("[+] Using keyvault: %s\n", viper.GetString("keyvault.uri"))
	}

	// Parse the Credentials structure, either get it from the config or from the keyvault
	// The values in keyvault are equal to the field names in the Credentials struct
	// the values in the config are equal specified using a tag in the Credentials struct
	creds := internal.Credentials{}
	t := reflect.TypeOf(internal.Credentials{})
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		// Get the field tag value
		tag := field.Tag.Get("config")
		var value string
		if keyvaultFlag {
			value, err = GetSecretFromAzureKeyVault(viper.GetString("keyvault.uri"), field.Name, viper.GetString("keyvault.authType"))
			if err != nil {
				LogInfo("[!] %s not in keyvault, grabbing it from the config...", field.Name)
				value = viper.GetString(tag)
			}
		} else {
			value = viper.GetString(tag)
		}
		setCredValue(&creds, field.Name, value)
	}
	return creds
}
