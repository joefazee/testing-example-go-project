package main

import (
	"net/http"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func Test_application_routes(t *testing.T) {

	testCases := []struct {
		route  string
		method string
	}{
		{"/", "GET"},
		{"/login", "POST"},
		{"/user/profile", "GET"},
		{"/static/*", "GET"},
	}

	mux := app.routes()

	chiRoutes := mux.(chi.Routes)

	for _, route := range testCases {
		if !routeExists(route.route, route.method, chiRoutes) {
			t.Errorf("expect %s %s in routes;", route.method, route.route)
		}
	}

}

func routeExists(testRoute string, testMethod string, chiRoutes chi.Routes) bool {

	found := false

	_ = chi.Walk(chiRoutes, func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {

		if strings.EqualFold(method, testMethod) && strings.EqualFold(route, testRoute) {
			found = true
		}

		return nil
	})

	return found
}
