package main

import (
	"encoding/gob"
	"flag"
	"log"
	"net/http"
	"webapp/pkg/data"
	"webapp/pkg/repository"
	"webapp/pkg/repository/dbrepo"

	"github.com/alexedwards/scs/v2"
)

type application struct {
	Session *scs.SessionManager
	DB      repository.DatabaseRepo
	DSN     string
}

func main() {

	// register type
	gob.Register(data.User{})
	// setup an app config
	app := application{}

	flag.StringVar(&app.DSN, "dsn", "host=localhost port=5432 user=postgres password=postgres dbname=users sslmode=disable timezone=UTC connect_timeout=5", "postres connection string")

	flag.Parse()

	conn, err := app.connectToDB()
	if err != nil {
		log.Fatal(err)
	}
	app.DB = &dbrepo.PostgresDBRepo{DB: conn}
	app.Session = getSession()

	defer conn.Close()

	// routes
	mux := app.routes()

	// print out message

	log.Println("starting server on port 8080")

	// start server

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
