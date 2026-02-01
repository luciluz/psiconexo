-- SECTION: Professionals (Antes Psychologists)

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

-- name: GetProfessionalSettings :one
SELECT id, cancellation_window_hours 
FROM professionals 
WHERE id = $1;

-- name: ListProfessionals :many
SELECT id, name, email, phone
FROM professionals;


-- SECTION: Clients (Antes Patients)

-- name: CreateClient :one
-- Nota: active se pasa explícitamente o se deja default en true
INSERT INTO clients (name, email, phone, professional_id, active)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListClients :many
SELECT * FROM clients
WHERE professional_id = $1 AND active = TRUE
ORDER BY name;

-- name: GetClient :one
SELECT * FROM clients WHERE id = $1 LIMIT 1;


-- SECTION: Schedule Configuration (Availability)

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
-- Agregamos 'price' y 'start_date'. Active va fijo en TRUE al crear.
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
-- Usada por el worker. Solo trae reglas activas de clientes activos.
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
-- Agregamos 'price' (que viene de la regla o del input manual)
INSERT INTO appointments (
    professional_id, client_id, date, start_time, duration_minutes, price, status, rescheduled_from_id, recurring_rule_id
)
VALUES ($1, $2, $3, $4, $5, $6, 'scheduled', $7, $8)
RETURNING *;

-- name: ListAppointmentsInDateRange :many
-- En Postgres usamos $2 y $3 para las fechas.
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
-- Usamos NOW() para Postgres
UPDATE appointments
SET status = $1, updated_at = NOW()
WHERE id = $2
RETURNING *;

-- name: GetAppointment :one
SELECT * FROM appointments WHERE id = $1 LIMIT 1;

-- name: CheckAppointmentExistsForRule :one
-- Query auxiliar
SELECT EXISTS(
    SELECT 1 FROM appointments 
    WHERE recurring_rule_id = $1 
    AND date = $2::date
    AND status != 'cancelled'
);