package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/luciluz/psiconexo/internal/db"
)

// DTO (Data Transfer Object)
type CreatePsychologistRequest struct {
	Name                    string
	Email                   string
	Phone                   string
	CancellationWindowHours int
}

func (s *Service) CreatePsychologist(ctx context.Context, req CreatePsychologistRequest) (*db.Psychologist, error) {
	// Aquí podrías validar formato de email, etc.

	psy, err := s.queries.CreatePsychologist(ctx, db.CreatePsychologistParams{
		Name:                    req.Name,
		Email:                   req.Email,
		Phone:                   sql.NullString{String: req.Phone, Valid: req.Phone != ""},
		CancellationWindowHours: sql.NullInt64{Int64: int64(req.CancellationWindowHours), Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("error creando psicólogo: %w", err)
	}
	return &psy, nil
}

// --- Pacientes ---

type CreatePatientRequest struct {
	Name           string
	Email          string
	Phone          string
	PsychologistID int64
}

func (s *Service) CreatePatient(ctx context.Context, req CreatePatientRequest) (*db.Patient, error) {
	// Validación básica: Verificar si el psicólogo existe antes de intentar insertar
	// (Aunque la FK lo atrapa, a veces es mejor checkearlo antes para dar un error más lindo)

	pat, err := s.queries.CreatePatient(ctx, db.CreatePatientParams{
		Name:           req.Name,
		Email:          req.Email,
		Phone:          sql.NullString{String: req.Phone, Valid: req.Phone != ""},
		PsychologistID: req.PsychologistID,
	})
	if err != nil {
		return nil, fmt.Errorf("error creando paciente: %w", err)
	}
	return &pat, nil
}

func (s *Service) ListPsychologists(ctx context.Context) ([]db.ListPsychologistsRow, error) {
	psy, err := s.queries.ListPsychologists(ctx)
	if err != nil {
		return nil, fmt.Errorf("error listando pacientes: %w", err)
	}
	return psy, nil
}

func (s *Service) ListPatients(ctx context.Context, psychologistID int64) ([]db.Patient, error) {
	patients, err := s.queries.ListPatients(ctx, psychologistID)
	if err != nil {
		return nil, fmt.Errorf("error listando pacientes para psicólogo %d: %w", psychologistID, err)
	}
	return patients, nil
}
