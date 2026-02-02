package service

import (
	"context"

	"github.com/luciluz/psiconexo/internal/db"
)

type ScheduleBlock struct {
	DayOfWeek int
	StartTime string
	EndTime   string
}

type UpdateScheduleRequest struct {
	ProfessionalID int64
	Blocks         []ScheduleBlock
}

func (s *Service) UpdateSchedule(ctx context.Context, req UpdateScheduleRequest) error {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	qtx := s.queries.WithTx(tx)

	// 1. Borramos configuración anterior
	if err := qtx.DeleteScheduleConfigs(ctx, req.ProfessionalID); err != nil {
		return err
	}

	// 2. Insertamos bloques nuevos
	for _, block := range req.Blocks {
		_, err := qtx.CreateScheduleConfig(ctx, db.CreateScheduleConfigParams{
			ProfessionalID: req.ProfessionalID,
			DayOfWeek:      int32(block.DayOfWeek),
			StartTime:      block.StartTime,
			EndTime:        block.EndTime,
		})
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Service) ListSchedule(ctx context.Context, professionalID int64) ([]db.ScheduleConfig, error) {
	return s.queries.ListScheduleConfigs(ctx, professionalID)
}
