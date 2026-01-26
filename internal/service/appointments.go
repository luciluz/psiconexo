package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/luciluz/psiconexo/internal/db"
)

// CreateAppointmentParams es lo que el Frontend nos manda
type CreateAppointmentRequest struct {
	PsychologistID int64
	PatientID      int64
	Date           time.Time // La fecha (sin hora)
	StartTime      string    // "10:00"
	Duration       int       // 60 (minutos)
}

// CheckAvailability verifica si el hueco está libre en memoria
// TODO: se podría hacer la verificación de fechas en SQL?
func (s *Service) CheckAvailability(ctx context.Context, psyID int64, date time.Time, newStartStr string, duration int) error {
	// 1. Traer todos los turnos confirmados de ese día
	existingAppts, err := s.queries.GetDayAppointments(ctx, db.GetDayAppointmentsParams{
		PsychologistID: psyID,
		Date:           date,
	})
	if err != nil {
		return fmt.Errorf("error obteniendo agenda del día: %w", err)
	}

	// 2. Parsear el horario del NUEVO turno
	// Usamos una fecha base arbitraria para poder comparar horas
	layout := "15:04"
	newStart, err := time.Parse(layout, newStartStr)
	if err != nil {
		return fmt.Errorf("formato de hora inválido (use HH:MM): %w", err)
	}
	newEnd := newStart.Add(time.Duration(duration) * time.Minute)

	// 3. Recorrer y comparar (Algoritmo de detección de colisiones)
	for _, appt := range existingAppts {
		// Parsear el horario del turno EXISTENTE
		existStart, _ := time.Parse(layout, appt.StartTime)
		existEnd := existStart.Add(time.Duration(appt.DurationMinutes) * time.Minute)

		// ¿Se superponen?
		// Lógica: (NuevoInicio < ViejoFin) Y (NuevoFin > ViejoInicio)
		if newStart.Before(existEnd) && newEnd.After(existStart) {
			return fmt.Errorf("horario no disponible: colisiona con un turno de %s a %s",
				appt.StartTime, existEnd.Format(layout))
		}
	}

	return nil // Está libre
}

// CreateAppointment orquesta la validación y el guardado
func (s *Service) CreateAppointment(ctx context.Context, req CreateAppointmentRequest) (*db.Appointment, error) {
	// 1. Validar Disponibilidad
	if err := s.CheckAvailability(ctx, req.PsychologistID, req.Date, req.StartTime, req.Duration); err != nil {
		return nil, err
	}

	// 2. Si pasa, guardamos en DB
	appt, err := s.queries.CreateAppointment(ctx, db.CreateAppointmentParams{
		PsychologistID:    req.PsychologistID,
		PatientID:         req.PatientID,
		Date:              req.Date,
		StartTime:         req.StartTime,
		DurationMinutes:   int64(req.Duration),
		RescheduledFromID: sql.NullInt64{Valid: false}, // Es nuevo
	})

	if err != nil {
		return nil, err
	}

	return &appt, nil
}

// ListAppointments devuelve los turnos dentro de un rango de fechas.
func (s *Service) ListAppointments(ctx context.Context, psyID int64, start, end time.Time) ([]db.ListAppointmentsInDateRangeRow, error) {
	// Llamamos a la query que ya definimos en SQL hace días
	appts, err := s.queries.ListAppointmentsInDateRange(ctx, db.ListAppointmentsInDateRangeParams{
		PsychologistID: psyID,
		Date:           start,
		Date_2:         end,
	})
	if err != nil {
		return nil, fmt.Errorf("error obteniendo agenda: %w", err)
	}
	return appts, nil
}
