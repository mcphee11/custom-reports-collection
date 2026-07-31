package genesys

import (
	"github.com/mypurecloud/platform-client-sdk-go/v188/platformclientv2"
)

func GetOrganization(config *platformclientv2.Configuration) (msg, orgID, orgName string, err error) {
	apiInstance := platformclientv2.NewOrganizationApiWithConfig(config)
	data, _, err := apiInstance.GetOrganizationsMe()
	if err != nil {
		return err.Error(), "", "", err
	} else {
		return "Organization details recieved", *data.Id, *data.Name, nil
	}
}
