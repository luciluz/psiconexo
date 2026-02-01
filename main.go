package main

import (
	"database/sql"
	"log"
	"os"

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

	log.Println("Verificando estructura de base de datos...")
	schema, err := os.ReadFile("internal/db/schema.sql")
	if err != nil {
		log.Fatal("Error leyendo el archivo schema.sql: ", err)
	}

	if _, err := conn.Exec(string(schema)); err != nil {
		log.Fatal("Error ejecutando schema.sql: ", err)
	}
	log.Println("Base de datos lista.")

	queries := db.New(conn)
	svc := service.NewService(queries, conn)
	handler := api.NewHandler(svc)

	r := api.NewRouter(handler)

	log.Println("Servidor corriendo en http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
