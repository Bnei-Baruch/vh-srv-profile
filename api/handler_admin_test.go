package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
  "context"
  "io"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitlab.bbdev.team/vh/vh-srv-profile/api/middleware"
	"gitlab.bbdev.team/vh/vh-srv-profile/common"
)

func NewRequestAsRoot(method, target string, body io.Reader) *http.Request {
	r := httptest.NewRequest(method, target, body)
  ctx := r.Context()
  claims := &middleware.IDTokenClaims{RealmAccess: middleware.Roles{ Roles: []string{common.RoleRoot} }}
  ctx = context.WithValue(ctx, common.CtxAuthClaims, claims)
  return r.WithContext(ctx)
}

func Test_profileHandler_hardDelete_succeeds(t *testing.T) {
	sm := storageMock{}
	sm.On("HardDeleteProfile", mock.Anything, uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")).Return(nil)
	profile := NewProfileManager(&sm)
	g := gin.New()
	g.DELETE("/:keycloak_id", profile.hardDelete)

	r := NewRequestAsRoot(http.MethodDelete, "/11000000-0000-0000-0000-000000000000", nil)
	w := httptest.NewRecorder()
	g.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func Test_profileHandler_hardDelete_returns_404_when_storage_returns_errProfileNotFound(t *testing.T) {
	sm := storageMock{}
	sm.On("HardDeleteProfile", mock.Anything, mock.Anything).Return(fmt.Errorf("%w: %q",
		common.ErrProfileNotFound, "example string"))
	profile := NewProfileManager(&sm)
	g := gin.New()
	g.DELETE("/:keycloak_id", profile.hardDelete)

	r := NewRequestAsRoot(http.MethodDelete, "/11000000-0000-0000-0000-000000000000", nil)
	w := httptest.NewRecorder()
	g.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_profileHandler_hardDelete_returns_500_when_storage_returns_error(t *testing.T) {
	sm := storageMock{}
	sm.On("HardDeleteProfile", mock.Anything, mock.Anything).Return(fmt.Errorf("some error"))
	profile := NewProfileManager(&sm)
	g := gin.New()
	g.DELETE("/:keycloak_id", profile.hardDelete)

	r := NewRequestAsRoot(http.MethodDelete, "/11000000-0000-0000-0000-000000000000", nil)
	w := httptest.NewRecorder()
	g.ServeHTTP(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
