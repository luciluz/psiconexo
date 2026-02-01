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
	Price          float64 // Nuevo campo
}

type CreateRecurringRuleRequest struct {
	ProfessionalID int64
	ClientID       int64
	DayOfWeek      int
	StartTime      string
	Duration       int
	Price          float64   // Nuevo campo
	StartDate      time.Time // Nuevo campo (opcional)
}

// CheckAvailability verifica si hay colisiones.
func (s *Service) CheckAvailability(ctx context.Context, profID int64, date time.Time, newStartStr string, duration int) error {
	existingAppts, err := s.queries.GetDayAppointments(ctx, db.GetDayAppointmentsParams{
		ProfessionalID: profID,
		Column2:        date,
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

	// Conversión de precio float64 a string para Postgres (DECIMAL)
	// sqlc con lib/pq suele generar string para tipos numeric/decimal
	priceStr := fmt.Sprintf("%.2f", req.Price)

	// 2. Insertar turno
	appt, err := s.queries.CreateAppointment(ctx, db.CreateAppointmentParams{
		ProfessionalID:    req.ProfessionalID,
		ClientID:          req.ClientID,
		Date:              req.Date,
		StartTime:         req.StartTime,
		DurationMinutes:   int32(req.Duration),
		Price:             sql.NullString{String: priceStr, Valid: true},
		RescheduledFromID: sql.NullInt64{Valid: false},
		RecurringRuleID:   sql.NullInt64{Valid: false},
	})

	if err != nil {
		return nil, err
	}

	return &appt, nil
}

func (s *Service) ListAppointments(ctx context.Context, profID int64, start, end time.Time) ([]db.ListAppointmentsInDateRangeRow, error) {
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

	priceStr := fmt.Sprintf("%.2f", req.Price)

	// Manejo de StartDate opcional
	var startDate sql.NullTime
	if !req.StartDate.IsZero() {
		startDate = sql.NullTime{Time: req.StartDate, Valid: true}
	}

	// 1. Guardar la Regla Maestra
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

	// 2. Materializar turnos futuros
	if err := s.generateFutureAppointments(ctx, &rule, 8); err != nil {
		log.Printf("Error generando turnos futuros para regla %d: %v", rule.ID, err)
	}

	return &rule, nil
}

func (s *Service) ListRecurringRules(ctx context.Context, profID int64) ([]db.ListRecurringRulesRow, error) {
	rules, err := s.queries.ListRecurringRules(ctx, profID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo reglas recurrentes: %w", err)
	}
	return rules, nil
}

// --- MATERIALIZACIÓN ---

func (s *Service) generateFutureAppointments(ctx context.Context, rule *db.RecurringRule, weeksAhead int) error {
	targetDayOfWeek := time.Weekday(rule.DayOfWeek)
	if rule.DayOfWeek == 7 {
		targetDayOfWeek = time.Sunday
	}

	currentDate := time.Now()

	// Si la regla tiene fecha de inicio futura, empezamos a contar desde ahí
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

		// Verificamos disponibilidad usando ProfessionalID
		if err := s.CheckAvailability(ctx, rule.ProfessionalID, targetDate, rule.StartTime, int(rule.DurationMinutes)); err != nil {
			log.Printf("Saltando generación turno recurrente para %s: Ocupado", targetDate.Format("2006-01-02"))
			continue
		}

		// Copiamos el precio de la regla al turno (Snapshot)
		// Nota: rule.Price en DB suele ser NullString, lo pasamos tal cual
		price := rule.Price

		_, err := s.queries.CreateAppointment(ctx, db.CreateAppointmentParams{
			ProfessionalID:    rule.ProfessionalID,
			ClientID:          rule.ClientID,
			Date:              targetDate,
			StartTime:         rule.StartTime,
			DurationMinutes:   rule.DurationMinutes,
			Price:             price, // <-- Precio desde la regla
			RescheduledFromID: sql.NullInt64{Valid: false},
			RecurringRuleID:   sql.NullInt64{Int64: rule.ID, Valid: true},
		})

		if err != nil {
			log.Printf("ℹ️ Turno ya existía o error db para %s: %v", targetDate.Format("2006-01-02"), err)
		}
	}

	return nil
}
