package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/luciluz/psiconexo/internal/db"
	"github.com/luciluz/psiconexo/internal/service"
)

type Handler struct {
	svc *service.Service
}

func NewHandler(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

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
	Date           string `json:"date" binding:"required"`
	StartTime      string `json:"start_time" binding:"required"`
	Duration       int    `json:"duration" binding:"required,gt=0"`
}

type createRecurringSlotDTO struct {
	PsychologistID int64  `json:"psychologist_id" binding:"required"`
	PatientID      int64  `json:"patient_id" binding:"required"`
	DayOfWeek      int    `json:"day_of_week" binding:"required,min=1,max=7"`
	StartTime      string `json:"start_time" binding:"required"`
	Duration       int    `json:"duration" binding:"required,gt=0"`
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
	})

	if err != nil {
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

func (h *Handler) ListAppointments(c *gin.Context) {

	psyIDStr := c.Query("psychologist_id")
	startStr := c.Query("start_date")
	endStr := c.Query("end_date")

	if psyIDStr == "" || startStr == "" || endStr == "" {
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

func (h *Handler) CreateRecurringSlot(c *gin.Context) {

	var req createRecurringSlotDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	slot, err := h.svc.CreateRecurringSlot(c.Request.Context(), service.CreateRecurringSlotRequest{
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

	c.JSON(http.StatusCreated, slot)
}
