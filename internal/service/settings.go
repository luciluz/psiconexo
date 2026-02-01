package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/luciluz/psiconexo/internal/db"
)

type UpdateSettingsRequest struct {
	ProfessionalID         int64
	DefaultDurationMinutes int
	BufferMinutes          int
	TimeIncrementMinutes   int
	MinBookingNoticeHours  int
	MaxDailyAppointments   *int
}

func (s *Service) UpdateSettings(ctx context.Context, req UpdateSettingsRequest) (*db.ProfessionalSetting, error) {

	// Conversión especial para el campo que puede ser NULL
	var maxDaily sql.NullInt32
	if req.MaxDailyAppointments != nil {
		maxDaily = sql.NullInt32{Int32: int32(*req.MaxDailyAppointments), Valid: true}
	} else {
		maxDaily = sql.NullInt32{Valid: false}
	}

	// Usamos Upsert: Crea o Actualiza
	settings, err := s.queries.UpsertProfessionalSettings(ctx, db.UpsertProfessionalSettingsParams{
		ProfessionalID:         req.ProfessionalID,
		DefaultDurationMinutes: sql.NullInt32{Int32: int32(req.DefaultDurationMinutes), Valid: true},
		BufferMinutes:          sql.NullInt32{Int32: int32(req.BufferMinutes), Valid: true},
		TimeIncrementMinutes:   sql.NullInt32{Int32: int32(req.TimeIncrementMinutes), Valid: true},
		MinBookingNoticeHours:  sql.NullInt32{Int32: int32(req.MinBookingNoticeHours), Valid: true},
		MaxDailyAppointments:   maxDaily,
	})

	if err != nil {
		return nil, fmt.Errorf("error actualizando configuración: %w", err)
	}

	return &settings, nil
}

func (s *Service) GetSettings(ctx context.Context, profID int64) (*db.ProfessionalSetting, error) {
	settings, err := s.queries.GetProfessionalSettings(ctx, profID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Si no tiene configuración, devolvemos nil (el handler decidirá si manda 404 o defaults)
			return nil, nil
		}
		return nil, fmt.Errorf("error obteniendo configuración: %w", err)
	}
	return &settings, nil
}
