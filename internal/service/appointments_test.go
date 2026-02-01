package service_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3" // Importante para que funcione sqlite en memoria
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/luciluz/psiconexo/internal/db"
	"github.com/luciluz/psiconexo/internal/service"
)

func SetupTestDB(t *testing.T) *service.Service {
	// Usamos :memory: para que sea una DB volatil y rápida
	sqldb, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Schema con created_at Y start_date para que coincida con lo que espera sqlc
	schema := `
    CREATE TABLE psychologists (
        id INTEGER PRIMARY KEY, 
        name TEXT, 
        email TEXT, 
        phone TEXT, 
        cancellation_window_hours INT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );
    
    CREATE TABLE patients (
        id INTEGER PRIMARY KEY, 
        name TEXT, 
        psychologist_id INT, 
        email TEXT, 
        phone TEXT, 
        active BOOLEAN DEFAULT TRUE,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );
    
    CREATE TABLE schedule_configs (
        id INTEGER PRIMARY KEY, 
        psychologist_id INT, 
        day_of_week INT, 
        start_time TEXT, 
        end_time TEXT, 
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        UNIQUE(psychologist_id, day_of_week, start_time)
    );

    CREATE TABLE recurring_rules (
        id INTEGER PRIMARY KEY, 
        psychologist_id INT, 
        patient_id INT, 
        day_of_week INT, 
        start_time TEXT, 
        duration_minutes INT, 
        active BOOLEAN DEFAULT TRUE, 
        start_date DATE,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );

    CREATE TABLE appointments (
        id INTEGER PRIMARY KEY, 
        psychologist_id INT, 
        patient_id INT, 
        date DATE, 
        start_time TEXT, 
        duration_minutes INT, 
        status TEXT DEFAULT 'scheduled', 
        rescheduled_from_id INT, 
        recurring_rule_id INT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        UNIQUE(psychologist_id, date, start_time)
    );
    `
	_, err = sqldb.Exec(schema)
	require.NoError(t, err)

	queries := db.New(sqldb)

	// Retornamos el servicio con la conexión
	return service.NewService(queries, sqldb)
}

func TestMaterializationAndCollision(t *testing.T) {
	svc := SetupTestDB(t)
	ctx := context.Background()

	// 1. DATOS PREVIOS: Crear Psicólogo y Paciente
	psy, err := svc.CreatePsychologist(ctx, service.CreatePsychologistRequest{
		Name: "Dr. House", Email: "house@test.com",
	})
	require.NoError(t, err)

	pat, err := svc.CreatePatient(ctx, service.CreatePatientRequest{
		Name: "Jesse Pinkman", Email: "jesse@test.com", PsychologistID: psy.ID,
	})
	require.NoError(t, err)

	// 2. ACCIÓN: Crear Regla Recurrente (Lunes a las 10:00)
	// Importante: Elegimos un día y hora.
	// Nota: 1 = Lunes.
	startTime := "10:00"
	rule, err := svc.CreateRecurringRule(ctx, service.CreateRecurringRuleRequest{
		PsychologistID: psy.ID,
		PatientID:      pat.ID,
		DayOfWeek:      1, // Lunes
		StartTime:      startTime,
		Duration:       60,
	})
	require.NoError(t, err)
	assert.NotZero(t, rule.ID)

	// 3. VERIFICACIÓN 1: Materialización
	// El servicio debió generar turnos para los próximos lunes.
	// Busquemos turnos para los próximos 30 días.
	now := time.Now()
	monthLater := now.AddDate(0, 0, 30)

	appts, err := svc.ListAppointments(ctx, psy.ID, now, monthLater)
	require.NoError(t, err)

	// Debería haber al menos 3 o 4 turnos generados (uno por semana)
	assert.GreaterOrEqual(t, len(appts), 3, "Deberían haberse generado turnos automáticos")

	// Tomamos el primer turno generado para probar colisión
	firstGeneratedAppt := appts[0]
	t.Logf("Turno generado automáticamente: Fecha %v Hora %s", firstGeneratedAppt.Date, firstGeneratedAppt.StartTime)

	// 4. VERIFICACIÓN 2: Colisión (La prueba de fuego)
	// Intentamos crear MANUALMENTE un turno a la misma hora del turno generado.

	_, err = svc.CreateAppointment(ctx, service.CreateAppointmentRequest{
		PsychologistID: psy.ID,
		PatientID:      pat.ID,                  // Puede ser el mismo u otro paciente
		Date:           firstGeneratedAppt.Date, // Usamos la misma fecha que generó el sistema
		StartTime:      startTime,               // Misma hora (10:00)
		Duration:       60,
	})

	// ESTO DEBE FALLAR.
	// Si falla, significa que nuestro sistema de protección funciona.
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "horario no disponible", "El error debe ser de disponibilidad")
}
