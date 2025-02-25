package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/volatiletech/null/v9"

	"gitlab.bbdev.team/vh/vh-srv-profile/common"
)

type insertPageNoteRequest struct {
	PageId         int    `json:"page_id"`
	PageKeycloakId string `json:"page_keycloak_id"`
	Content        string `json:"content"`
}

type deletePageNoteRequest struct {
	NoteId int `json:"note_id"`
}

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
	nullablePageKeycloakId := null.NewString(req.PageKeycloakId, req.PageKeycloakId != "")

	noteId, err := p.repo.InsertPageNote(c, req.PageId, nullablePageKeycloakId, keycloakId, req.Content)
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

	keycloakId, ok := p.GetKeycloakIdFromRequest(c)
	if !ok {
		c.Status(http.StatusForbidden)
		return
	}

	var req deletePageNoteRequest
	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := p.repo.DeletePageNote(c, req.NoteId, keycloakId)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		_ = c.Error(fmt.Errorf("repo.DeletePageNote: %w", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Deleted!", "note_id": req.NoteId})
}
