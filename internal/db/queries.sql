-- SECTION: Psychologists
-- name: CreatePsychologist :one
INSERT INTO psychologists (name, email, phone, cancellation_window_hours)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetPsychologistByEmail :one
SELECT * FROM psychologists 
WHERE email = ? LIMIT 1;

-- name: GetPsychologistSettings :one
SELECT id, cancellation_window_hours 
FROM psychologists 
WHERE id = ?;

-- name: ListPsychologists :many
SELECT id, name, email, phone
FROM psychologists;


-- SECTION: Patients
-- name: CreatePatient :one
INSERT INTO patients (name, psychologist_id, email, phone)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: ListPatients :many
SELECT * FROM patients
WHERE psychologist_id = ? AND active = TRUE
ORDER BY name;

-- name: GetPatient :one
SELECT * FROM patients WHERE id = ? LIMIT 1;


-- SECTION: Schedule Configuration (Availability)
-- name: CreateScheduleConfig :one
INSERT INTO schedule_configs (psychologist_id, day_of_week, start_time, end_time)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: ListScheduleConfigs :many
SELECT * FROM schedule_configs
WHERE psychologist_id = ?
ORDER BY day_of_week, start_time;

-- name: DeleteScheduleConfigs :exec
DELETE FROM schedule_configs WHERE psychologist_id = ?;


-- SECTION: Recurring Rules (Formerly Recurring Slots)
-- name: CreateRecurringRule :one
INSERT INTO recurring_rules (psychologist_id, patient_id, day_of_week, start_time, duration_minutes, active)
VALUES (?, ?, ?, ?, ?, TRUE)
RETURNING *;

-- name: ListRecurringRules :many
SELECT r.*, p.name as patient_name
FROM recurring_rules r
JOIN patients p ON r.patient_id = p.id
WHERE r.psychologist_id = ?
ORDER BY r.day_of_week, r.start_time;

-- name: GetActiveRecurringRules :many
-- Usada por el worker para generar turnos futuros.
-- Solo trae reglas activas de pacientes activos.
SELECT r.* FROM recurring_rules r
JOIN patients p ON r.patient_id = p.id
WHERE r.psychologist_id = ? 
  AND r.active = TRUE 
  AND p.active = TRUE;

-- name: ToggleRecurringRule :one
UPDATE recurring_rules
SET active = ?
WHERE id = ?
RETURNING *;


-- SECTION: Appointments (Calendar)

-- name: CreateAppointment :one
-- Ahora aceptamos recurring_rule_id (puede ser NULL para turnos eventuales)
INSERT INTO appointments (
    psychologist_id, patient_id, date, start_time, duration_minutes, status, rescheduled_from_id, recurring_rule_id
)
VALUES (?, ?, ?, ?, ?, 'scheduled', ?, ?)
RETURNING *;

-- name: ListAppointmentsInDateRange :many
-- Esta query ahora trae TODO (fijos materializados y eventuales)
SELECT a.*, p.name as patient_name
FROM appointments a
JOIN patients p ON a.patient_id = p.id
WHERE a.psychologist_id = ? 
  AND a.date >= ? 
  AND a.date <= ?
ORDER BY a.date, a.start_time;

-- name: GetDayAppointments :many
SELECT id, start_time, duration_minutes, status
FROM appointments
WHERE psychologist_id = ? 
  AND date = ? 
  AND status != 'cancelled';

-- name: UpdateAppointmentStatus :one
UPDATE appointments
SET status = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: GetAppointment :one
SELECT * FROM appointments WHERE id = ? LIMIT 1;

-- name: CheckAppointmentExistsForRule :one
-- Query auxiliar para evitar duplicar turnos al correr el script de generación
SELECT EXISTS(
    SELECT 1 FROM appointments 
    WHERE recurring_rule_id = ? 
    AND date = ?
    AND status != 'cancelled'
);