package middleware

import (
	"fmt"
	"github.com/joefazee/ladiwork/data"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (m *Middleware) CheckRemember(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.App.Session.Exists(r.Context(), "userID") {
			// user not logged in
			cookie, err := r.Cookie(fmt.Sprintf("_%s_remember", m.App.AppName))
			if err != nil {
				next.ServeHTTP(w, r)
			} else {
				// we found a cookie, check it
				key := cookie.Value
				var u data.User
				if len(key) > 0 {
					split := strings.Split(key, "|")
					uid, hash := split[0], split[1] // todo: check for index
					id, _ := strconv.Atoi(uid)
					validHash := u.CheckForRememberToken(id, hash)
					if !validHash {
						m.deleteRememberCookie(w, r)
						m.App.Session.Put(r.Context(), "error", "You`ve been logged out from another device")
						next.ServeHTTP(w, r)
					} else {
						user, _ := u.Get(id)
						m.App.Session.Put(r.Context(), "userID", user.ID)
						m.App.Session.Put(r.Context(), "remember_token", hash)
						m.App.InfoLog.Printf("Just logged in user %d with remember token %s", user.ID, hash)
						next.ServeHTTP(w, r)
					}
				} else {
					// key length is zero, left over cookie
					m.deleteRememberCookie(w, r)
					next.ServeHTTP(w, r)
				}

			}
		} else {
			next.ServeHTTP(w, r)
		}

	})
}

func (m *Middleware) deleteRememberCookie(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.RenewToken(r.Context())

	newCookie := http.Cookie{
		Name:     fmt.Sprintf("_%s_remember", m.App.AppName),
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-100 * time.Hour),
		HttpOnly: true,
		Domain:   m.App.Session.Cookie.Domain,
		MaxAge:   -1,
		Secure:   m.App.Session.Cookie.Secure,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &newCookie)

	m.App.Session.Remove(r.Context(), "userID")
	m.App.Session.Destroy(r.Context())

	_ = m.App.Session.RenewToken(r.Context())
}
