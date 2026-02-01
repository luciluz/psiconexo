package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/luciluz/psiconexo/internal/service"
)

type createAppointmentDTO struct {
	PsychologistID int64  `json:"psychologist_id" binding:"required"`
	PatientID      int64  `json:"patient_id" binding:"required"`
	Date           string `json:"date" binding:"required"`
	StartTime      string `json:"start_time" binding:"required"`
	Duration       int    `json:"duration" binding:"required,gt=0"`
}

// Renombramos DTO para reflejar que es una Regla
type createRecurringRuleDTO struct {
	PsychologistID int64  `json:"psychologist_id" binding:"required"`
	PatientID      int64  `json:"patient_id" binding:"required"`
	DayOfWeek      int    `json:"day_of_week" binding:"required,min=1,max=7"`
	StartTime      string `json:"start_time" binding:"required"`
	Duration       int    `json:"duration" binding:"required,gt=0"`
}

func (h *Handler) CreateAppointment(c *gin.Context) {
	var req createAppointmentDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	parsedDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "formato de fecha inválido, use YYYY-MM-DD"})
		return
	}

	appt, err := h.svc.CreateAppointment(c.Request.Context(), service.CreateAppointmentRequest{
		PsychologistID: req.PsychologistID,
		PatientID:      req.PatientID,
		Date:           parsedDate,
		StartTime:      req.StartTime,
		Duration:       req.Duration,
		// RecurringRuleID va implícito como nil en el servicio para turnos manuales
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, appt)
}

func (h *Handler) ListAppointments(c *gin.Context) {

	if c.Query("psychologist_id") == "" || c.Query("start_date") == "" || c.Query("end_date") == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "faltan parámetros requeridos: psychologist_id, start_date, end_date"})
		return
	}

	var req struct {
		PsychologistID int64  `form:"psychologist_id" binding:"required"`
		StartDate      string `form:"start_date" binding:"required"`
		EndDate        string `form:"end_date" binding:"required"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "parametros inválidos: " + err.Error()})
		return
	}

	layout := "2006-01-02"
	start, err1 := time.Parse(layout, req.StartDate)
	end, err2 := time.Parse(layout, req.EndDate)

	if err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "formato de fecha inválido (use YYYY-MM-DD)"})
		return
	}

	appts, err := h.svc.ListAppointments(c.Request.Context(), req.PsychologistID, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, appts)
}

// Renombrado: CreateRecurringSlot -> CreateRecurringRule
func (h *Handler) CreateRecurringRule(c *gin.Context) {
	var req createRecurringRuleDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Llamamos al servicio (que internamente creará la regla Y generará los turnos futuros)
	rule, err := h.svc.CreateRecurringRule(c.Request.Context(), service.CreateRecurringRuleRequest{
		PsychologistID: req.PsychologistID,
		PatientID:      req.PatientID,
		DayOfWeek:      req.DayOfWeek,
		StartTime:      req.StartTime,
		Duration:       req.Duration,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, rule)
}

// Renombrado: ListRecurringSlots -> ListRecurringRules
func (h *Handler) ListRecurringRules(c *gin.Context) {
	if c.Query("psychologist_id") == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "falta psychologist_id"})
		return
	}

	var req struct {
		PsychologistID int64 `form:"psychologist_id" binding:"required"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rules, err := h.svc.ListRecurringRules(c.Request.Context(), req.PsychologistID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rules)
}
