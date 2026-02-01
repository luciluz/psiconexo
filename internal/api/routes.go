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

		// Agenda (Eventual y Materializada)
		v1.POST("/appointments", h.CreateAppointment)
		v1.GET("/appointments", h.ListAppointments)

		// Reglas de Recurrencia (Contratos fijos)
		v1.POST("/recurring-rules", h.CreateRecurringRule)
		v1.GET("/recurring-rules", h.ListRecurringRules)

		// Bloques de trabajo (Disponibilidad / Configuración)
		v1.POST("/schedule", h.UpdateSchedule)
		v1.GET("/schedule", h.ListSchedule)
	}

	return r
}
