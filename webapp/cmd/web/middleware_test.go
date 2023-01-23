package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"webapp/pkg/data"
)

func Test_application_addIpToContext(t *testing.T) {

	testCases := []struct {
		headerName  string
		headerValue string
		addr        string
		emeptyAddr  bool
	}{
		{"", "", "", false},
		{"", "", "", true},
		{"X-Forwaded-For", "198.3.2.1", "", false},
		{"", "", "hello:world", false},
	}
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		val := r.Context().Value(contextUserKey)

		if val == nil {
			t.Error(contextUserKey, "not present")
		}

		ip, ok := val.(string)
		if !ok {
			t.Error("not a string")
		}

		t.Log(ip)
	})

	for _, tt := range testCases {
		handlerToTest := app.addIpToContext(nextHandler)

		req := httptest.NewRequest("GET", "http://testing", nil)

		if tt.emeptyAddr {
			req.RemoteAddr = ""
		}

		if len(tt.headerName) > 0 {
			req.Header.Add(tt.headerName, tt.headerValue)
		}

		if len(tt.addr) > 0 {
			req.RemoteAddr = tt.addr
		}

		handlerToTest.ServeHTTP(httptest.NewRecorder(), req)
	}
}

func Test_application_iPFromContext(t *testing.T) {

	expected := "127.0.0.1"
	ctx := context.WithValue(context.Background(), contextUserKey, expected)

	result := app.ipFromContext(ctx)

	if expected != result {
		t.Errorf("expect context to have %v; got %v", expected, result)
	}
}

func Test_application_auth(t *testing.T) {

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})

	testCases := []struct {
		name   string
		isAuth bool
	}{
		{"logged in", true},
		{"not logged in", false},
	}

	for _, tt := range testCases {
		handlerToTest := app.auth(handler)

		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/user/profile", nil)
			req = addContextAndSessiontToRequest(req, app)
			if tt.isAuth {
				app.Session.Put(req.Context(), "user", data.User{ID: 1})
			}

			rr := httptest.NewRecorder()
			handlerToTest.ServeHTTP(rr, req)

			if tt.isAuth && rr.Code != http.StatusOK {
				t.Errorf("expect status code to be 200; got %d", rr.Code)
			}

			if !tt.isAuth && rr.Code != http.StatusTemporaryRedirect {
				t.Errorf("expect status code to be %d; got %d", http.StatusTemporaryRedirect, rr.Code)
			}
		})
	}
}
