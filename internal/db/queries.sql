-- SECTION: Professionals

-- name: CreateProfessional :one
INSERT INTO professionals (name, email, phone, slug, cancellation_window_hours)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetProfessional :one
SELECT * FROM professionals 
WHERE id = $1 LIMIT 1;

-- name: GetProfessionalByEmail :one
SELECT * FROM professionals 
WHERE email = $1 LIMIT 1;

-- name: GetProfessionalBySlug :one
SELECT id, name, slug, photo_url, title, license_number, bio, email, phone
FROM professionals
WHERE slug = $1 LIMIT 1;

-- name: UpdateProfessionalProfile :one
UPDATE professionals
SET name = $1, phone = $2, slug = $3, photo_url = $4, title = $5, license_number = $6, bio = $7
WHERE id = $8
RETURNING *;


-- SECTION: Professional Settings

-- name: UpsertProfessionalSettings :one
-- "Upsert": Si existe lo actualiza, si no existe lo crea.
INSERT INTO professional_settings (
    professional_id, 
    default_duration_minutes, default_price, buffer_minutes, time_increment_minutes,
    min_booking_notice_hours, max_daily_appointments,
    bank_cbu, bank_alias, bank_name, bank_holder_name, send_alias_by_email,
    mp_access_token, mp_user_id,
    afip_crt_url, afip_key_url, afip_point_of_sale, 
    notify_by_email, notify_by_whatsapp
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
)
ON CONFLICT (professional_id) DO UPDATE SET
    default_duration_minutes = EXCLUDED.default_duration_minutes,
    default_price = EXCLUDED.default_price,
    buffer_minutes = EXCLUDED.buffer_minutes,
    time_increment_minutes = EXCLUDED.time_increment_minutes,
    min_booking_notice_hours = EXCLUDED.min_booking_notice_hours,
    max_daily_appointments = EXCLUDED.max_daily_appointments,
    bank_cbu = EXCLUDED.bank_cbu,
    bank_alias = EXCLUDED.bank_alias,
    bank_name = EXCLUDED.bank_name,
    bank_holder_name = EXCLUDED.bank_holder_name,
    send_alias_by_email = EXCLUDED.send_alias_by_email,
    mp_access_token = EXCLUDED.mp_access_token,
    mp_user_id = EXCLUDED.mp_user_id,
    afip_crt_url = EXCLUDED.afip_crt_url,
    afip_key_url = EXCLUDED.afip_key_url,
    afip_point_of_sale = EXCLUDED.afip_point_of_sale,
    notify_by_email = EXCLUDED.notify_by_email,
    notify_by_whatsapp = EXCLUDED.notify_by_whatsapp,
    updated_at = NOW()
RETURNING *;

-- name: GetProfessionalSettings :one
SELECT * FROM professional_settings
WHERE professional_id = $1;


-- SECTION: Clients

