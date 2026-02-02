package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/luciluz/psiconexo/internal/api"
	"github.com/luciluz/psiconexo/internal/db"
	"github.com/luciluz/psiconexo/internal/service"

	_ "github.com/lib/pq" // <--- CAMBIO CLAVE: Driver de Postgres
)

func main() {
	// 1. Leemos la URL de conexión desde las variables de entorno (definidas en Docker)
	// Si no existe (ej: ejecutando local sin docker), usamos un default para desarrollo.
	dbSource := os.Getenv("DB_SOURCE")
	if dbSource == "" {
		dbSource = "postgres://psico_user:psico_password@localhost:5432/psiconexo?sslmode=disable"
	}

	// 2. Abrimos la conexión con el driver "postgres"
	conn, err := sql.Open("postgres", dbSource)
	if err != nil {
		log.Fatal("Error abriendo conexión a DB:", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	// 3. Verificamos que la base de datos responda (Ping)
	if err := conn.Ping(); err != nil {
		log.Fatal("No se pudo conectar a la base de datos Postgres:", err)
	}

	log.Println("Verificando estructura de base de datos...")

	// Leemos el archivo SQL
	schema, err := os.ReadFile("internal/db/schema.sql")
	if err != nil {
		log.Fatal("Error leyendo el archivo schema.sql: ", err)
	}

	// Ejecutamos el schema
	if _, err := conn.Exec(string(schema)); err != nil {
		log.Fatal("Error ejecutando schema.sql: ", err)
	}
	log.Println("Base de datos lista.")

	// 4. Inicialización de Capas
	queries := db.New(conn)
	svc := service.NewService(queries, conn)
	handler := api.NewHandler(svc)

	r := api.NewRouter(handler)

	log.Println("Servidor corriendo en http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
