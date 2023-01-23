package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestForm_Has(t *testing.T) {

	form := NewForm(nil)

	has := form.Has("invalid-field")

	if has {
		t.Error("form.Has should return false")
	}

	formData := url.Values{}
	formData.Add("name", "john")
	form = NewForm(formData)

	has = form.Has("name")

	if !has {
		t.Error("form.Has should return true")
	}

}

func TestForm_required(t *testing.T) {

	r := httptest.NewRequest("POST", "/action", nil)

	form := NewForm(r.PostForm)
	form.Required("name", "email")

	if form.Valid() {
		t.Error("form shows valid when required fields are missing")
	}

	postedData := url.Values{}
	postedData.Add("name", "john")
	postedData.Add("email", "john@doe.com")

	r, _ = http.NewRequest("POST", "/action", nil)
	r.PostForm = postedData
	form = NewForm(r.PostForm)
	form.Required("name", "email")

	if !form.Valid() {
		t.Error("shows post does not have required field when it does")
	}

}

func TestForm_check(t *testing.T) {

	form := NewForm(nil)

	form.Check(false, "email", "password is required")

	if form.Valid() {
		t.Error("form.Valid() returned true when it should be false")
	}
}

func TestForm_ErrorGet(t *testing.T) {

	form := NewForm(nil)

	err := "password is required"
	form.Check(false, "email", err)

	s := form.Errors.Get("email")

	if err != s {
		t.Errorf("expect Errors.Get() to return error message '%v'; got '%v'", err, s)
	}

	s = form.Errors.Get("invalid-field")
	if s != "" {
		t.Errorf("we expect Errors.Get() to return empty string for invalid-fields; got %v", s)
	}
}
