package keycloak

import (
	"context"
	"fmt"

	"github.com/Nerzal/gocloak/v13"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
)

type KeycloakService interface {
	UpdateUser(ctx context.Context, keycloakID string, firstName *string, lastName *string) error
}

type KeycloakServiceFactory func() KeycloakService

func KeycloakAPIFactory() KeycloakService {
	return NewKeycloakAPI()
}

type KeycloakAPI struct {
	kc *gocloak.GoCloak
}

func NewKeycloakAPI() *KeycloakAPI {
	c := new(KeycloakAPI)
	c.kc = gocloak.NewClient(common.Config.KeycloakServerUrl)
	gocloak.SetLegacyWildFlySupport()(c.kc)
	return c
}

func (c *KeycloakAPI) UpdateUser(ctx context.Context, keycloakID string, firstName *string, lastName *string) error {
	tokenSource := ctx.Value(common.CtxTokenSource).(TokenSource)
	token, err := tokenSource.Token()
	if err != nil {
		return fmt.Errorf("tokenSource.Token(): %w", err)
	}
	user, infoErr := c.kc.GetUserByID(ctx, token, common.Config.KeycloakRealm, keycloakID)

	if infoErr != nil {
		return infoErr
	}

	if firstName != nil {
		firstName = user.FirstName
	}

	if lastName != nil {
		lastName = user.LastName
	}

	// only update the user if user details are not same
	if *user.ID == keycloakID &&
		user.FirstName == firstName &&
		user.LastName == lastName {
		return nil
	}

	// Only update firstName & lastName
	user.FirstName = firstName
	user.LastName = lastName

	updaterErr := c.kc.UpdateUser(ctx, token, common.Config.KeycloakRealm, *user)
	if updaterErr != nil {
		return updaterErr
	}

	return nil
}
