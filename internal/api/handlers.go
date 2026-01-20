package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/luciluz/psiconexo/internal/service"
)

// Handler agrupa todos los endpoints.
// Usa inyección de dependencias para acceder al Service.
type Handler struct {
	svc *service.Service
}

func NewHandler(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

// --- DTOs (Data Transfer Objects) ---
// Definimos structs locales para leer el JSON.
// No usamos los del service directamente para desacoplar el formato de entrada (string) del tipo de dato interno (time.Time).

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

type createAppointmentDTO struct {
	PsychologistID int64  `json:"psychologist_id" binding:"required"`
	PatientID      int64  `json:"patient_id" binding:"required"`
	Date           string `json:"date" binding:"required"`       // Recibimos "YYYY-MM-DD"
	StartTime      string `json:"start_time" binding:"required"` // "HH:MM"
	Duration       int    `json:"duration" binding:"required"`
}

// --- Métodos Handler ---

func (h *Handler) CreatePsychologist(c *gin.Context) {
	var req createPsychologistDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Llamada al servicio
	psy, err := h.svc.CreatePsychologist(c.Request.Context(), service.CreatePsychologistRequest{
		Name:                    req.Name,
		Email:                   req.Email,
		Phone:                   req.Phone,
		CancellationWindowHours: req.CancellationWindowHours,
	})

	if err != nil {
		// Aquí podrías diferenciar errores (ej: duplicado vs error interno)
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

func (h *Handler) CreateAppointment(c *gin.Context) {
	var req createAppointmentDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. Convertir fecha de string a time.Time
	parsedDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "formato de fecha inválido, use YYYY-MM-DD"})
		return
	}

	// 2. Llamar al servicio (que hará la validación de colisiones)
	appt, err := h.svc.CreateAppointment(c.Request.Context(), service.CreateAppointmentRequest{
		PsychologistID: req.PsychologistID,
		PatientID:      req.PatientID,
		Date:           parsedDate,
		StartTime:      req.StartTime,
		Duration:       req.Duration,
	})

	if err != nil {
		// Si es error de colisión, idealmente devolveríamos 409 Conflict.
		// Por ahora devolvemos 400/500 genérico.
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, appt)
}
