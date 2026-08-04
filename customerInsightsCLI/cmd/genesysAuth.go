package genesys

import (
	"time"

	"github.com/mypurecloud/platform-client-sdk-go/v188/platformclientv2"
)

func CheckForEnv(region, clientID, secret, startDate, endDate, timeZone string) (message string) {
	response := ""
	if region == "" {
		response += "missing region\n"
	}
	if clientID == "" {
		response += "missing clientId\n"
	}
	if secret == "" {
		response += "missing secret\n"
	}
	if startDate == "" {
		response += "missing startDate\n"
	}
	if endDate == "" {
		response += "missing endDate\n"
	}
	if timeZone == "" {
		response += "missing timeZone\n"
	}
	return response
}

func GenesysAuth(region, clientID, secret string) (conf *platformclientv2.Configuration, err error) {
	// Do Genesys Cloud OAuth
	config := platformclientv2.GetDefaultConfiguration()
	config.BasePath = "https://api." + region
	config.RetryConfiguration = &platformclientv2.RetryConfiguration{
		RetryWaitMin: 5 * time.Second,
		RetryWaitMax: 60 * time.Second,
		RetryMax:     5,
	}
	config.AutomaticTokenRefresh = true
	if err := config.AuthorizeClientCredentials(clientID, secret); err != nil {
		return config, err
	}
	return config, nil
}
