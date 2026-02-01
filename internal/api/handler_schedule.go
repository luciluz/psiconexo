package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/luciluz/psiconexo/internal/db"
	"github.com/luciluz/psiconexo/internal/service"
)

type updateScheduleDTO struct {
	ProfessionalID int64 `json:"professional_id" binding:"required"`
	Blocks         []struct {
		DayOfWeek int    `json:"day_of_week" binding:"required,min=1,max=7"`
		StartTime string `json:"start_time" binding:"required"`
		EndTime   string `json:"end_time" binding:"required"`
	} `json:"blocks" binding:"required"`
}

func (h *Handler) UpdateSchedule(c *gin.Context) {
	var req updateScheduleDTO

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	serviceBlocks := make([]service.ScheduleBlock, len(req.Blocks))
	for i, b := range req.Blocks {
		serviceBlocks[i] = service.ScheduleBlock{
			DayOfWeek: b.DayOfWeek,
			StartTime: b.StartTime,
			EndTime:   b.EndTime,
		}
	}

	err := h.svc.UpdateSchedule(c.Request.Context(), service.UpdateScheduleRequest{
		ProfessionalID: req.ProfessionalID,
		Blocks:         serviceBlocks,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update schedule: " + err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func (h *Handler) ListSchedule(c *gin.Context) {
	var req struct {
		ProfessionalID int64 `form:"professional_id" binding:"required"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "professional_id requerido"})
		return
	}

	schedule, err := h.svc.ListSchedule(c.Request.Context(), req.ProfessionalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if schedule == nil {
		schedule = []db.ScheduleConfig{}
	}

	c.JSON(http.StatusOK, schedule)
}
