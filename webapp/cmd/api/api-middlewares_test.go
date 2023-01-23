package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"webapp/pkg/data"
)

func Test_api_app_enableCORS(t *testing.T) {

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	testCases := []struct {
		name         string
		method       string
		expectHeader bool
	}{
		{"preflight", http.MethodOptions, true},
		{"get", http.MethodGet, false},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			handlerToTest := app.enableCORS(next)

			req := httptest.NewRequest(tt.method, "http://test", nil)
			rr := httptest.NewRecorder()

			handlerToTest.ServeHTTP(rr, req)

			if tt.expectHeader && rr.Header().Get("Access-Control-Allow-Credentials") == "" {
				t.Error("expect to get header: Access-Control-Allow-Credentials")
			}

			if !tt.expectHeader && rr.Header().Get("Access-Control-Allow-Credentials") != "" {
				t.Error("expect no headers, but gone one: Access-Control-Allow-Credentials")
			}
		})
	}
}

func Test_api_app_authRequired(t *testing.T) {

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	testUser := data.User{
		ID:        1,
		FirstName: "Admin",
		LastName:  "User",
		Email:     "admin@example.com",
	}

	tokes, _ := app.generateTokenPair(&testUser)

	testCases := []struct {
		name             string
		token            string
		expectAuthorized bool
		setHeader        bool
	}{
		{name: "valid token", token: fmt.Sprintf("Bearer %s", tokes.Token), expectAuthorized: true, setHeader: true},
		{name: "no token", token: "", expectAuthorized: false, setHeader: false},
		{name: "invalid token", token: fmt.Sprintf("Bearer %s", expiredToken), expectAuthorized: false, setHeader: true},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/", nil)
			if tt.setHeader {
				req.Header.Set("Authorization", tt.token)
			}

			rr := httptest.NewRecorder()
			handlerToTest := app.authRequired(next)
			handlerToTest.ServeHTTP(rr, req)

			if tt.expectAuthorized && rr.Code == http.StatusUnauthorized {
				t.Error("expect user to be expectAuthorized; got got 401")
			}

			if !tt.expectAuthorized && rr.Code != http.StatusUnauthorized {
				t.Errorf("expect status code to be 401; got %d", rr.Code)
			}
		})
	}
}
