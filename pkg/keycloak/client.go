package keycloak

import (
	"fmt"
	"strings"

	"github.com/Nerzal/gocloak"
)

type KeycloakClient interface {
	UpdateUser(authToken string, keycloakID string, firstName string, lastName string) error
}

type Client struct {
	serverUrl string
	realm     string
	kc        gocloak.GoCloak
}

func NewClient(serverUrl string, realm string) *Client {
	c := new(Client)
	c.serverUrl = serverUrl
	c.realm = realm
	c.kc = gocloak.NewClient(serverUrl)
	return c
}

func (c *Client) UpdateUser(authToken string, keycloakID string, firstName string, lastName string) error {
	tokenParts := strings.Split(authToken, " ")

	validToken := tokenParts[1]
	if validToken == "" {
		return fmt.Errorf("no token found")
	}

	keycloakUserInfo, infoErr := c.kc.GetUserByID(validToken, c.realm, keycloakID)

	if infoErr != nil {
		return infoErr
	}

	if firstName == "" {
		firstName = keycloakUserInfo.FirstName
	}

	if lastName == "" {
		lastName = keycloakUserInfo.LastName
	}

	// only update the user if user details are not same
	if keycloakUserInfo.ID == keycloakID &&
		keycloakUserInfo.FirstName == firstName &&
		keycloakUserInfo.LastName == lastName {
		return nil
	}

	// Only update firstName & lastName
	updateObj := gocloak.User{
		ID:                         keycloakID,
		FirstName:                  firstName,
		LastName:                   lastName,
		CreatedTimestamp:           keycloakUserInfo.CreatedTimestamp,
		Username:                   keycloakUserInfo.Username,
		Enabled:                    keycloakUserInfo.Enabled,
		Totp:                       keycloakUserInfo.Totp,
		EmailVerified:              keycloakUserInfo.EmailVerified,
		Email:                      keycloakUserInfo.Email,
		FederationLink:             keycloakUserInfo.FederationLink,
		Attributes:                 keycloakUserInfo.Attributes,
		DisableableCredentialTypes: keycloakUserInfo.DisableableCredentialTypes,
		RequiredActions:            keycloakUserInfo.RequiredActions,
		Access:                     keycloakUserInfo.Access,
	}

	updaterErr := c.kc.UpdateUser(validToken, c.realm, updateObj)

	if updaterErr != nil {
		return updaterErr
	}

	return nil
}
