package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreatPsychologist(t *testing.T) {
	ctx := context.Background()

	// definimos casos (success, fail, edge cases)
	tests := []struct {
		name        string
		inputName   string
		inputEmail  string
		inputPhone  string
		expectError bool
		setup       func(q *Queries)
	}{
		{
			name:        "Success - Creación estándar",
			inputName:   "Sigmund Freud",
			inputEmail:  "tu.sigmuncito@psicomails.com",
			inputPhone:  "666",
			expectError: false,
		},
		{
			name:        "Fail - Email duplicado",
			inputName:   "Sigmund Perez",
			inputEmail:  "tu.sigmuncito@psicomails.com",
			inputPhone:  "123",
			expectError: true,
			setup: func(q *Queries) {
				_, _ = q.CreatePsychologist(ctx, CreatePsychologistParams{
					Name:  "Sigmund Freud",
					Email: "tu.sigmuncito@psicomails.com",
					Phone: sql.NullString{String: "123-456", Valid: true},
				})
			},
		},
		{
			name:        "Fail - Teléfono duplicado",
			inputName:   "Doctor Octopus",
			inputEmail:  "octopus@doc.com",
			inputPhone:  "555-1111",
			expectError: true,
			setup: func(q *Queries) {
				_, _ = q.CreatePsychologist(ctx, CreatePsychologistParams{
					Name:  "Otto Octavius",
					Email: "octavius@uba.com",
					Phone: sql.NullString{String: "555-1111", Valid: true},
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// creamos un DB nueva para testear
			queries := setupTestDB(t)

			if tt.setup != nil {
				tt.setup(queries)
			}

			// intentamos crear el psicólogo del test
			psy, err := queries.CreatePsychologist(ctx, CreatePsychologistParams{
				Name:  tt.inputName,
				Email: tt.inputEmail,
				Phone: sql.NullString{String: tt.inputPhone, Valid: true},
			})

			if tt.expectError {
				assert.Error(t, err, "Se esperaba un error y no ocurrió")
			} else {
				assert.NoError(t, err, "Ocurrió un error inesperado")
				assert.Equal(t, tt.inputName, psy.Name)
				assert.Equal(t, tt.inputEmail, psy.Email)
				assert.NotZero(t, psy.ID)
			}
		})
	}
}
