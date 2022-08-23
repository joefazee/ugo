package ugo

import (
	"github.com/justinas/nosurf"
	"net/http"
	"strconv"
)

func (u *Ugo) SessionLoad(next http.Handler) http.Handler {
	return u.Session.LoadAndSave(next)
}

func (u *Ugo) NoSurf(next http.Handler) http.Handler {

	csrfHandler := nosurf.New(next)
	secure, _ := strconv.ParseBool(u.config.cookie.secure)

	csrfHandler.ExemptGlob("/api/*")

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		Domain:   u.config.cookie.domain,
	})

	return csrfHandler
}
