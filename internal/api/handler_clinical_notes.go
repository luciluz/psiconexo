package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/luciluz/psiconexo/internal/service"
)

type createNoteDTO struct {
	ProfessionalID int64  `json:"professional_id" binding:"required"`
	ClientID       int64  `json:"client_id" binding:"required"`
	AppointmentID  *int64 `json:"appointment_id"` // Opcional
	Content        string `json:"content" binding:"required"`
}

type updateNoteDTO struct {
	Content string `json:"content" binding:"required"`
}

func (h *Handler) CreateClinicalNote(c *gin.Context) {
	var req createNoteDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	note, err := h.svc.CreateClinicalNote(c.Request.Context(), service.CreateClinicalNoteRequest{
		ProfessionalID: req.ProfessionalID,
		ClientID:       req.ClientID,
		AppointmentID:  req.AppointmentID,
		Content:        req.Content,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, note)
}

func (h *Handler) ListClinicalNotes(c *gin.Context) {
	// Obtenemos params del Query String (?client_id=1&professional_id=2)
	clientIDStr := c.Query("client_id")
	profIDStr := c.Query("professional_id")

	if clientIDStr == "" || profIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_id y professional_id son requeridos"})
		return
	}

	clientID, err1 := strconv.ParseInt(clientIDStr, 10, 64)
	profID, err2 := strconv.ParseInt(profIDStr, 10, 64)

	if err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "IDs inválidos"})
		return
	}

	notes, err := h.svc.ListClinicalNotes(c.Request.Context(), clientID, profID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, notes)
}

func (h *Handler) UpdateClinicalNote(c *gin.Context) {

	idStr := c.Param("id")
	noteID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de nota inválido"})
		return
	}

	var req updateNoteDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	note, err := h.svc.UpdateClinicalNote(c.Request.Context(), service.UpdateClinicalNoteRequest{
		NoteID:  noteID,
		Content: req.Content,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, note)
}
