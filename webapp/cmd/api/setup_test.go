package main

import (
	"os"
	"testing"
	"webapp/pkg/repository/dbrepo"
)

var app application
var expiredToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbiI6dHJ1ZSwiYXVkIjoiZXhhbXBsZS5jb20iLCJleHAiOjE2NjkzOTg3NTgsImlzcyI6ImV4YW1wbGUuY29tIiwibmFtZSI6IkpvaG4gRG9lIiwic3ViIjoiMSJ9.YTddz_3beQtllgaD-HvMEy1ZhZnbeK-T1x4G1SNhO1E"

func TestMain(m *testing.M) {

	app.DB = &dbrepo.MockDBRepo{}
	app.Domain = "example.com"
	app.JWTSecret = "b2xlIjoiQWRtaW4iLCJJc3N1ZXIiOiJJc3N1ZXIiLCJVc2VybmFtZSI6IkphdmFJblVzZSIsImV4cCI6MTY2OTY4MjE1NiwiaWF0IjoxNjY5NjgyMTU2fQ"

	os.Exit(m.Run())

}
