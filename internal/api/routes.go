package api

import "github.com/gin-gonic/gin"

func NewRouter(h *Handler) *gin.Engine {
	r := gin.Default()
	r.Use(CorsMiddleware())

	v1 := r.Group("/api/v1")
	{
		// Usuarios (Profesionales y Clientes)
		v1.POST("/professionals", h.CreateProfessional)
		v1.GET("/professionals", h.ListProfessionals)

		v1.POST("/clients", h.CreateClient)
		v1.GET("/clients", h.ListClients)

		// Agenda (Eventual y Materializada)
		v1.POST("/appointments", h.CreateAppointment)
		v1.GET("/appointments", h.ListAppointments)

		// Reglas de Recurrencia (Contratos fijos)
		v1.POST("/recurring-rules", h.CreateRecurringRule)
		v1.GET("/recurring-rules", h.ListRecurringRules)

		// Bloques de trabajo (Disponibilidad / Configuración)
		v1.POST("/schedule", h.UpdateSchedule)
		v1.GET("/schedule", h.ListSchedule)

		// Configuración Avanzada (Settings)
		v1.PUT("/settings", h.UpdateSettings)
		v1.GET("/settings", h.GetSettings)

		// Notas Clínicas (Historia Clínica)
		v1.POST("/clinical-notes", h.CreateClinicalNote)
		v1.GET("/clinical-notes", h.ListClinicalNotes)
		v1.PUT("/clinical-notes/:id", h.UpdateClinicalNote) // Nota el :id en la URL
	}

	return r
}
