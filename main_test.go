package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func okHandler(c *gin.Context) {
	c.Status(http.StatusOK)
}

func Test_initApp_configures_profile_create_route(t *testing.T) {
	app := initApp(appHandlers{
		create: okHandler,
	})
	recorder := httptest.NewRecorder()
	app.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/v1/profile", nil))

	assert.Equal(t, http.StatusOK, recorder.Code)
}
