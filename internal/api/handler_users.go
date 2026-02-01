package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/luciluz/psiconexo/internal/db"
	"github.com/luciluz/psiconexo/internal/service"
)

// DTOs actualizados
type createProfessionalDTO struct {
	Name                    string `json:"name" binding:"required"`
	Email                   string `json:"email" binding:"required,email"`
	Phone                   string `json:"phone"`
	CancellationWindowHours int    `json:"cancellation_window_hours"`
}

type createClientDTO struct {
	Name           string `json:"name" binding:"required"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	ProfessionalID int64  `json:"professional_id" binding:"required"`
}

// --- Handlers Profesionales ---

func (h *Handler) CreateProfessional(c *gin.Context) {
	var req createProfessionalDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	prof, err := h.svc.CreateProfessional(c.Request.Context(), service.CreateProfessionalRequest{
		Name:                    req.Name,
		Email:                   req.Email,
		Phone:                   req.Phone,
		CancellationWindowHours: req.CancellationWindowHours,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, prof)
}

func (h *Handler) ListProfessionals(c *gin.Context) {
	profs, err := h.svc.ListProfessionals(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, profs)
}

// --- Handlers Clientes ---

func (h *Handler) CreateClient(c *gin.Context) {
	var req createClientDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	client, err := h.svc.CreateClient(c.Request.Context(), service.CreateClientRequest{
		Name:           req.Name,
		Email:          req.Email,
		Phone:          req.Phone,
		ProfessionalID: req.ProfessionalID,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, client)
}

func (h *Handler) ListClients(c *gin.Context) {
	var req struct {
		ProfessionalID int64 `form:"professional_id" binding:"required"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "professional_id es inválido o requerido"})
		return
	}

	clients, err := h.svc.ListClients(c.Request.Context(), req.ProfessionalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch clients"})
		return
	}

	if clients == nil {
		clients = []db.Client{}
	}

	c.JSON(http.StatusOK, clients)
}
