package app

import (
	"github.com/abbot/go-http-auth"
	log "github.com/sirupsen/logrus"
	"net/http"
)

// ModificationLogHandler logs each incoming request, which has a method
// to create, update or delete a file or directory. If the request carries
// authentication information, the respective username will also be logged.
func ModificationLogHandler(r *http.Request, e error) {
	if r.Method == "PUT" || r.Method == "POST" || r.Method == "MKCOL" ||
		r.Method == "DELETE" || r.Method == "COPY" || r.Method == "MOVE" {

		var contextLogger *log.Entry
		authInfo := auth.FromContext(r.Context())

		if authInfo == nil || !authInfo.Authenticated {
			contextLogger = log.WithFields(log.Fields{
				"url":    r.URL.Path,
				"method": r.Method,
			})
		} else {
			contextLogger = log.WithFields(log.Fields{
				"url":    r.URL.Path,
				"method": r.Method,
				"user":   authInfo.Username,
			})
		}

		contextLogger.Info("File modified")
	}
}
