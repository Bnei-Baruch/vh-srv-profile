package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
	"gitlab.bbdev.team/vh/vh-srv-profile/pkg/utils"
	"gitlab.bbdev.team/vh/vh-srv-profile/repo"
)

func (p *ProfileManager) updateRequest(c *gin.Context) {
	id, ok := c.Params.Get("id")

	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID missing"})
		return
	}

	// String conversion to int
	intID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id! Accepted value is INTEGER"})
		return
	}

	var request repo.NewRequest
	if err := c.Bind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		_ = c.Error(err)
		return
	}

	if *request.Status != "REQUESTED" && *request.Status != "APPROVED" && *request.Status != "DENIED" {
		err := fmt.Errorf("invalid status: %s", *request.Status)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		_ = c.Error(err)
		return
	}

	if request.Type != nil {
		if *request.Type != "hhticket" && *request.Type != "hhmembership" && *request.Type != "arvut" {
			err := fmt.Errorf("invalid type: %s", *request.Type)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			_ = c.Error(err)
			return
		}
	}

	kc_id, updateErr := p.repo.UpdateRequest(c.Request.Context(), intID, request)
	if updateErr != nil {
		if errors.Is(updateErr, common.ErrNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": updateErr.Error()})
			return
		}
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("error while updating request %q: %w", intID, updateErr))
		return
	}

	if request.Status != nil && request.EventSlug != nil && kc_id != "" {
		eventParticipationStatusUpdateTrigger(c, kc_id, *request.EventSlug, *request.Status)
	}

	c.Status(http.StatusOK)
}

func eventParticipationStatusUpdateTrigger(req *gin.Context, kcId string, slug string, status string) {

	authHeader := req.Request.Header.Get("Authorization")

	var confirmedStatus bool
	var postBody []byte

	eventUpdateFullUrl := "https://api.kli.one/events/v1/participation-status/kcid/" + kcId + "/event_slug/" + slug

	if status == "APPROVED" || status == "REQUESTED" {
		if status == "APPROVED" {
			confirmedStatus = true
		} else {
			confirmedStatus = false
		}
		postBody, _ = json.Marshal(map[string]interface{}{
			"confirmed": confirmedStatus,
		})
	} else {
		postBody, _ = json.Marshal(map[string]interface{}{
			"deleted": true,
		})
	}

	buffPostBody := bytes.NewBuffer(postBody)

	_, _ = utils.HTTPCallAndGetBody(eventUpdateFullUrl, authHeader, buffPostBody, "PATCH")

}
