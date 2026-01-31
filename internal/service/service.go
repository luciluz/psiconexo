package service

import (
	"database/sql"

	"github.com/luciluz/psiconexo/internal/db"
)

type Service struct {
	queries *db.Queries
	db      *sql.DB
}

func NewService(queries *db.Queries, dbConn *sql.DB) *Service {
	return &Service{
		queries: queries,
		db:      dbConn,
	}
}
