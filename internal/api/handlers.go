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

func (h *Handler) ListPsychologists(c *gin.Context) {
	appts, err := h.svc.ListPsychologists(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, appts)
}

// ListAppointments maneja la petición GET /appointments
func (h *Handler) ListAppointments(c *gin.Context) {
	// 1. Leer parámetros de la URL (Query Params)
	// Ejemplo de URL: /appointments?psychologist_id=1&start_date=2026-01-20&end_date=2026-01-27

	psyIDStr := c.Query("psychologist_id")
	startStr := c.Query("start_date")
	endStr := c.Query("end_date")

	if psyIDStr == "" || startStr == "" || endStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "faltan parámetros requeridos: psychologist_id, start_date, end_date"})
		return
	}

	// 2. Convertir string a int64 (ID)
	// (Necesitamos strconv para esto, asegúrate de importarlo arriba si Go no lo hace solo)
	// Ojo: En Go moderno, Gin no convierte tipos automáticamente en Query Params tan fácil sin structs,
	// así que haremos una conversión manual rápida o usaremos BindQuery.
	// Para simplificar y ser explícitos:

	var req struct {
		PsychologistID int64  `form:"psychologist_id" binding:"required"`
		StartDate      string `form:"start_date" binding:"required"`
		EndDate        string `form:"end_date" binding:"required"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "parametros inválidos: " + err.Error()})
		return
	}

	// 3. Convertir fechas
	layout := "2006-01-02"
	start, err1 := time.Parse(layout, req.StartDate)
	end, err2 := time.Parse(layout, req.EndDate)

	if err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "formato de fecha inválido (use YYYY-MM-DD)"})
		return
	}

	// 4. Llamar al servicio
	appts, err := h.svc.ListAppointments(c.Request.Context(), req.PsychologistID, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, appts)
}
