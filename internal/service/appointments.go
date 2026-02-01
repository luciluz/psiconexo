package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/luciluz/psiconexo/internal/db"
)

type CreateAppointmentRequest struct {
	ProfessionalID int64
	ClientID       int64
	Date           time.Time
	StartTime      string
	Duration       int
	Price          float64
	Notes          string // <--- NUEVO CAMPO
}

type CreateRecurringRuleRequest struct {
	ProfessionalID int64
	ClientID       int64
	DayOfWeek      int
	StartTime      string
	Duration       int
	Price          float64
	StartDate      time.Time
	// No agregamos Notes a la regla recurrente por ahora
}

// CheckAvailability sin cambios...
func (s *Service) CheckAvailability(ctx context.Context, profID int64, date time.Time, newStartStr string, duration int) error {
	// Nota: Asegúrate que sqlc generó el nombre 'Date' o 'Column2'. Usaremos Date asumiendo regeneración correcta.
	// Si te sigue dando error de Column2, mantenlo como lo tenías.
	existingAppts, err := s.queries.GetDayAppointments(ctx, db.GetDayAppointmentsParams{
		ProfessionalID: profID,
		Column2:        date, // <--- Ajustar según tu sqlc generado
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
	// 1. Verificar disponibilidad
	if err := s.CheckAvailability(ctx, req.ProfessionalID, req.Date, req.StartTime, req.Duration); err != nil {
		return nil, err
	}

	priceStr := fmt.Sprintf("%.2f", req.Price)

	// 2. Insertar turno
	appt, err := s.queries.CreateAppointment(ctx, db.CreateAppointmentParams{
		ProfessionalID:    req.ProfessionalID,
		ClientID:          req.ClientID,
		Date:              req.Date,
		StartTime:         req.StartTime,
		DurationMinutes:   int32(req.Duration),
		Price:             sql.NullString{String: priceStr, Valid: true},
		Notes:             sql.NullString{String: req.Notes, Valid: req.Notes != ""}, // <--- NUEVO: Pasamos la nota
		RescheduledFromID: sql.NullInt64{Valid: false},
		RecurringRuleID:   sql.NullInt64{Valid: false},
	})

	if err != nil {
		return nil, err
	}

	return &appt, nil
}

func (s *Service) ListAppointments(ctx context.Context, profID int64, start, end time.Time) ([]db.ListAppointmentsInDateRangeRow, error) {
	// Revisa nombres de parametros generados (Date vs Column2)
	appts, err := s.queries.ListAppointmentsInDateRange(ctx, db.ListAppointmentsInDateRangeParams{
		ProfessionalID: profID,
		Column2:        start,
		Column3:        end,
	})
	if err != nil {
		return nil, fmt.Errorf("error obteniendo agenda: %w", err)
	}
	return appts, nil
}

func (s *Service) CreateRecurringRule(ctx context.Context, req CreateRecurringRuleRequest) (*db.RecurringRule, error) {
	// Sin cambios en la lógica, solo en la materialización abajo
	priceStr := fmt.Sprintf("%.2f", req.Price)

	var startDate sql.NullTime
	if !req.StartDate.IsZero() {
		startDate = sql.NullTime{Time: req.StartDate, Valid: true}
	}

	rule, err := s.queries.CreateRecurringRule(ctx, db.CreateRecurringRuleParams{
		ProfessionalID:  req.ProfessionalID,
		ClientID:        req.ClientID,
		DayOfWeek:       int32(req.DayOfWeek),
		StartTime:       req.StartTime,
		DurationMinutes: int32(req.Duration),
		Price:           sql.NullString{String: priceStr, Valid: true},
		StartDate:       startDate,
	})

	if err != nil {
		return nil, fmt.Errorf("error guardando regla recurrente: %w", err)
	}

	if err := s.generateFutureAppointments(ctx, &rule, 8); err != nil {
		log.Printf("Error generando turnos futuros para regla %d: %v", rule.ID, err)
	}

	return &rule, nil
}

// ListRecurringRules sin cambios...
func (s *Service) ListRecurringRules(ctx context.Context, profID int64) ([]db.ListRecurringRulesRow, error) {
	rules, err := s.queries.ListRecurringRules(ctx, profID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo reglas recurrentes: %w", err)
	}
	return rules, nil
}

// --- MATERIALIZACIÓN ---

func (s *Service) generateFutureAppointments(ctx context.Context, rule *db.RecurringRule, weeksAhead int) error {
	// ... Lógica de fechas sin cambios ...
	targetDayOfWeek := time.Weekday(rule.DayOfWeek)
	if rule.DayOfWeek == 7 {
		targetDayOfWeek = time.Sunday
	}
	currentDate := time.Now()
	if rule.StartDate.Valid && rule.StartDate.Time.After(currentDate) {
		currentDate = rule.StartDate.Time
	}

	for i := 0; i < weeksAhead; i++ {
		weekStart := currentDate.AddDate(0, 0, i*7)
		daysUntil := int(targetDayOfWeek) - int(weekStart.Weekday())
		if daysUntil < 0 {
			daysUntil += 7
		}
		targetDate := weekStart.AddDate(0, 0, daysUntil)

		if targetDate.Before(time.Now().Truncate(24 * time.Hour)) {
			continue
		}

		if err := s.CheckAvailability(ctx, rule.ProfessionalID, targetDate, rule.StartTime, int(rule.DurationMinutes)); err != nil {
			log.Printf("Saltando generación turno recurrente para %s: Ocupado", targetDate.Format("2006-01-02"))
			continue
		}

		price := rule.Price

		// NUEVO: La query CreateAppointment ahora exige el campo Notes.
		// Como es un turno automático generado por regla, va vacío (NULL).
		_, err := s.queries.CreateAppointment(ctx, db.CreateAppointmentParams{
			ProfessionalID:    rule.ProfessionalID,
			ClientID:          rule.ClientID,
			Date:              targetDate,
			StartTime:         rule.StartTime,
			DurationMinutes:   rule.DurationMinutes,
			Price:             price,
			Notes:             sql.NullString{Valid: false}, // <--- NUEVO: Nota vacía
			RescheduledFromID: sql.NullInt64{Valid: false},
			RecurringRuleID:   sql.NullInt64{Int64: rule.ID, Valid: true},
		})

		if err != nil {
			log.Printf("ℹ️ Turno ya existía o error db para %s: %v", targetDate.Format("2006-01-02"), err)
		}
	}

	return nil
}
