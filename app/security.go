package app

import (
	"context"
	"github.com/abbot/go-http-auth"
	log "github.com/sirupsen/logrus"
	"net/http"
)

// Authorize returns a SecretProvider
func Authorize(config *Config) auth.SecretProvider {
	return func(username, realm string) string {
		user := config.Users[username]

		if user != nil {
			return user.Password
		}

		log.WithField("user", username).Warn("Username not found")
		return ""
	}
}

// AuthWebdavHandler provides a ServeHTTP function with context and an application reference.
type AuthWebdavHandler interface {
	ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request, a *App)
}

// AuthWebdavHandlerFunc is a type definition which holds a context and application reference to
// match the AuthWebdavHandler interface.
type AuthWebdavHandlerFunc func(c context.Context, w http.ResponseWriter, r *http.Request, a *App)

// ServeHTTP simply calls the AuthWebdavHandlerFunc with given parameters
func (f AuthWebdavHandlerFunc) ServeHTTP(c context.Context, w http.ResponseWriter, r *http.Request, a *App) {
	f(c, w, r, a)
}

// Handle checks user authentification and calls the app related webdav handler.
func Handle(ctx context.Context, w http.ResponseWriter, r *http.Request, a *App) {
	authInfo := auth.FromContext(ctx)
	authInfo.UpdateHeaders(w.Header())
	if authInfo == nil || !authInfo.Authenticated {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	a.Handler.ServeHTTP(w, r.WithContext(ctx))
}

// AuthenticatedHandler returns a new http.Handler which creates a new context and calls
// the ServeHTTP function of the AuthWebdavHandler.
func AuthenticatedHandler(a auth.AuthenticatorInterface, h AuthWebdavHandler, application *App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := a.NewContext(context.Background(), r)

		h.ServeHTTP(ctx, w, r, application)
	})
}

// NewBasicAuthWebdavHandler creates a new http.Handler with a basic authenticator and a desired
// handler for basic auth and webdav.
func NewBasicAuthWebdavHandler(a *App) http.Handler {
	authenticator := auth.NewBasicAuthenticator(a.Config.Address, Authorize(a.Config))
	return AuthenticatedHandler(authenticator, AuthWebdavHandlerFunc(Handle), a)
}
