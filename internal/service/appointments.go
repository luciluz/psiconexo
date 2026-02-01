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
	PsychologistID int64
	PatientID      int64
	Date           time.Time
	StartTime      string
	Duration       int
}

// Renombrado para coincidir con la nueva lógica
type CreateRecurringRuleRequest struct {
	PsychologistID int64
	PatientID      int64
	DayOfWeek      int // 1=Lunes, ... 7=Domingo
	StartTime      string
	Duration       int
}

// CheckAvailability verifica si hay colisiones en la tabla appointments.
// Al usar materialización, esta función sirve tanto para turnos manuales como recurrentes.
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

		// Lógica de solapamiento
		if newStart.Before(existEnd) && newEnd.After(existStart) {
			return fmt.Errorf("horario no disponible: colisiona con un turno de %s a %s",
				appt.StartTime, existEnd.Format(layout))
		}
	}

	return nil
}

func (s *Service) CreateAppointment(ctx context.Context, req CreateAppointmentRequest) (*db.Appointment, error) {
	// 1. Verificamos disponibilidad antes de intentar insertar
	if err := s.CheckAvailability(ctx, req.PsychologistID, req.Date, req.StartTime, req.Duration); err != nil {
		return nil, err
	}

	// 2. Insertamos el turno manual (RecurringRuleID es NULL)
	appt, err := s.queries.CreateAppointment(ctx, db.CreateAppointmentParams{
		PsychologistID:    req.PsychologistID,
		PatientID:         req.PatientID,
		Date:              req.Date,
		StartTime:         req.StartTime,
		DurationMinutes:   int64(req.Duration),
		RescheduledFromID: sql.NullInt64{Valid: false},
		RecurringRuleID:   sql.NullInt64{Valid: false}, // Es manual
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

// CreateRecurringRule guarda la regla Y genera los turnos futuros (Materialización)
func (s *Service) CreateRecurringRule(ctx context.Context, req CreateRecurringRuleRequest) (*db.RecurringRule, error) {

	// 1. Guardar la Regla Maestra
	rule, err := s.queries.CreateRecurringRule(ctx, db.CreateRecurringRuleParams{
		PsychologistID:  req.PsychologistID,
		PatientID:       req.PatientID,
		DayOfWeek:       int64(req.DayOfWeek),
		StartTime:       req.StartTime,
		DurationMinutes: int64(req.Duration),
	})

	if err != nil {
		return nil, fmt.Errorf("error guardando regla recurrente: %w", err)
	}

	// 2. Materializar: Generar turnos para las próximas 8 semanas
	// Lo hacemos en una goroutine para no bloquear la respuesta si tarda un poco,
	// o síncrono si queremos asegurar que se creen ya. Para MVP, síncrono está bien.
	if err := s.generateFutureAppointments(ctx, &rule, 8); err != nil {
		// Logueamos el error pero no fallamos la request principal, la regla ya se creó.
		// En un sistema real, usarías un worker background.
		log.Printf("Error generando turnos futuros para regla %d: %v", rule.ID, err)
	}

	return &rule, nil
}

func (s *Service) ListRecurringRules(ctx context.Context, psyID int64) ([]db.ListRecurringRulesRow, error) {
	rules, err := s.queries.ListRecurringRules(ctx, psyID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo reglas recurrentes: %w", err)
	}
	return rules, nil
}

// --- LÓGICA DE MATERIALIZACIÓN ---

func (s *Service) generateFutureAppointments(ctx context.Context, rule *db.RecurringRule, weeksAhead int) error {
	targetDayOfWeek := time.Weekday(rule.DayOfWeek)
	// Nota: En Go time.Sunday=0, time.Monday=1.
	// Asegúrate que tu frontend manda 1 para Lunes. Si manda 0 para lunes, ajusta aquí.
	// Asumimos standard ISO: 1=Lunes... 7=Domingo.
	// Si tu DB usa 1=Lunes y Go usa 1=Lunes, estamos bien. Pero Go usa 0=Domingo.
	// Ajuste pequeño si rule.DayOfWeek es 7 (Domingo en DB) -> 0 (Domingo en Go)
	if rule.DayOfWeek == 7 {
		targetDayOfWeek = time.Sunday
	}

	currentDate := time.Now()

	// Iteramos X semanas hacia adelante
	for i := 0; i < weeksAhead; i++ {
		// Calculamos la fecha de la próxima ocurrencia
		// Avanzamos 'i' semanas
		weekStart := currentDate.AddDate(0, 0, i*7)

		// Buscamos el día correcto de esa semana
		daysUntil := int(targetDayOfWeek) - int(weekStart.Weekday())
		if daysUntil < 0 {
			daysUntil += 7
		}
		targetDate := weekStart.AddDate(0, 0, daysUntil)

		// Si la fecha calculada es hoy pero la hora ya pasó, o es ayer, saltamos
		if targetDate.Before(time.Now().Truncate(24 * time.Hour)) {
			continue
		}

		// 1. Verificar Disponibilidad
		// Si ese día específico está ocupado (ej: feriado o turno manual previo),
		// saltamos la creación (no forzamos colisión)
		if err := s.CheckAvailability(ctx, rule.PsychologistID, targetDate, rule.StartTime, int(rule.DurationMinutes)); err != nil {
			log.Printf("Saltando generación turno recurrente para %s: Ocupado", targetDate.Format("2006-01-02"))
			continue
		}

		// 2. Insertar Turno Materializado
		_, err := s.queries.CreateAppointment(ctx, db.CreateAppointmentParams{
			PsychologistID:    rule.PsychologistID,
			PatientID:         rule.PatientID,
			Date:              targetDate,
			StartTime:         rule.StartTime,
			DurationMinutes:   rule.DurationMinutes,
			RescheduledFromID: sql.NullInt64{Valid: false},
			RecurringRuleID:   sql.NullInt64{Int64: rule.ID, Valid: true}, // LINK IMPORTANTE
		})

		if err != nil {
			// Si falla por Unique Constraint (ya existe), lo ignoramos silenciosamente
			log.Printf("ℹTurno ya existía o error db para %s: %v", targetDate.Format("2006-01-02"), err)
		}
	}

	return nil
}
