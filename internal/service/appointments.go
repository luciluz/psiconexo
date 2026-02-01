package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/luciluz/psiconexo/internal/db"
)

type CreateAppointmentRequest struct {
	PsychologistID int64
	PatientID      int64
	Date           time.Time
	StartTime      string
	Duration       int
}

type CreateRecurringSlotRequest struct {
	PsychologistID int64
	PatientID      int64
	DayOfWeek      int
	StartTime      string
	Duration       int
}

func (s *Service) CheckAvailability(ctx context.Context, psyID int64, date time.Time, newStartStr string, duration int) error {
	existingAppts, err := s.queries.GetDayAppointments(ctx, db.GetDayAppointmentsParams{
		PsychologistID: psyID,
		Date:           date,
	})
	if err != nil {
		return fmt.Errorf("error obteniendo agenda del día: %w", err)
	}

	layout := "15:04"
	newStart, err := time.Parse(layout, newStartStr)
	if err != nil {
		return fmt.Errorf("formato de hora inválido (use HH:MM): %w", err)
	}
	newEnd := newStart.Add(time.Duration(duration) * time.Minute)

	for _, appt := range existingAppts {

		existStart, _ := time.Parse(layout, appt.StartTime)
		existEnd := existStart.Add(time.Duration(appt.DurationMinutes) * time.Minute)

		if newStart.Before(existEnd) && newEnd.After(existStart) {
			return fmt.Errorf("horario no disponible: colisiona con un turno de %s a %s",
				appt.StartTime, existEnd.Format(layout))
		}
	}

	return nil
}

func (s *Service) CreateAppointment(ctx context.Context, req CreateAppointmentRequest) (*db.Appointment, error) {
	if err := s.CheckAvailability(ctx, req.PsychologistID, req.Date, req.StartTime, req.Duration); err != nil {
		return nil, err
	}

	appt, err := s.queries.CreateAppointment(ctx, db.CreateAppointmentParams{
		PsychologistID:    req.PsychologistID,
		PatientID:         req.PatientID,
		Date:              req.Date,
		StartTime:         req.StartTime,
		DurationMinutes:   int64(req.Duration),
		RescheduledFromID: sql.NullInt64{Valid: false},
	})

	if err != nil {
		return nil, err
	}

	return &appt, nil
}

func (s *Service) ListAppointments(ctx context.Context, psyID int64, start, end time.Time) ([]db.ListAppointmentsInDateRangeRow, error) {
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

func (s *Service) CreateRecurringSlot(ctx context.Context, req CreateRecurringSlotRequest) (*db.RecurringSlot, error) {

	slot, err := s.queries.CreateRecurringSlot(ctx, db.CreateRecurringSlotParams{
		PsychologistID:  req.PsychologistID,
		PatientID:       req.PatientID,
		DayOfWeek:       int64(req.DayOfWeek),
		StartTime:       req.StartTime,
		DurationMinutes: int64(req.Duration),
	})

	if err != nil {
		return nil, fmt.Errorf("error guardando horario fijo: %w", err)
	}

	return &slot, nil
}

func (s *Service) ListRecurringSlots(ctx context.Context, psyID int64) ([]db.ListRecurringSlotsRow, error) {
	slots, err := s.queries.ListRecurringSlots(ctx, psyID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo horarios fijos: %w", err)
	}
	return slots, nil
}
