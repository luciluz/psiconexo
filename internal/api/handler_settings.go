package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/luciluz/psiconexo/internal/service"
)

type settingsDTO struct {
	ProfessionalID         int64 `json:"professional_id" binding:"required"`
	DefaultDurationMinutes int   `json:"default_duration_minutes"`
	BufferMinutes          int   `json:"buffer_minutes"`
	TimeIncrementMinutes   int   `json:"time_increment_minutes"`
	MinBookingNoticeHours  int   `json:"min_booking_notice_hours"`
	MaxDailyAppointments   *int  `json:"max_daily_appointments"`
}

func (h *Handler) UpdateSettings(c *gin.Context) {
	var req settingsDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	settings, err := h.svc.UpdateSettings(c.Request.Context(), service.UpdateSettingsRequest{
		ProfessionalID:         req.ProfessionalID,
		DefaultDurationMinutes: req.DefaultDurationMinutes,
		BufferMinutes:          req.BufferMinutes,
		TimeIncrementMinutes:   req.TimeIncrementMinutes,
		MinBookingNoticeHours:  req.MinBookingNoticeHours,
		MaxDailyAppointments:   req.MaxDailyAppointments,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}

func (h *Handler) GetSettings(c *gin.Context) {
	var req struct {
		ProfessionalID int64 `form:"professional_id" binding:"required"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "professional_id es requerido"})
		return
	}

	settings, err := h.svc.GetSettings(c.Request.Context(), req.ProfessionalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if settings == nil {
		c.JSON(http.StatusOK, gin.H{
			"professional_id":          req.ProfessionalID,
			"default_duration_minutes": 50, // Default hardcodeado por seguridad UX
			"buffer_minutes":           0,
			"time_increment_minutes":   30,
			"min_booking_notice_hours": 24,
			"max_daily_appointments":   nil,
		})
		return
	}

	c.JSON(http.StatusOK, settings)
}
