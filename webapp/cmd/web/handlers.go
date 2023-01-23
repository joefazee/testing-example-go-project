package main

import (
	"html/template"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"
	"webapp/pkg/data"
)

var (
	templatePath = "./template/"
	uploadPath   = "./static/img"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {

	tp := make(map[string]any)
	if !app.Session.Exists(r.Context(), "test") {
		tp["test"] = "No data"
		app.Session.Put(r.Context(), "test", "welcome "+time.Now().UTC().GoString())
	} else {
		tp["test"] = app.Session.GetString(r.Context(), "test")
	}
	app.render(w, r, "home.gohtml", &templateData{Data: tp})

}

type templateData struct {
	IP, Flash, Error string
	Data             map[string]any
	User             data.User
}

func (app *application) render(w http.ResponseWriter, r *http.Request, t string, td *templateData) error {

	tmpFiles := []string{
		path.Join(templatePath, t),
		path.Join(templatePath, "base.layout.gohtml"),
	}
	parsedTemplate, err := template.ParseFiles(tmpFiles...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	td.IP = app.ipFromContext(r.Context())
	td.Error = app.Session.PopString(r.Context(), "error")
	td.Flash = app.Session.GetString(r.Context(), "flash")

	if app.Session.Exists(r.Context(), "user") {
		td.User = app.Session.Get(r.Context(), "user").(data.User)
	}

	err = parsedTemplate.Execute(w, td)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	return nil
}

func (app *application) login(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	form := NewForm(r.PostForm)
	form.Required("email", "password")

	if !form.Valid() {
		app.redirectWithError(w, r, "/", "invalid login")
		return
	}

	email := r.PostForm.Get("email")
	password := r.PostForm.Get("password")

	user, err := app.DB.GetUserByEmail(email)
	if err != nil {
		app.redirectWithError(w, r, "/", "invalid login")
		return
	}

	if !app.authenticate(w, r, user, password) {
		app.redirectWithError(w, r, "/", "invalid login")
		return
	}

	_ = app.Session.RenewToken(r.Context())

	app.redirectWithMessage(w, r, "/user/profile", "flash", "successfully logged in!")

}

func (app *application) profilePage(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "profile.gohtml", &templateData{})
}

func (app *application) redirect(w http.ResponseWriter, r *http.Request, to string) {
	http.Redirect(w, r, to, http.StatusSeeOther)
}

func (app *application) redirectWithMessage(w http.ResponseWriter, r *http.Request, to, key, message string) {
	app.Session.Put(r.Context(), key, message)
	app.redirect(w, r, to)
}

func (app *application) redirectWithError(w http.ResponseWriter, r *http.Request, to, message string) {
	app.redirectWithMessage(w, r, to, "error", message)
}

func (app *application) authenticate(w http.ResponseWriter, r *http.Request, user *data.User, password string) bool {

	if valid, err := user.PasswordMatches(password); err != nil || !valid {
		return false
	}
	app.Session.Put(r.Context(), "user", user)

	return true
}

func (app *application) uploadProfilePicture(w http.ResponseWriter, r *http.Request) {

	files, err := app.uploadFiles(r, uploadPath)
	if err != nil {
		app.redirectWithMessage(w, r, "/user/profile", "error", err.Error())
		return
	}

	user := app.Session.Get(r.Context(), "user").(data.User)

	var userImg = data.UserImage{
		UserID:   user.ID,
		FileName: files[0].OriginalFileName,
	}

	_, err = app.DB.InsertUserImage(userImg)
	if err != nil {
		app.redirectWithError(w, r, "/user/profile", err.Error())
		return
	}

	updatedUser, err := app.DB.GetUser(user.ID)
	if err != nil {
		app.redirectWithError(w, r, "/user/profile", err.Error())
		return
	}

	app.Session.Put(r.Context(), "user", updatedUser)

	app.redirectWithMessage(w, r, "/user/profile", "flash", "profile photo uploaded!")

}

type UploadedFile struct {
	OriginalFileName string
	FileSize         int64
}

func (app *application) uploadFiles(r *http.Request, uploadDir string) ([]*UploadedFile, error) {

	var uploadedFiles []*UploadedFile

	err := r.ParseMultipartForm(int64(1024 * 1024 * 5))
	if err != nil {
		return nil, err
	}

	for _, fHeaders := range r.MultipartForm.File {
		for _, hdr := range fHeaders {
			uploadedFiles, err = func(uploadedFiles []*UploadedFile) ([]*UploadedFile, error) {

				var uploadeFile UploadedFile
				infile, err := hdr.Open()
				if err != nil {
					return nil, err
				}

				defer infile.Close()
				uploadeFile.OriginalFileName = hdr.Filename

				var outfile *os.File
				defer outfile.Close()

				if outfile, err = os.Create(filepath.Join(uploadDir, uploadeFile.OriginalFileName)); err != nil {
					return nil, err
				} else {
					fileSize, err := io.Copy(outfile, infile)
					if err != nil {
						return nil, err
					}
					uploadeFile.FileSize = fileSize
				}

				uploadedFiles = append(uploadedFiles, &uploadeFile)
				return uploadedFiles, nil

			}(uploadedFiles)

			if err != nil {
				return uploadedFiles, err
			}
		}
	}

	return uploadedFiles, nil
}
