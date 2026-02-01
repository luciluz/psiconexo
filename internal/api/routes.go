package api

import "github.com/gin-gonic/gin"

func NewRouter(h *Handler) *gin.Engine {
	r := gin.Default()
	r.Use(CorsMiddleware())

	v1 := r.Group("/api/v1")
	{
		// Usuarios
		v1.POST("/psychologists", h.CreatePsychologist)
		v1.GET("/psychologists", h.ListPsychologists)
		v1.POST("/patients", h.CreatePatient)
		v1.GET("/patients", h.ListPatients)

		// Agenda
		v1.POST("/appointments", h.CreateAppointment)
		v1.GET("/appointments", h.ListAppointments)

		// Horarios fijos
		v1.POST("/recurring-slots", h.CreateRecurringSlot)
		v1.GET("/recurring-slots", h.ListRecurringSlots)

		// Bloques de trabajo
		v1.POST("/schedule", h.UpdateSchedule)
		v1.GET("/schedule", h.ListSchedule)
	}

	return r
}
