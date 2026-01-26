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


-- SECTION: Schedule Configuration
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


-- SECTION: Recurring Slots (Weekly Contracts)
-- name: CreateRecurringSlot :one
INSERT INTO recurring_slots (psychologist_id, patient_id, day_of_week, start_time, duration_minutes)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: ListRecurringSlots :many
SELECT r.*, p.name as patient_name
FROM recurring_slots r
JOIN patients p ON r.patient_id = p.id
WHERE r.psychologist_id = ?
ORDER BY r.day_of_week, r.start_time;

-- name: GetActiveRecurringSlotsForGeneration :many
SELECT r.* FROM recurring_slots r
JOIN patients p ON r.patient_id = p.id
WHERE r.psychologist_id = ? 
  AND p.active = TRUE;


-- SECTION: Appointments (Calendar)

-- name: CreateAppointment :one
INSERT INTO appointments (
    psychologist_id, patient_id, date, start_time, duration_minutes, status, rescheduled_from_id
)
VALUES (?, ?, ?, ?, ?, 'scheduled', ?)
RETURNING *;

-- name: ListAppointmentsInDateRange :many
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