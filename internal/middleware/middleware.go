package middleware

import (
	"log"
	"net/http"

	"github.com/JustinLi007/whatdoing-server/internal/database"
	"github.com/JustinLi007/whatdoing-server/internal/tokens"
	"github.com/JustinLi007/whatdoing-server/internal/utils"
)

type Middleware struct {
	dbsUsers database.DbsUsers
	dbsJwt   database.DbsJwt
}

var allowedOrigins = map[string]bool{
	"http://localhost:5173": true,
}

var middlewareInstance *Middleware

func NewMiddleware(dbsUsers database.DbsUsers, dbsJwt database.DbsJwt) *Middleware {
	if middlewareInstance != nil {
		return middlewareInstance
	}

	newMiddleware := &Middleware{
		dbsUsers: dbsUsers,
		dbsJwt:   dbsJwt,
	}
	middlewareInstance = newMiddleware

	return middlewareInstance
}

func (m *Middleware) Cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// defer drainBuffer(r.Body)

		origin := r.Header.Get("Origin")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) RequireJwt(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = utils.SetUser(r, database.AnonymousUser)

		cookie, err := r.Cookie("whatdoing-jwt")
		if err != nil {
			log.Printf("error: middleware LoggedIn: jwt cookie: %v", err)
			next.ServeHTTP(w, r)
			return
		}

		jwtValidate := &tokens.Jwt{
			Token: &tokens.Token{
				PlainText: cookie.Value,
			},
		}
		user, err := m.dbsUsers.AuthenticateByJwt(jwtValidate)
		if err != nil {
			log.Printf("error: middleware LoggedIn: AuthenticateByJwt: %v", err)
			next.ServeHTTP(w, r)
			return
		}

		r = utils.SetUser(r, user)
		next.ServeHTTP(w, r)
		return
	})
}

func (m *Middleware) RequireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := utils.GetUser(r)
		if user == database.AnonymousUser {
			log.Printf("error: middleware RequireUser")
			utils.WriteJson(w, http.StatusUnauthorized, utils.Envelope{
				"error": "must be logged in",
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}
