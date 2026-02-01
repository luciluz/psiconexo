package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/luciluz/psiconexo/internal/service"
)

type createAppointmentDTO struct {
	ProfessionalID int64   `json:"professional_id" binding:"required"`
	ClientID       int64   `json:"client_id" binding:"required"`
	Date           string  `json:"date" binding:"required"`
	StartTime      string  `json:"start_time" binding:"required"`
	Duration       int     `json:"duration" binding:"required,gt=0"`
	Price          float64 `json:"price"` // Nuevo campo opcional (puede ser 0)
}

type createRecurringRuleDTO struct {
	ProfessionalID int64   `json:"professional_id" binding:"required"`
	ClientID       int64   `json:"client_id" binding:"required"`
	DayOfWeek      int     `json:"day_of_week" binding:"required,min=1,max=7"`
	StartTime      string  `json:"start_time" binding:"required"`
	Duration       int     `json:"duration" binding:"required,gt=0"`
	Price          float64 `json:"price"`      // Nuevo campo
	StartDate      string  `json:"start_date"` // Nuevo campo opcional
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
		ProfessionalID: req.ProfessionalID,
		ClientID:       req.ClientID,
		Date:           parsedDate,
		StartTime:      req.StartTime,
		Duration:       req.Duration,
		Price:          req.Price, // Pasamos el precio
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, appt)
}

func (h *Handler) ListAppointments(c *gin.Context) {

	if c.Query("professional_id") == "" || c.Query("start_date") == "" || c.Query("end_date") == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "faltan parámetros requeridos: professional_id, start_date, end_date"})
		return
	}

	var req struct {
		ProfessionalID int64  `form:"professional_id" binding:"required"`
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

	appts, err := h.svc.ListAppointments(c.Request.Context(), req.ProfessionalID, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, appts)
}

func (h *Handler) CreateRecurringRule(c *gin.Context) {
	var req createRecurringRuleDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parseo de StartDate si viene, sino nil/time.Zero
	var startDateParsed time.Time
	if req.StartDate != "" {
		var err error
		startDateParsed, err = time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "start_date inválido"})
			return
		}
	}

	rule, err := h.svc.CreateRecurringRule(c.Request.Context(), service.CreateRecurringRuleRequest{
		ProfessionalID: req.ProfessionalID,
		ClientID:       req.ClientID,
		DayOfWeek:      req.DayOfWeek,
		StartTime:      req.StartTime,
		Duration:       req.Duration,
		Price:          req.Price,
		StartDate:      startDateParsed,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, rule)
}

func (h *Handler) ListRecurringRules(c *gin.Context) {
	var req struct {
		ProfessionalID int64 `form:"professional_id" binding:"required"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rules, err := h.svc.ListRecurringRules(c.Request.Context(), req.ProfessionalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rules)
}
