package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/luciluz/psiconexo/internal/db"
	"github.com/luciluz/psiconexo/internal/service"
)

type createPsychologistDTO struct {
	Name                    string `json:"name" binding:"required"`
	Email                   string `json:"email" binding:"required,email"`
	Phone                   string `json:"phone"`
	CancellationWindowHours int    `json:"cancellation_window_hours"`
}

type createPatientDTO struct {
	Name           string `json:"name" binding:"required"`
	Email          string `json:"email" binding:"required,email"`
	Phone          string `json:"phone"`
	PsychologistID int64  `json:"psychologist_id" binding:"required"`
}

func (h *Handler) CreatePsychologist(c *gin.Context) {
	var req createPsychologistDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	psy, err := h.svc.CreatePsychologist(c.Request.Context(), service.CreatePsychologistRequest{
		Name:                    req.Name,
		Email:                   req.Email,
		Phone:                   req.Phone,
		CancellationWindowHours: req.CancellationWindowHours,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, psy)
}

func (h *Handler) CreatePatient(c *gin.Context) {
	var req createPatientDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pat, err := h.svc.CreatePatient(c.Request.Context(), service.CreatePatientRequest{
		Name:           req.Name,
		Email:          req.Email,
		Phone:          req.Phone,
		PsychologistID: req.PsychologistID,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, pat)
}

func (h *Handler) ListPsychologists(c *gin.Context) {
	appts, err := h.svc.ListPsychologists(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, appts)
}

func (h *Handler) ListPatients(c *gin.Context) {
	var req struct {
		PsychologistID int64 `form:"psychologist_id" binding:"required"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "psychologist_id es inválido o requerido"})
		return
	}

	patients, err := h.svc.ListPatients(c.Request.Context(), req.PsychologistID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch patients"})
		return
	}

	if patients == nil {
		patients = []db.Patient{}
	}

	c.JSON(http.StatusOK, patients)
}
