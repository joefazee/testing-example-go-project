package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
	"webapp/pkg/data"

	"github.com/go-chi/chi/v5"
)

func Test__api_app_authentication(t *testing.T) {

	testCases := []struct {
		name               string
		requestBody        string
		expectedStatusCode int
	}{
		{"valid user", `{"email": "admin@example.com", "password": "secret"}`, http.StatusOK},
		{"not json", `sample`, http.StatusBadRequest},
		{"empty json", `{}`, http.StatusUnauthorized},
		{"empty email", `{"email": "", "password": "secret"}`, http.StatusUnauthorized},
		{"empty password", `{"email": "admin@example.com", "password": ""}`, http.StatusUnauthorized},
		{"invalid-password", `{"email": "admin@example.com", "password": "secret1"}`, http.StatusUnauthorized},
		{"invalid-email", `{"email": "adminsss@example.com", "password": "secret"}`, http.StatusUnauthorized},
		{"invalid-email-and-password", `{"email": "adminsss@example.com", "password": "secret1"}`, http.StatusUnauthorized},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/auth", reader)
			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(app.authenticate)
			handler.ServeHTTP(rr, req)

			if tt.expectedStatusCode != rr.Code {
				t.Errorf("expect HTTP status code of %d; got %d", rr.Code, tt.expectedStatusCode)
			}
		})
	}

}

