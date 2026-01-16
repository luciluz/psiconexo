package db

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

// setupTestDB crea una base de datos en memoria NUEVA cada vez que la llamas.
func setupTestDB(t *testing.T) *Queries {

	conn, err := sql.Open("sqlite3", ":memory:")
	assert.NoError(t, err)

	schema, err := os.ReadFile("schema.sql")
	assert.NoError(t, err)

	_, err = conn.Exec(string(schema))
	assert.NoError(t, err)

	return New(conn)
}
