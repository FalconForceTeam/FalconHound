package input_cmd

import (
	"context"
	"encoding/json"
	"fmt"
	abstractions "github.com/microsoft/kiota-abstractions-go"
	msgraphsdk "github.com/microsoftgraph/msgraph-beta-sdk-go"
	"github.com/microsoftgraph/msgraph-beta-sdk-go/models"
	graphusers "github.com/microsoftgraph/msgraph-beta-sdk-go/users"
	msgraphcore "github.com/microsoftgraph/msgraph-sdk-go-core"
	"strings"
)

func GetMFA(graphClient msgraphsdk.GraphServiceClient) ([]byte, error) {
	// Requires User.Read.All and UserAuthenticationMethod.Read.All permissions
	// or one of the following roles: Global Reader, Privileged Authentication Administrator
	// Authentication Administrator is also possible, but only shows masked phone numbers and email addresses.
	// Both Administrator roles are not recommended for production use, since they have way too many permissions.

	type MFASettings struct {
		SignInPreference        string
		AuthMethods             []string
		MfaAuthMethods          string
		PhoneNumber             string
		SmsSignInState          string
		HelloDevice             string
		ObjectId                string
		UserName                string
		AuthenticatorDeviceName string
		FidoDeviceName          string
		FidoModel               string
		MfaEmailAddress         string
	}

	conversionMap := map[string]string{
		"#microsoft.graph.phoneAuthenticationMethod":                   "Phone",
		"#microsoft.graph.windowsHelloForBusinessAuthenticationMethod": "WindowsHelloForBusiness",
		"#microsoft.graph.microsoftAuthenticatorAuthenticationMethod":  "MicrosoftAuthenticator",
		"#microsoft.graph.fido2AuthenticationMethod":                   "Fido2",
		"#microsoft.graph.softwareOathAuthenticationMethod":            "SoftwareOath",
		"#microsoft.graph.emailAuthenticationMethod":                   "Email",
		"#microsoft.graph.passwordAuthenticationMethod":                "Password",
		"#microsoft.graph.passwordlessMicrosoftAuthenticatorMethods":   "PasswordlessMicrosoftAuthenticator",
	}

	mfaSettingsPerUser := make([]MFASettings, 0)

	headers := abstractions.NewRequestHeaders()
	headers.Add("ConsistencyLevel", "eventual")
	// only get required information to conserve memory and bandwith
	requestParameters := &graphusers.UsersRequestBuilderGetQueryParameters{
		Select: []string{"id", "displayName"},
	}
	configuration := &graphusers.UsersRequestBuilderGetRequestConfiguration{
		Headers:         headers,
		QueryParameters: requestParameters,
	}

	usersResponse, err := graphClient.Users().Get(context.Background(), configuration)
	if err != nil {
		fmt.Println("Error getting users:", err)
		fmt.Println("Make sure to assign the User.Read.All and UserAuthenticationMethod.Read.All permissions to the app.")
		return nil, err
	}

	pageIterator, err := msgraphcore.NewPageIterator[models.Userable](usersResponse, graphClient.GetAdapter(), models.CreateUserCollectionResponseFromDiscriminatorValue)

	err = pageIterator.Iterate(context.Background(), func(user models.Userable) bool {
		settings := MFASettings{}
		settings.ObjectId = *user.GetId()
		settings.UserName = *user.GetDisplayName()
		methods, err := graphClient.Users().ByUserId(*user.GetId()).Authentication().Methods().Get(context.Background(), nil)
		if err != nil {
			fmt.Println("Error getting MFA Authentication methods for user", *user.GetId(), ":", err)
			return false
		}
		signInPreferences, err := graphClient.Users().ByUserId(*user.GetId()).Authentication().SignInPreferences().Get(context.Background(), nil)
		if err != nil {
			fmt.Println("Error getting MFA Authentication methods for user", *user.GetId(), ":", err)
			return false
		}

		if signInPreferences.GetUserPreferredMethodForSecondaryAuthentication() != nil {
			settings.SignInPreference = (*signInPreferences.GetUserPreferredMethodForSecondaryAuthentication()).String()
		} else {
			settings.SignInPreference = "no-mfa"
		}

		authMethods := methods.GetValue()
		for _, method := range authMethods {
			methodType := *method.GetOdataType()
			if methodType == "#microsoft.graph.passwordAuthenticationMethod" {
				// Skip this iteration if methodType is "#microsoft.graph.passwordAuthenticationMethod"
				continue
			}
			if newValue, ok := conversionMap[methodType]; ok {
				settings.AuthMethods = append(settings.AuthMethods, newValue)
			} else {
				settings.AuthMethods = append(settings.AuthMethods, methodType)
			}
			if *method.GetOdataType() == "#microsoft.graph.phoneAuthenticationMethod" {
				phone, _ := graphClient.Users().ByUserId(*user.GetId()).Authentication().PhoneMethods().Get(context.Background(), nil)
				phoneData := phone.GetValue()
				settings.PhoneNumber = *phoneData[0].GetPhoneNumber()
				settings.SmsSignInState = (*phoneData[0].GetSmsSignInState()).String()
			}
			if *method.GetOdataType() == "#microsoft.graph.windowsHelloForBusinessAuthenticationMethod" {
				hello, _ := graphClient.Users().ByUserId(*user.GetId()).Authentication().WindowsHelloForBusinessMethods().Get(context.Background(), nil)
				helloData := hello.GetValue()
				settings.HelloDevice = *helloData[0].GetDisplayName()
			}
			if *method.GetOdataType() == "#microsoft.graph.microsoftAuthenticatorAuthenticationMethod" || *method.GetOdataType() == "#microsoft.graph.passwordlessMicrosoftAuthenticatorMethods" {
				ma, _ := graphClient.Users().ByUserId(*user.GetId()).Authentication().MicrosoftAuthenticatorMethods().Get(context.Background(), nil)
				maData := ma.GetValue()
				settings.AuthenticatorDeviceName = *maData[0].GetDisplayName()
			}
			if *method.GetOdataType() == "#microsoft.graph.fido2AuthenticationMethod" {
				fido, _ := graphClient.Users().ByUserId(*user.GetId()).Authentication().Fido2Methods().Get(context.Background(), nil)
				fidoData := fido.GetValue()
				settings.FidoDeviceName = *fidoData[0].GetDisplayName()
				settings.FidoModel = *fidoData[0].GetModel()
			}
			if *method.GetOdataType() == "#microsoft.graph.emailAuthenticationMethod" {
				email, _ := graphClient.Users().ByUserId(*user.GetId()).Authentication().EmailMethods().Get(context.Background(), nil)
				emailData := email.GetValue()
				settings.MfaEmailAddress = *emailData[0].GetEmailAddress()
			}
			// add the auth methods to a single string for easier ingestion into neo4j
			settings.MfaAuthMethods = strings.Join(settings.AuthMethods, ",")
		}
		mfaSettingsPerUser = append(mfaSettingsPerUser, settings)
		return true
	})
	if err != nil {
		fmt.Println("Error iterating over users:", err)
		return nil, err
	}

	json, err := json.MarshalIndent(mfaSettingsPerUser, "", "  ")
	if err != nil {
		fmt.Println("Error converting to JSON:", err)
		return nil, err
	}

	return json, nil

}