func Test_api_app_refresh(t *testing.T) {
	testCases := []struct {
		name               string
		token              string
		expectedStatusCode int
		resetRereshTime    bool
	}{
		{"valid", "", http.StatusOK, true},
		{"valid but too early to refresh", "", http.StatusTooEarly, false},
		{"expired token", expiredToken, http.StatusBadRequest, false},
	}

	testUser := data.User{
		ID:        1,
		FirstName: "Admin",
		LastName:  "User",
		Email:     "admin@example.com",
	}

	oldRefreshTime := refreshTokenExpiry

	for _, tt := range testCases {

		t.Run(tt.name, func(t *testing.T) {

			var tkn string
			if tt.resetRereshTime {
				refreshTokenExpiry = time.Second * 5
			}
			if tt.token == "" {
				tokens, _ := app.generateTokenPair(&testUser)
				tkn = tokens.RefreshToken
			} else {
				tkn = tt.token
			}

			postedData := url.Values{
				"refresh_token": {tkn},
			}

			req, _ := http.NewRequest(http.MethodPost, "/refresh-token", strings.NewReader(postedData.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(app.refresh)
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatusCode {
				t.Errorf("expect status code %d; got %d", tt.expectedStatusCode, rr.Code)
			}
		})

		refreshTokenExpiry = oldRefreshTime

	}

}

func Test_api_app_userHandlers(t *testing.T) {

	testCases := []struct {
		name           string
		method         string
		json           string
		paramID        string
		handler        http.HandlerFunc
		expectedStatus int
	}{
		{"allUsers", http.MethodGet, "", "", app.allUsers, http.StatusOK},
		{"deleteUser", http.MethodDelete, "", "1", app.deleteUser, http.StatusNoContent},
		{"deleteUser bad param", http.MethodDelete, "", "s", app.deleteUser, http.StatusBadRequest},
		{"deleteUser invalid user id", http.MethodDelete, "", "2", app.deleteUser, http.StatusBadRequest},
		{"getUser valid", http.MethodGet, "", "1", app.getUser, http.StatusOK},
		{"getUser invalid id", http.MethodGet, "", "2", app.getUser, http.StatusBadRequest},
		{"getUser invalid param", http.MethodGet, "", "s", app.getUser, http.StatusBadRequest},
		{
			"UpdateUser valid",
			http.MethodPatch,
			`{"id": 1, "first_name": "Admin New",  "last_name": "User new", "email": "admin@example.com"}`,
			"",
			app.updateUser,
			http.StatusNoContent,
		},

		{
			"UpdateUser invalid json",
			http.MethodPatch,
			`{"id": 1, "first_name": "Admin New",  "last_name": "User new", "email": admin@example.com"}`,
			"",
			app.updateUser,
			http.StatusBadRequest,
		},

		{
			"UpdateUser invalid user",
			http.MethodPatch,
			`{"id": 9999, "first_name": "Admin New",  "last_name": "User new", "email": "admin@example.com"}`,
			"",
			app.updateUser,
			http.StatusBadRequest,
		},

		{
			"insertUser valid",
			http.MethodPut,
			`{"first_name": "Jack",  "last_name": "Neo", "email": "neo@example.com"}`,
			"",
			app.insertUser,
			http.StatusNoContent,
		},

		{
			"insertUser invalid valid",
			http.MethodPut,
			`{"first_name": "Jack", "foo": "bar",  "last_name": "Neo", "email": "neo@example.com"}`,
			"",
			app.insertUser,
			http.StatusBadRequest,
		},

		{
			"insertUser invalid json",
			http.MethodPut,
			`{"first_name": "Jack", "last_name": "Neo", "email": neo@example.com"}`,
			"",
			app.insertUser,
			http.StatusBadRequest,
		},

		{
			"insertUser invalid db operation",
			http.MethodPut,
			`{"first_name": "Jack",  "last_name": "Neo", "email": "invalid@example.com"}`,
			"",
			app.insertUser,
			http.StatusBadRequest,
		},
	}

	for _, tt := range testCases {
		var req *http.Request
		if tt.json == "" {
			req, _ = http.NewRequest(tt.method, "/", nil)
		} else {
			req, _ = http.NewRequest(tt.method, "/", strings.NewReader(tt.json))
		}

		if tt.paramID != "" {
			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("userID", tt.paramID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		}

		rr := httptest.NewRecorder()

		handlerToTest := http.HandlerFunc(tt.handler)
		handlerToTest.ServeHTTP(rr, req)

		if rr.Code != tt.expectedStatus {
			t.Errorf("expected status code %d; got %d", tt.expectedStatus, rr.Code)
		}
	}
}

func Test_api_app_refreshUsingCookie(t *testing.T) {

	testUser := data.User{
		ID:        1,
		FirstName: "Admin",
		LastName:  "User",
		Email:     "admin@example.com",
	}

	tokens, _ := app.generateTokenPair(&testUser)

	testCookie := &http.Cookie{
		Name:     "__Host-refresh_token",
		Path:     "/",
		Value:    tokens.RefreshToken,
		Expires:  time.Now().Add(refreshTokenExpiry),
		MaxAge:   int(refreshTokenExpiry.Seconds()),
		SameSite: http.SameSiteStrictMode,
		Domain:   "localhost",
		HttpOnly: true,
		Secure:   true,
	}

	badCookie := &http.Cookie{
		Name:     "__Host-refresh_token",
		Path:     "/",
		Value:    "invalidvalue",
		Expires:  time.Now().Add(refreshTokenExpiry),
		MaxAge:   int(refreshTokenExpiry.Seconds()),
		SameSite: http.SameSiteStrictMode,
		Domain:   "localhost",
		HttpOnly: true,
		Secure:   true,
	}

	testCases := []struct {
		name           string
		addCookie      bool
		cookie         *http.Cookie
		expectedStatus int
	}{
		{"valid cookie", true, testCookie, http.StatusOK},
		{"invalid valid cookie", true, badCookie, http.StatusBadRequest},
		{"no cookie", false, nil, http.StatusBadRequest},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/", nil)

			if tt.addCookie {
				req.AddCookie(tt.cookie)
			}

			handlerToTest := http.HandlerFunc(app.refreshUsingCookie)
			handlerToTest.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status code of %d; got %d", tt.expectedStatus, rr.Code)
			}
		})
	}

}

func Test_api_app_deleteRefreshCookie(t *testing.T) {

	req, _ := http.NewRequest("GET", "/logout", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(app.deleteRefreshCookie)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Errorf("expect status code %d; got %d", http.StatusAccepted, rr.Code)
	}

	cookieName := "__Host-refresh_token"

	foundCookie := false

	for _, c := range rr.Result().Cookies() {
		if c.Name == cookieName {
			foundCookie = true
			if c.Expires.After(time.Now()) {
				t.Errorf("cookie expiration in the future; the time should be in the past %v", c.Expires.UTC())
			}
			break
		}

	}

	if !foundCookie {
		t.Errorf("did not find cookie in response: %s", cookieName)
	}
}
