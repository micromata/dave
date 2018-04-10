package main

import (
	"fmt"
	"github.com/abbot/go-http-auth"
	"github.com/micromata/swd/app"
	"golang.org/x/net/webdav"
	"log"
	"net/http"
)

func main() {
	config := app.ParseConfig()

	wdHandler := &webdav.Handler{
		Prefix:     config.Prefix,
		FileSystem: webdav.Dir(config.Dir),
		LockSystem: webdav.NewMemLS(),
	}

	a := &app.App{
		Config:  config,
		Handler: wdHandler,
	}

	authenticator := auth.NewBasicAuthenticator(config.Address, app.Authorize(config))
	http.Handle("/", app.AuthenticatedHandler(authenticator, app.AuthWebdavHandlerFunc(app.Handle), a))

	connAddr := fmt.Sprintf("%s:%s", config.Address, config.Port)

	if config.TLS != nil {
		fmt.Printf("TLS Server is starting and listening at: %s\n", connAddr)
		log.Fatal(http.ListenAndServeTLS(connAddr, config.TLS.CertFile, config.TLS.KeyFile, nil))
	} else {
		fmt.Printf("Server is starting and listening at: %s\n", connAddr)
		log.Fatal(http.ListenAndServe(connAddr, nil))
	}
}
