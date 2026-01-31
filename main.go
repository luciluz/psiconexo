package main

import (
	"database/sql"
	"log"

	"github.com/luciluz/psiconexo/internal/api"
	"github.com/luciluz/psiconexo/internal/db"
	"github.com/luciluz/psiconexo/internal/service"

	_ "github.com/mattn/go-sqlite3"
)

func main() {

	conn, err := sql.Open("sqlite3", "./psiconexo.db")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	if _, err := conn.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		log.Fatal(err)
	}

	queries := db.New(conn)
	svc := service.NewService(queries, conn)
	handler := api.NewHandler(svc)

	r := api.NewRouter(handler)

	log.Println("Servidor corriendo en http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
