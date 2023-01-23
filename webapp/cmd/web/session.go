package main

import (
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
)

func getSession() *scs.SessionManager {
	sess := scs.New()
	sess.Lifetime = 24 * time.Hour
	sess.Cookie.Persist = true
	sess.Cookie.SameSite = http.SameSiteLaxMode
	sess.Cookie.Secure = true

	return sess
}
