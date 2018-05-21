package app

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

type contextKey int

var authInfoKey contextKey

// AuthInfo holds the username and authentication status
type AuthInfo struct {
	Username      string
	Authenticated bool
}

// authWebdavHandlerFunc is a type definition which holds a context and application reference to
// match the AuthWebdavHandler interface.
type authWebdavHandlerFunc func(c context.Context, w http.ResponseWriter, r *http.Request, a *App)

// ServeHTTP simply calls the AuthWebdavHandlerFunc with given parameters
func (f authWebdavHandlerFunc) ServeHTTP(c context.Context, w http.ResponseWriter, r *http.Request, a *App) {
	f(c, w, r, a)
}

// NewBasicAuthWebdavHandler creates a new http handler with basic auth features.
// The handler will use the application config for user and password lookups.
func NewBasicAuthWebdavHandler(a *App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		handlerFunc := authWebdavHandlerFunc(handle)
		handlerFunc.ServeHTTP(ctx, w, r, a)
	})
}

func authorize(config *Config, username, password string) *AuthInfo {
	if username == "" || password == "" {
		log.WithField("user", username).Warn("Username not found or password empty")
		return &AuthInfo{Authenticated: false}
	}

	user := config.Users[username]
	if user == nil {
		log.WithField("user", username).Warn("User not found")
		return &AuthInfo{Authenticated: false}
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		log.WithField("user", username).Warn("Password doesn't match")
		return &AuthInfo{Authenticated: false}
	}

	return &AuthInfo{Username: username, Authenticated: true}
}

// AuthFromContext returns information about the authentication state of the current user.
func AuthFromContext(ctx context.Context) *AuthInfo {
	info, ok := ctx.Value(authInfoKey).(*AuthInfo)
	if !ok {
		return nil
	}

	return info
}

func handle(ctx context.Context, w http.ResponseWriter, r *http.Request, a *App) {
	username, password, ok := r.BasicAuth()

	if !ok {
		writeUnauthorized(w, a.Config.Realm)
		return
	}

	authInfo := authorize(a.Config, username, password)
	if !authInfo.Authenticated {
		writeUnauthorized(w, a.Config.Realm)
		return
	}

	ctx = context.WithValue(ctx, authInfoKey, authInfo)
	a.Handler.ServeHTTP(w, r.WithContext(ctx))
}

func writeUnauthorized(w http.ResponseWriter, realm string) {
	w.Header().Set("WWW-Authenticate", "Basic realm="+realm)
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(fmt.Sprintf("%d %s", http.StatusUnauthorized, "Unauthorized")))
}
