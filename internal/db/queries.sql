-- name: CreatePsychologist :one
INSERT INTO psychologists (name, email, phone)
VALUES (?, ?, ?)
RETURNING *;

-- name: CreatePatient :one
INSERT INTO patients (name, email, phone, psychologist_id)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: ListPatients :many
SELECT * FROM patients
WHERE psychologist_id = ?;