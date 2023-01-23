package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"testing"
	"webapp/pkg/data"
)

func Test_application_handlers(t *testing.T) {

	testCases := []struct {
		name                    string
		url                     string
		expectedCode            int
		expectedUrl             string
		expectedFirstStatusCode int
	}{
		{"home", "/", http.StatusOK, "/", http.StatusOK},
		{"404", "/not-found-routes", http.StatusNotFound, "/not-found-routes", http.StatusNotFound},
		{"profile", "/user/profile", http.StatusOK, "/", http.StatusTemporaryRedirect},
	}

	routes := app.routes()

	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	client := &http.Client{
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			res, err := ts.Client().Get(ts.URL + tt.url)
			if err != nil {
				t.Fatal(err)
			}

			if res.StatusCode != tt.expectedCode {
				t.Errorf("expect route '%v' status code to be %d; got %d", tt.url, tt.expectedCode, res.StatusCode)
			}

			if res.Request.URL.Path != tt.expectedUrl {
				t.Errorf("expect final url of to be %s; but got %s", tt.expectedUrl, res.Request.URL.Path)
			}

			resp2, _ := client.Get(ts.URL + tt.url)
			if resp2.StatusCode != tt.expectedFirstStatusCode {
				t.Errorf("expect first returned status code to be %d; got %d", tt.expectedFirstStatusCode, resp2.StatusCode)
			}
		})
	}
}

func TestApp_Home(t *testing.T) {

	testCases := []struct {
		name         string
		putInSession string
		expectedHTML string
	}{
		{"first vist", "", `<h2>Session:`},
		{"second visit", "hello world", "hello world"},
	}

	for _, tt := range testCases {

		req, _ := http.NewRequest("GET", "/", nil)
		req = addContextAndSessiontToRequest(req, app)

		app.Session.Destroy(req.Context())

		if tt.putInSession != "" {
			app.Session.Put(req.Context(), "test", tt.putInSession)
		}

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(app.home)

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expect home page to return 200 status code; got %d", rr.Code)
		}

		body, _ := io.ReadAll(rr.Body)
		got := string(body)
		if !strings.Contains(got, tt.expectedHTML) {
			t.Errorf("expect HTML '%v'; got '%s'", tt.expectedHTML, got)
		}

	}

}

func TestApp_render_bad_template(t *testing.T) {

	oldTemplatePath := templatePath
	templatePath = "./../../testdata/"

	req, _ := http.NewRequest("GET", "/", nil)

	rr := httptest.NewRecorder()

	err := app.render(rr, req, "bad-template.gohtml", &templateData{})
	if err == nil {
		t.Error("expected render() to return error for bad templates")
	}

	templatePath = oldTemplatePath

}
func getCtx(r *http.Request) context.Context {
	return context.WithValue(r.Context(), contextUserKey, "user")
}

func addContextAndSessiontToRequest(req *http.Request, app application) *http.Request {
	req = req.WithContext(getCtx(req))

	ctx, _ := app.Session.Load(req.Context(), req.Header.Get("X-Session"))

	return req.WithContext(ctx)
}

func Test_application_login(t *testing.T) {

	testCases := []struct {
		name               string
		postedData         url.Values
		expectedStatusCode int
		expectedLoc        string
		nillBody           bool
	}{
		{name: "valid user", postedData: url.Values{"email": {"admin@example.com"}, "password": {"secret"}}, expectedStatusCode: http.StatusSeeOther, expectedLoc: "/user/profile"},
		{name: "in-valid email", postedData: url.Values{"email": {"invalid-email@example.com"}, "password": {"secret"}}, expectedStatusCode: http.StatusSeeOther, expectedLoc: "/"},
		{name: "empty email and password", postedData: nil, expectedStatusCode: http.StatusSeeOther, expectedLoc: "/"},
		{name: "invalid login", postedData: url.Values{"email": {"admin@example.com"}, "password": {"invalid-password"}}, expectedStatusCode: http.StatusSeeOther, expectedLoc: "/"},
		{name: "invalid body", postedData: nil, expectedStatusCode: http.StatusBadRequest, expectedLoc: "/", nillBody: true},
		{name: "invalid res from sql", postedData: url.Values{"email": {"invalid@sql.com"}, "password": {"invalid-password"}}, expectedStatusCode: http.StatusSeeOther, expectedLoc: "/"},
	}

	for _, tt := range testCases {

		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/login", strings.NewReader(tt.postedData.Encode()))
			if err != nil {
				t.Fatal(err)
			}
			if tt.nillBody {
				req.Body = nil
			}

			req = addContextAndSessiontToRequest(req, app)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(app.login)
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatusCode {
				t.Errorf("expect status code %d, got %d", tt.expectedStatusCode, rr.Code)
			}

			loc, err := rr.Result().Location()
			if err != nil && err != http.ErrNoLocation {
				t.Error(err)
			}

			if len(tt.expectedLoc) > 0 && loc != nil && loc.String() != tt.expectedLoc {
				t.Errorf("expect Location header to be %s, got %s", tt.expectedLoc, loc)
			}
		})
	}

}

func Test_application_UploadFiles(t *testing.T) {

	// set up some pips
	pr, pw := io.Pipe()

	// create a new writer. of the type *io.Writer
	writer := multipart.NewWriter(pw)

	// create a wg
	wg := &sync.WaitGroup{}
	wg.Add(1)

	// simulate uploading a file using a gooutine and our writer
	go simulatePNGUpload("./testdata/img/test.png", writer, t, wg)

	// read frpm the pipe which receives the data
	request := httptest.NewRequest("POST", "/", pr)
	request.Header.Add("Content-Type", writer.FormDataContentType())

	// call the app.UploadFiles
	uploadedFiles, err := app.uploadFiles(request, "./testdata/uploads")
	if err != nil {
		t.Error(err)
	}

	// perform our tests
	newUploadedFile := fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].OriginalFileName)
	if _, err = os.Stat(newUploadedFile); os.IsNotExist(err) {
		t.Errorf("expected file to exists; %s", err.Error())
	}

	// clean up
	_ = os.Remove(newUploadedFile)
	wg.Wait()

}

func simulatePNGUpload(fileToUpload string, writer *multipart.Writer, t *testing.T, wg *sync.WaitGroup) {
	defer writer.Close()
	defer wg.Done()

	part, err := writer.CreateFormFile("file", path.Base(fileToUpload))
	if err != nil {
		t.Error(err)
	}

	f, err := os.Open(fileToUpload)
	if err != nil {
		t.Error(err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		t.Error(err)
	}

	err = png.Encode(part, img)

	if err != nil {
		t.Error(err)
	}

}

func Test_application_UploadProfilePic(t *testing.T) {
	uploadPath = "./testdata/uploads"
	filePath := "./testdata/img/test.png"

	fieldName := "file"

	body := new(bytes.Buffer)

	mw := multipart.NewWriter(body)
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatal(err)
	}

	w, err := mw.CreateFormFile(fieldName, filePath)
	if err != nil {
		t.Fatal(err)
	}

	if _, err = io.Copy(w, file); err != nil {
		t.Fatal(err)
	}

	mw.Close()

	req := httptest.NewRequest("POST", "/user/profile", body)
	req = addContextAndSessiontToRequest(req, app)
	app.Session.Put(req.Context(), "user", data.User{ID: 1})
	req.Header.Add("Content-Type", mw.FormDataContentType())

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.uploadProfilePicture)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("wrong status code; expect %d; got %d", http.StatusSeeOther, rr.Code)
	}

	_ = os.Remove(uploadPath + "/test.png")
}
