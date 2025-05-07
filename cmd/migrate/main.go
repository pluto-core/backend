package main

import (
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: migrate <auth|manifest>")
	}
	svc := os.Args[1]
	path := "migrations/" + svc
	m, err := migrate.New("file://"+path, os.Getenv("PLUTO_DATABASE_DSN"))
	if err != nil {
		log.Fatal(err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}
	log.Printf("migrations for %s applied", svc)
}
