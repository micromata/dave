package main

import (
	"fmt"
	"github.com/micromata/swd/app"
	"golang.org/x/net/webdav"
	"log"
	"net/http"
)

func main() {
	config := app.ParseConfig()

	wdHandler := &webdav.Handler{
		Prefix: config.Prefix,
		FileSystem: &app.UserDir{
			BaseDir: config.Dir,
		},
		LockSystem: webdav.NewMemLS(),
		Logger:     app.ModificationLogHandler,
	}

	a := &app.App{
		Config:  config,
		Handler: wdHandler,
	}

	http.Handle("/", app.NewBasicAuthWebdavHandler(a))

	connAddr := fmt.Sprintf("%s:%s", config.Address, config.Port)

	if config.TLS != nil {
		fmt.Printf("TLS Server is starting and listening at: %s\n", connAddr)
		log.Fatal(http.ListenAndServeTLS(connAddr, config.TLS.CertFile, config.TLS.KeyFile, nil))
	} else {
		fmt.Printf("Server is starting and listening at: %s\n", connAddr)
		log.Fatal(http.ListenAndServe(connAddr, nil))
	}
}
