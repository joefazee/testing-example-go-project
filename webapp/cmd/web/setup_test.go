package main

import (
	"os"
	"testing"
	"webapp/pkg/repository/dbrepo"
)

var app application

func TestMain(m *testing.M) {
	templatePath = "./../../template/"

	app.Session = getSession()

	app.DB = &dbrepo.MockDBRepo{}
	app.Session = getSession()

	os.Exit(m.Run())

}
