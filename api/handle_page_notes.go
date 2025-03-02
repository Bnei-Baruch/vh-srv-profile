package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
)

type insertPageNoteRequest struct {
	PageId         int    `json:"page_id"`
	PageKeycloakId string `json:"page_keycloak_id"`
	Content        string `json:"content"`
}

func (p *ProfileManager) handleFetchPageNotes(c *gin.Context) {
	if !p.HasAnyRole(c, common.RoleRoot, common.RoleAdmin) {
		return
	}
	pageKeycloakId := c.Query("page_keycloak_id")
	pageId := c.Query("page_id")

	intPageId, err := strconv.Atoi(pageId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pageId value! Accepted value is INTEGER"})
		return
	}

	res, err := p.repo.FetchPageNotes(c.Request.Context(), intPageId, pageKeycloakId)
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

	keycloakId, ok := p.GetKeycloakIdFromRequest(c)
	if !ok {
		c.Status(http.StatusForbidden)
		return
	}

	var req insertPageNoteRequest
	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	noteId, err := p.repo.InsertPageNote(c, req.PageId, req.PageKeycloakId, keycloakId, req.Content)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.InsertPageNote: %w", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Added!", "note_id": noteId})
}

func (p *ProfileManager) handleDeletePageNote(c *gin.Context) {
	if !p.HasAnyRole(c, common.RoleRoot, common.RoleAdmin) {
		return
	}

	noteId := c.Param("id")
	noteIdInt, err := strconv.Atoi(noteId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	keycloakId, ok := p.GetKeycloakIdFromRequest(c)
	if !ok {
		c.Status(http.StatusForbidden)
		return
	}

	err = p.repo.DeletePageNote(c, noteIdInt, keycloakId)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.DeletePageNote: %w", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Deleted!", "note_id": noteId})
}