-- name: CreateClient :one
INSERT INTO clients (
    name, email, phone, professional_id, 
    birth_date, medications, emergency_contact_name, emergency_contact_phone, 
    active
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: UpdateClient :one
UPDATE clients
SET name = $1, email = $2, phone = $3, 
    birth_date = $4, medications = $5, 
    emergency_contact_name = $6, emergency_contact_phone = $7
WHERE id = $8
RETURNING *;

-- name: ListClients :many
SELECT * FROM clients
WHERE professional_id = $1 AND active = TRUE
ORDER BY name;

-- name: GetClient :one
SELECT * FROM clients WHERE id = $1 LIMIT 1;


-- SECTION: Clinical Notes (NUEVO - Privacidad)

-- name: CreateClinicalNote :one
INSERT INTO clinical_notes (
    professional_id, client_id, appointment_id, 
    type, content, key_version, status, signed_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: ListClinicalNotes :many
SELECT * FROM clinical_notes
WHERE client_id = $1 AND professional_id = $2
ORDER BY created_at DESC;

-- name: GetNoteById :one
SELECT * FROM clinical_notes WHERE id = $1 LIMIT 1;

-- name: UpdateClinicalNote :one
-- Solo permite editar si NO está firmada (controlar esto en backend también)
UPDATE clinical_notes
SET content = $1, updated_at = NOW()
WHERE id = $2 AND status = 'draft'
RETURNING *;

-- name: SignClinicalNote :one
-- "Firma" la nota: cambia estado a signed y pone fecha
UPDATE clinical_notes
SET status = 'signed', signed_at = NOW(), updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetDraftNotes :many
-- Para el dashboard de "Notas Pendientes"
SELECT n.*, c.name as client_name 
FROM clinical_notes n
JOIN clients c ON n.client_id = c.id
WHERE n.professional_id = $1 AND n.status = 'draft'
ORDER BY n.created_at DESC;


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
INSERT INTO recurring_rules (
    professional_id, client_id, day_of_week, start_time, duration_minutes, 
    modality, price, start_date, active
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, TRUE)
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


-- SECTION: Appointments & Finanzas

-- name: CreateAppointment :one
INSERT INTO appointments (
    professional_id, client_id, date, start_time, duration_minutes, 
    modality, meeting_url, 
    price, concept, notes,
    status, payment_status, payment_method, 
    rescheduled_from_id, recurring_rule_id
)
VALUES (
    $1, $2, $3, $4, $5, 
    $6, $7, 
    $8, $9, $10,
    'scheduled', 'pending', $11, $12, $13
)
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

-- name: GetAppointment :one
SELECT a.*, c.name as client_name, c.email as client_email
FROM appointments a
JOIN clients c ON a.client_id = c.id
WHERE a.id = $1 LIMIT 1;

-- name: UpdateAppointmentStatus :one
UPDATE appointments
SET status = $1, updated_at = NOW()
WHERE id = $2
RETURNING *;

-- name: UpdateAppointmentPayment :one
-- Cuando se recibe el webhook de Mercado Pago o se aprueba transferencia
UPDATE appointments
SET payment_status = $1, payment_method = $2, payment_proof_url = $3, 
    payment_confirmed_at = CASE WHEN $1 = 'paid' THEN NOW() ELSE NULL END,
    updated_at = NOW()
WHERE id = $4
RETURNING *;

-- name: UpdateAppointmentInvoice :one
-- Cuando se genera la factura en AFIP
UPDATE appointments
SET invoice_status = $1, invoice_url = $2, invoice_cae = $3, updated_at = NOW()
WHERE id = $4
RETURNING *;

-- name: GetFinancesDashboard :many
-- Query para la pantalla de "Finanzas" (Tabla principal)
-- Trae todos los turnos que no estén cancelados, ordenados por fecha desc
SELECT a.id, a.date, a.concept, a.price, 
       a.payment_status, a.payment_method, a.payment_proof_url,
       a.invoice_status, a.invoice_url,
       c.name as client_name
FROM appointments a
JOIN clients c ON a.client_id = c.id
WHERE a.professional_id = $1 
  AND a.status != 'cancelled'
ORDER BY a.date DESC
LIMIT $2 OFFSET $3;

-- name: GetFinancialSummary :one
-- KPIs de Finanzas: Sumas rápidas para las tarjetas de arriba
SELECT 
    COALESCE(SUM(CASE WHEN date >= DATE_TRUNC('month', CURRENT_DATE) AND payment_status = 'paid' THEN price ELSE 0 END), 0)::DECIMAL as current_month_income,
    COALESCE(SUM(CASE WHEN payment_status = 'pending' THEN price ELSE 0 END), 0)::DECIMAL as pending_collection,
    COALESCE(SUM(CASE WHEN payment_status = 'paid' AND invoice_status = 'pending' THEN price ELSE 0 END), 0)::DECIMAL as pending_invoicing
FROM appointments
WHERE professional_id = $1 AND status != 'cancelled';

-- name: CheckAppointmentExistsForRule :one
SELECT EXISTS(
    SELECT 1 FROM appointments 
    WHERE recurring_rule_id = $1 
    AND date = $2::date
    AND status != 'cancelled'
);

-- name: CheckSlugAvailability :one
-- Verifica si un slug está disponible (para Configuración - Perfil)
SELECT NOT EXISTS(
    SELECT 1 FROM professionals 
    WHERE slug = $1 AND id != $2
) as available;

-- name: UpdateAppointmentNotes :one
-- Actualiza el comentario simple del turno (ej: "paciente llegó tarde")
UPDATE appointments
SET notes = $1, updated_at = NOW()
WHERE id = $2
RETURNING *;

-- name: UpdateRecurringRule :one
-- Edita una regla de recurrencia existente
UPDATE recurring_rules
SET 
    day_of_week = $1,
    start_time = $2,
    duration_minutes = $3,
    modality = $4,
    price = $5,
    active = $6
WHERE id = $7
RETURNING *;