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
	PsychologistID int64
	Blocks         []ScheduleBlock
}

func (s *Service) UpdateSchedule(ctx context.Context, req UpdateScheduleRequest) error {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := s.queries.WithTx(tx)

	if err := qtx.DeleteScheduleConfigs(ctx, req.PsychologistID); err != nil {
		return err
	}

	for _, block := range req.Blocks {
		_, err := qtx.CreateScheduleConfig(ctx, db.CreateScheduleConfigParams{
			PsychologistID: req.PsychologistID,
			DayOfWeek:      int64(block.DayOfWeek),
			StartTime:      block.StartTime,
			EndTime:        block.EndTime,
		})
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Service) ListSchedule(ctx context.Context, psychologistID int64) ([]db.ScheduleConfig, error) {
	return s.queries.ListScheduleConfigs(ctx, psychologistID)
}
