-- SECTION: Professionals

-- name: CreateProfessional :one
INSERT INTO professionals (name, email, phone, cancellation_window_hours)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetProfessional :one
SELECT * FROM professionals 
WHERE id = $1 LIMIT 1;

-- name: GetProfessionalByEmail :one
SELECT * FROM professionals 
WHERE email = $1 LIMIT 1;

-- name: ListProfessionals :many
SELECT id, name, email, phone
FROM professionals;


-- SECTION: Professional Settings (NUEVO)

-- name: UpsertProfessionalSettings :one
-- "Upsert": Si existe lo actualiza, si no existe lo crea.
INSERT INTO professional_settings (
    professional_id, default_duration_minutes, buffer_minutes, 
    time_increment_minutes, min_booking_notice_hours, max_daily_appointments
) VALUES (
    $1, $2, $3, $4, $5, $6
)
ON CONFLICT (professional_id) DO UPDATE SET
    default_duration_minutes = EXCLUDED.default_duration_minutes,
    buffer_minutes = EXCLUDED.buffer_minutes,
    time_increment_minutes = EXCLUDED.time_increment_minutes,
    min_booking_notice_hours = EXCLUDED.min_booking_notice_hours,
    max_daily_appointments = EXCLUDED.max_daily_appointments,
    updated_at = NOW()
RETURNING *;

-- name: GetProfessionalSettings :one
SELECT * FROM professional_settings
WHERE professional_id = $1;


-- SECTION: Clients

-- name: CreateClient :one
INSERT INTO clients (name, email, phone, professional_id, active)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListClients :many
SELECT * FROM clients
WHERE professional_id = $1 AND active = TRUE
ORDER BY name;

-- name: GetClient :one
SELECT * FROM clients WHERE id = $1 LIMIT 1;


-- SECTION: Clinical Notes (NUEVO - Privacidad)

-- name: CreateClinicalNote :one
INSERT INTO clinical_notes (professional_id, client_id, appointment_id, content, is_encrypted)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListClinicalNotes :many
-- Traemos las notas de un paciente ordenadas por fecha (más reciente arriba)
SELECT * FROM clinical_notes
WHERE client_id = $1 AND professional_id = $2
ORDER BY created_at DESC;

-- name: UpdateClinicalNote :one
UPDATE clinical_notes
SET content = $1, updated_at = NOW()
WHERE id = $2
RETURNING *;


-- SECTION: Schedule Configuration

-- name: CreateScheduleConfig :one
INSERT INTO schedule_configs (professional_id, day_of_week, start_time, end_time)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListScheduleConfigs :many
SELECT * FROM schedule_configs
WHERE professional_id = $1
ORDER BY day_of_week, start_time;

-- name: DeleteScheduleConfigs :exec
DELETE FROM schedule_configs WHERE professional_id = $1;


-- SECTION: Recurring Rules

-- name: CreateRecurringRule :one
INSERT INTO recurring_rules (professional_id, client_id, day_of_week, start_time, duration_minutes, price, start_date, active)
VALUES ($1, $2, $3, $4, $5, $6, $7, TRUE)
RETURNING *;

-- name: ListRecurringRules :many
SELECT r.*, c.name as client_name
FROM recurring_rules r
JOIN clients c ON r.client_id = c.id
WHERE r.professional_id = $1
ORDER BY r.day_of_week, r.start_time;

-- name: GetActiveRecurringRules :many
SELECT r.* FROM recurring_rules r
JOIN clients c ON r.client_id = c.id
WHERE r.professional_id = $1 
  AND r.active = TRUE 
  AND c.active = TRUE;

-- name: ToggleRecurringRule :one
UPDATE recurring_rules
SET active = $1
WHERE id = $2
RETURNING *;


-- SECTION: Appointments (Calendar)

-- name: CreateAppointment :one
-- UPDATE: Agregamos 'notes' y 'price'
INSERT INTO appointments (
    professional_id, client_id, date, start_time, duration_minutes, price, notes, status, rescheduled_from_id, recurring_rule_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, 'scheduled', $8, $9)
RETURNING *;

-- name: ListAppointmentsInDateRange :many
SELECT a.*, c.name as client_name
FROM appointments a
JOIN clients c ON a.client_id = c.id
WHERE a.professional_id = $1 
  AND a.date >= $2::date 
  AND a.date <= $3::date
ORDER BY a.date, a.start_time;

-- name: GetDayAppointments :many
SELECT id, start_time, duration_minutes, status
FROM appointments
WHERE professional_id = $1 
  AND date = $2::date 
  AND status != 'cancelled';

-- name: UpdateAppointmentStatus :one
UPDATE appointments
SET status = $1, updated_at = NOW()
WHERE id = $2
RETURNING *;

-- name: GetAppointment :one
SELECT * FROM appointments WHERE id = $1 LIMIT 1;

-- name: CheckAppointmentExistsForRule :one
SELECT EXISTS(
    SELECT 1 FROM appointments 
    WHERE recurring_rule_id = $1 
    AND date = $2::date
    AND status != 'cancelled'
);