package api

import "github.com/gin-gonic/gin"

func NewRouter(h *Handler) *gin.Engine {
	r := gin.Default()
	r.Use(CorsMiddleware())

	// Middleware básico (Logger y Recovery ya vienen en Default)

	// Grupo de rutas API v1
	v1 := r.Group("/api/v1")
	{
		// Usuarios
		v1.POST("/psychologists", h.CreatePsychologist)
		v1.POST("/patients", h.CreatePatient)

		// Agenda
		v1.POST("/appointments", h.CreateAppointment)
		v1.GET("/appointments", h.ListAppointments)
	}

	return r
}
