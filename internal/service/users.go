package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/luciluz/psiconexo/internal/db"
)

type CreateProfessionalRequest struct {
	Name                    string
	Email                   string
	Phone                   string
	CancellationWindowHours int
}

type CreateClientRequest struct {
	Name           string
	Email          string
	Phone          string
	ProfessionalID int64
}

func (s *Service) CreateProfessional(ctx context.Context, req CreateProfessionalRequest) (*db.Professional, error) {

	prof, err := s.queries.CreateProfessional(ctx, db.CreateProfessionalParams{
		Name:                    req.Name,
		Email:                   req.Email,
		Phone:                   sql.NullString{String: req.Phone, Valid: req.Phone != ""},
		CancellationWindowHours: sql.NullInt32{Int32: int32(req.CancellationWindowHours), Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("error creando profesional: %w", err)
	}
	return &prof, nil
}

func (s *Service) CreateClient(ctx context.Context, req CreateClientRequest) (*db.Client, error) {

	client, err := s.queries.CreateClient(ctx, db.CreateClientParams{
		Name:           req.Name,
		Email:          sql.NullString{String: req.Email, Valid: req.Email != ""}, // Ahora es nullable
		Phone:          sql.NullString{String: req.Phone, Valid: req.Phone != ""}, // Ahora es nullable
		ProfessionalID: req.ProfessionalID,
		Active:         sql.NullBool{Bool: true, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("error creando cliente: %w", err)
	}
	return &client, nil
}

func (s *Service) ListProfessionals(ctx context.Context) ([]db.ListProfessionalsRow, error) {
	profs, err := s.queries.ListProfessionals(ctx)
	if err != nil {
		return nil, fmt.Errorf("error listando profesionales: %w", err)
	}
	return profs, nil
}

func (s *Service) ListClients(ctx context.Context, professionalID int64) ([]db.Client, error) {
	clients, err := s.queries.ListClients(ctx, professionalID)
	if err != nil {
		return nil, fmt.Errorf("error listando clientes para profesional %d: %w", professionalID, err)
	}
	return clients, nil
}
