package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/luciluz/psiconexo/internal/db"
)

type CreateClinicalNoteRequest struct {
	ProfessionalID int64
	ClientID       int64
	AppointmentID  *int64 // Puntero: puede ser nil (nota general, no vinculada a turno)
	Content        string // Texto (idealmente encriptado desde el front)
}

type UpdateClinicalNoteRequest struct {
	NoteID  int64
	Content string
}

func (s *Service) CreateClinicalNote(ctx context.Context, req CreateClinicalNoteRequest) (*db.ClinicalNote, error) {

	// Manejo de AppointmentID opcional (Nullable en DB)
	var appID sql.NullInt64
	if req.AppointmentID != nil {
		appID = sql.NullInt64{Int64: *req.AppointmentID, Valid: true}
	} else {
		appID = sql.NullInt64{Valid: false}
	}

	// Asumimos is_encrypted = true por defecto, ya que la idea es que el front mande basura encriptada.
	// Si mandas texto plano, igual se guarda, pero el flag queda en true indicando la intención del sistema.
	note, err := s.queries.CreateClinicalNote(ctx, db.CreateClinicalNoteParams{
		ProfessionalID: req.ProfessionalID,
		ClientID:       req.ClientID,
		AppointmentID:  appID,
		Content:        req.Content,
		IsEncrypted:    sql.NullBool{Bool: true, Valid: true},
	})

	if err != nil {
		return nil, fmt.Errorf("error creando nota clínica: %w", err)
	}

	return &note, nil
}

func (s *Service) ListClinicalNotes(ctx context.Context, clientID, professionalID int64) ([]db.ClinicalNote, error) {
	notes, err := s.queries.ListClinicalNotes(ctx, db.ListClinicalNotesParams{
		ClientID:       clientID,
		ProfessionalID: professionalID,
	})
	if err != nil {
		return nil, fmt.Errorf("error listando historial clínico: %w", err)
	}
	return notes, nil
}

func (s *Service) UpdateClinicalNote(ctx context.Context, req UpdateClinicalNoteRequest) (*db.ClinicalNote, error) {
	note, err := s.queries.UpdateClinicalNote(ctx, db.UpdateClinicalNoteParams{
		Content: req.Content,
		ID:      req.NoteID,
	})
	if err != nil {
		return nil, fmt.Errorf("error actualizando nota clínica %d: %w", req.NoteID, err)
	}
	return &note, nil
}
