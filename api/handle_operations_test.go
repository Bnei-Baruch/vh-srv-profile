package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
)

func revertRequest(t *testing.T, storeErr error) int {
	t.Helper()
	sm := storageMock{}
	sm.On("RevertOperation", mock.Anything, "o'new@test.test", "o'old@test.test").Return(storeErr)
	g := gin.New()
	g.POST("/revert", NewProfileManager(&sm).handleOperationRevert)

	r := NewRequestAsRoot(http.MethodPost, "/revert",
		strings.NewReader(`{"new_email":"o'new@test.test","old_email":"o'old@test.test"}`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	g.ServeHTTP(w, r)
	sm.AssertExpectations(t)
	return w.Code
}

func Test_handleOperationRevert_status(t *testing.T) {
	assert.Equal(t, http.StatusOK, revertRequest(t, nil))
	assert.Equal(t, http.StatusNotFound, revertRequest(t, common.ErrOperationNotFound))
	assert.Equal(t, http.StatusConflict, revertRequest(t, common.ErrOperationReverted))
	assert.Equal(t, http.StatusConflict, revertRequest(t, common.ErrOperationNoMatch))
	assert.Equal(t, http.StatusConflict, revertRequest(t, common.ErrOperationConflict))
	assert.Equal(t, http.StatusInternalServerError, revertRequest(t, errors.New("db down")))
}

func createRequest(t *testing.T, storeErr error) int {
	t.Helper()
	sm := storageMock{}
	sm.On("PerformOperation", mock.Anything, mock.Anything).Return(7, storeErr)
	g := gin.New()
	g.POST("/", NewProfileManager(&sm).handleOperationCreate)

	r := NewRequestAsRoot(http.MethodPost, "/", strings.NewReader(
		`{"type":"email_update","new_email":"o'new@test.test","old_email":"o'old@test.test","new_keycloak_id":"kc"}`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	g.ServeHTTP(w, r)
	sm.AssertExpectations(t)
	return w.Code
}

func Test_handleOperationCreate_status(t *testing.T) {
	assert.Equal(t, http.StatusOK, createRequest(t, nil))
	assert.Equal(t, http.StatusConflict, createRequest(t, common.ErrOperationNoMatch))
	assert.Equal(t, http.StatusConflict, createRequest(t, common.ErrOperationConflict))
	assert.Equal(t, http.StatusInternalServerError, createRequest(t, errors.New("db down")))
}
