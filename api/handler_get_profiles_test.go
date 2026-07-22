package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// The keycloak_id list filter is capped — an oversized list must be rejected
// before reaching the repo.
func Test_getProfiles_TooManyKeycloakIDs_BadRequest(t *testing.T) {
	g := gin.New()
	g.GET("/v1/profiles", NewProfileManager(&storageMock{}).getProfiles)

	ids := make([]string, maxKeycloakIDsPerRequest+1)
	for i := range ids {
		ids[i] = fmt.Sprintf("11000000-0000-0000-0000-%012d", i)
	}

	r := NewRequestAsRoot(http.MethodGet, "/v1/profiles?limit=10&skip=0&keycloak_id="+strings.Join(ids, ","), nil)
	w := httptest.NewRecorder()
	g.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "too many keycloak_ids")
}
