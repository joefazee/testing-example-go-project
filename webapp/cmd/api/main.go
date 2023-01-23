package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"webapp/pkg/repository"
	"webapp/pkg/repository/dbrepo"
)

const (
	port = 8081
)

type application struct {
	DSN       string
	Port      int
	DB        repository.DatabaseRepo
	Domain    string
	JWTSecret string
}

func main() {
	var app application

	app.Port = port
	flag.StringVar(&app.Domain, "domain", "example.com", "Domain for the application e.g example.com")
	flag.StringVar(&app.DSN, "dsn", "host=localhost port=5432 user=postgres password=postgres dbname=users sslmode=disable timezone=UTC connect_timeout=5", "postres connection string")
	flag.StringVar(&app.JWTSecret, "jwt-secret", "b2xlIjoiQWRtaW4iLCJJc3N1ZXIiOiJJc3N1ZXIiLCJVc2VybmFtZSI6IkphdmFJblVzZSIsImV4cCI6MTY2OTY4MjE1NiwiaWF0IjoxNjY5NjgyMTU2fQ", "JWT Secret")
	flag.Parse()

	conn, err := app.connectToDB()
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	app.DB = &dbrepo.PostgresDBRepo{DB: conn}

	log.Printf("starting api on port %d...", port)

	err = http.ListenAndServe(fmt.Sprintf(":%d", port), app.routes())

	if err != nil {
		log.Fatal(err)
	}
}
