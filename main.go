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
	// 1. Inicializar DB
	// Nota: Si no existe el archivo, SQLite lo crea, pero necesitamos correr el schema manualmente
	// la primera vez o usar migraciones. Para dev rápido, asumo que ya tienes psiconexo.db
	conn, err := sql.Open("sqlite3", "./psiconexo.db")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// IMPORTANTE: Activar Foreign Keys en SQLite siempre
	if _, err := conn.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		log.Fatal(err)
	}

	// 2. Inicializar Capas
	queries := db.New(conn)                  // Capa de Datos
	svc := service.NewService(queries, conn) // Capa de Servicio
	handler := api.NewHandler(svc)           // Capa de API

	// 3. Configurar Router
	r := api.NewRouter(handler)

	// 4. Arrancar Servidor
	log.Println("🚀 Servidor corriendo en http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
