package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"webapp/pkg/data"
)

func Test_api_app_getTokenFromHeaderAndVerify(t *testing.T) {

	testUser := data.User{
		ID:        1,
		FirstName: "Admin",
		LastName:  "User",
		IsAdmin:   1,
		Email:     "admin@example.com",
	}

	tokens, err := app.generateTokenPair(&testUser)
	if err != nil {
		t.Fatal(err)
	}

	var tests = []struct {
		name          string
		token         string
		errorExpected bool
		setHeader     bool
		issuer        string
	}{
		{"valid", fmt.Sprintf("Bearer %s", tokens.Token), false, true, app.Domain},
		{"valid expired", fmt.Sprintf("Bearer %s", expiredToken), true, true, app.Domain},
		{"no headers", "", true, false, app.Domain},
		{"invalid token", "Bearer invalid", true, true, app.Domain},
		{"no Bearer", fmt.Sprintf("Bea %s", tokens.Token), true, true, app.Domain},
		{"three header parts", fmt.Sprintf("Bearer %s s", tokens.Token), true, true, app.Domain},
		// make sure this test is the last one
		{"wrong issuer", fmt.Sprintf("Bearer %s", tokens.Token), true, true, "dummy@domain.com"},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			if tt.issuer != app.Domain {
				app.Domain = tt.issuer
				tokens, _ = app.generateTokenPair(&testUser)
			}
			req, _ := http.NewRequest(http.MethodGet, "/", nil)
			if tt.setHeader {
				req.Header.Set("Authorization", tt.token)
			}

			rr := httptest.NewRecorder()
			_, _, err = app.getTokenFromHeaderAndVerify(rr, req)

			if err != nil && !tt.errorExpected {
				t.Errorf("did not expect error but got one; %s", err)
			}

			if err == nil && tt.errorExpected {
				t.Error("expected error, but did not get one")
			}

			app.Domain = "example.com"
		})
	}
}
