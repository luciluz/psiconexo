package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreatePatient(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		inputName      string
		inputEmail     string
		inputPhone     string
		psychologistID int64
		expectError    bool
		setup          func(q *Queries) int64
	}{
		{
			name:        "Success - Creación vinculada a psicólogo existente",
			inputName:   "Jesse Pinkman",
			inputEmail:  "jesse@cook.com",
			inputPhone:  "11-2222-3333",
			expectError: false,
			setup: func(q *Queries) int64 {
				psy, _ := q.CreatePsychologist(ctx, CreatePsychologistParams{
					Name:                    "Walter White",
					Email:                   "walter@meth.com",
					Phone:                   sql.NullString{String: "999", Valid: true},
					CancellationWindowHours: sql.NullInt64{Int64: 24, Valid: true},
				})
				return psy.ID
			},
		},
		{
			name:        "Fail - Foreign Key (Psicólogo inexistente)",
			inputName:   "Paciente Huerfano",
			inputEmail:  "orphan@test.com",
			inputPhone:  "000",
			expectError: true,
			setup: func(q *Queries) int64 {
				return 99999
			},
		},
		{
			name:        "Fail - Email duplicado",
			inputName:   "Clon de Jesse",
			inputEmail:  "jesse@cook.com",
			inputPhone:  "555-666",
			expectError: true,
			setup: func(q *Queries) int64 {
				psy, _ := q.CreatePsychologist(ctx, CreatePsychologistParams{
					Name:                    "Saul Goodman",
					Email:                   "saul@law.com",
					Phone:                   sql.NullString{String: "505", Valid: true},
					CancellationWindowHours: sql.NullInt64{Int64: 24, Valid: true},
				})
				_, _ = q.CreatePatient(ctx, CreatePatientParams{
					Name:           "Jesse Original",
					Email:          "jesse@cook.com",
					PsychologistID: psy.ID,
				})
				return psy.ID
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queries := setupTestDB(t)
			targetPsyID := tt.setup(queries)

			patient, err := queries.CreatePatient(ctx, CreatePatientParams{
				Name:           tt.inputName,
				Email:          tt.inputEmail,
				Phone:          sql.NullString{String: tt.inputPhone, Valid: true},
				PsychologistID: targetPsyID,
			})

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.inputName, patient.Name)
				assert.Equal(t, tt.inputEmail, patient.Email)
				assert.Equal(t, targetPsyID, patient.PsychologistID)
				assert.True(t, patient.Active.Bool)
			}
		})
	}
}

func TestListPatients(t *testing.T) {
	ctx := context.Background()
	queries := setupTestDB(t)

	psy, _ := queries.CreatePsychologist(ctx, CreatePsychologistParams{
		Name:  "Dr. Listas",
		Email: "listas@test.com",
	})

	names := []string{"Ana", "Beto", "Carla"}
	for _, name := range names {
		_, err := queries.CreatePatient(ctx, CreatePatientParams{
			Name:           name,
			Email:          name + "@test.com",
			PsychologistID: psy.ID,
		})
		assert.NoError(t, err)
	}

	// listar
	patients, err := queries.ListPatients(ctx, psy.ID)
	assert.NoError(t, err)

	// validar orden (ORDER BY name) y cantidad
	assert.Len(t, patients, 3)
	assert.Equal(t, "Ana", patients[0].Name)
	assert.Equal(t, "Beto", patients[1].Name)
	assert.Equal(t, "Carla", patients[2].Name)
}
