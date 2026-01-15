-- name: ListPatients :many
SELECT * FROM patients
WHERE psychologist_id = ?;