package service_test

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq" // Importante: Driver de Postgres
	"github.com/luciluz/psiconexo/internal/db"
	"github.com/luciluz/psiconexo/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB conecta a la DB de Docker y devuelve el servicio listo + función de limpieza
func setupTestDB(t *testing.T) (*service.Service, func()) {
	// Usamos la misma string de conexión que en tu docker-compose
	connStr := os.Getenv("DB_SOURCE")
	if connStr == "" {
		connStr = "postgres://psico_user:psico_password@localhost:5432/psiconexo?sslmode=disable"
	}

	conn, err := sql.Open("postgres", connStr)
	require.NoError(t, err)

	// Ping para asegurar que Docker está arriba
	err = conn.Ping()
	require.NoError(t, err, "Asegúrate de que 'docker compose up' esté corriendo")

	// Limpieza inicial (orden inverso a las FK para no romper constraints)
	cleanQuery := `
		TRUNCATE TABLE appointments, recurring_rules, clinical_notes, 
		schedule_configs, professional_settings, clients, professionals 
		RESTART IDENTITY CASCADE;
	`
	_, err = conn.Exec(cleanQuery)
	require.NoError(t, err)

	queries := db.New(conn)
	svc := service.NewService(queries, conn)

	// Devolvemos el servicio y una función para cerrar la conexión al final
	return svc, func() {
		conn.Close()
	}
}

// 1. TEST DE CONFIGURACIÓN AVANZADA (Settings)
func TestProfessionalSettings(t *testing.T) {
	svc, teardown := setupTestDB(t)
	defer teardown()
	ctx := context.Background()

	// A. Crear Profesional
	prof, err := svc.CreateProfessional(ctx, service.CreateProfessionalRequest{
		Name:  "Lic. Test Settings",
		Email: "settings@test.com",
	})
	require.NoError(t, err)

	// B. Crear/Actualizar Configuración
	maxDaily := 6
	req := service.UpdateSettingsRequest{
		ProfessionalID:         prof.ID,
		DefaultDurationMinutes: 45,
		BufferMinutes:          15,
		TimeIncrementMinutes:   60, // Modo Tetris
		MinBookingNoticeHours:  48,
		MaxDailyAppointments:   &maxDaily,
	}

	settings, err := svc.UpdateSettings(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, int32(45), settings.DefaultDurationMinutes.Int32)
	assert.Equal(t, int32(15), settings.BufferMinutes.Int32)

	// C. Leer Configuración
	readSettings, err := svc.GetSettings(ctx, prof.ID)
	require.NoError(t, err)
	assert.Equal(t, int32(6), readSettings.MaxDailyAppointments.Int32)
}

// 2. TEST DE NOTAS CLÍNICAS (Privacidad)
func TestClinicalNotes(t *testing.T) {
	svc, teardown := setupTestDB(t)
	defer teardown()
	ctx := context.Background()

	// A. Setup (Pro + Client)
	prof, _ := svc.CreateProfessional(ctx, service.CreateProfessionalRequest{Name: "Doc", Email: "doc@test.com"})
	client, _ := svc.CreateClient(ctx, service.CreateClientRequest{Name: "Paciente", ProfessionalID: prof.ID})

	// B. Crear Nota
	content := "U2FsdGVkX1+... (Contenido Encriptado)"
	note, err := svc.CreateClinicalNote(ctx, service.CreateClinicalNoteRequest{
		ProfessionalID: prof.ID,
		ClientID:       client.ID,
		Content:        content,
	})
	require.NoError(t, err)
	assert.NotZero(t, note.ID)

	// C. Listar Notas
	notes, err := svc.ListClinicalNotes(ctx, client.ID, prof.ID)
	require.NoError(t, err)
	require.Len(t, notes, 1)
	assert.Equal(t, content, notes[0].Content)
	assert.True(t, notes[0].IsEncrypted.Bool, "Debería marcarse como encriptada por defecto")

	// D. Actualizar Nota
	updatedContent := "Nuevo contenido encriptado"
	updatedNote, err := svc.UpdateClinicalNote(ctx, service.UpdateClinicalNoteRequest{
		NoteID:  note.ID,
		Content: updatedContent,
	})
	require.NoError(t, err)
	assert.Equal(t, updatedContent, updatedNote.Content)
}

// 3. TEST DE AGENDA CON PRECIOS Y NOTAS
func TestAppointmentFullFlow(t *testing.T) {
	svc, teardown := setupTestDB(t)
	defer teardown()
	ctx := context.Background()

	// A. Setup
	prof, _ := svc.CreateProfessional(ctx, service.CreateProfessionalRequest{Name: "Doc Agenda", Email: "agenda@test.com"})
	client, _ := svc.CreateClient(ctx, service.CreateClientRequest{Name: "Pac Agenda", ProfessionalID: prof.ID})

	// B. Crear Turno con Precio y Nota
	targetDate := time.Now().Add(24 * time.Hour).Truncate(24 * time.Hour) // Mañana
	req := service.CreateAppointmentRequest{
		ProfessionalID: prof.ID,
		ClientID:       client.ID,
		Date:           targetDate,
		StartTime:      "10:00",
		Duration:       50,
		Price:          15000.50,           // Float
		Notes:          "Paga en efectivo", // String
	}

	appt, err := svc.CreateAppointment(ctx, req)
	require.NoError(t, err)

	// C. Validaciones
	// Postgres guarda Price como Numeric/Decimal, sqlc lo trae como NullString.
	// Verificamos que guardó el string formateado correctamente.
	assert.Equal(t, "15000.50", appt.Price.String)
	assert.Equal(t, "Paga en efectivo", appt.Notes.String)
	assert.Equal(t, "scheduled", appt.Status.String) // Default value check

	// D. Verificar Colisión (CheckAvailability)
	// Intentamos crear otro turno a las 10:30 (solapado con el de 10:00-10:50)
	reqCollision := service.CreateAppointmentRequest{
		ProfessionalID: prof.ID,
		ClientID:       client.ID,
		Date:           targetDate,
		StartTime:      "10:30",
		Duration:       50,
	}
	_, errCol := svc.CreateAppointment(ctx, reqCollision)
	assert.Error(t, errCol, "Debería fallar por colisión de horario")
	assert.Contains(t, errCol.Error(), "colisiona con un turno")
}
