package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/volatiletech/null/v9"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
)

func (p *ProfileManager) handleFetchPageNotes(c *gin.Context) {
	if !p.HasAnyRole(c, common.RoleRoot, common.RoleAdmin) {
		return
	}
	pageKeycloakId := c.Query("pageKeycloakId")
	pageId := c.Query("pageId")

	intPageId, err := strconv.Atoi(pageId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pageId value! Accepted value is INTEGER"})
		return
	}
	nullablePageKeycloakId := null.NewString(pageKeycloakId, pageKeycloakId != "")

	res, err := p.repo.FetchPageNotes(c.Request.Context(), intPageId, nullablePageKeycloakId)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.FetchPageNotes: %w", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Fetched!", "data": res})
}

func (p *ProfileManager) handleInsertPageNote(c *gin.Context) {
	if !p.HasAnyRole(c, common.RoleRoot, common.RoleAdmin) {
		return
	}

	// TBD repo insert to db

	//c.JSON(http.StatusOK, gin.H{"status": true, "message": "Added!", "data": ID})
}

func (p *ProfileManager) handleDeletePageNote(c *gin.Context) {
	if !p.HasAnyRole(c, common.RoleRoot, common.RoleAdmin) {
		return
	}

	// TBD check that the note belong to author
	// TBD repo delete from db

	//c.JSON(http.StatusOK, gin.H{"status": true, "message": "Deleted!", "data": ID})
}
