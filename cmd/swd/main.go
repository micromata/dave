package main

import (
	"fmt"
	"github.com/micromata/swd/app"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/webdav"
	"net/http"
)

func main() {
	log.SetFormatter(&log.TextFormatter{})

	config := app.ParseConfig()

	wdHandler := &webdav.Handler{
		Prefix: config.Prefix,
		FileSystem: &app.Dir{
			Config: config,
		},
		LockSystem: webdav.NewMemLS(),
		Logger: func(request *http.Request, err error) {
			if err != nil {
				log.Error(err)
			}
		},
	}

	a := &app.App{
		Config:  config,
		Handler: wdHandler,
	}

	http.Handle("/", app.NewBasicAuthWebdavHandler(a))

	connAddr := fmt.Sprintf("%s:%s", config.Address, config.Port)

	if config.TLS != nil {
		log.WithFields(log.Fields{
			"address":  config.Address,
			"port":     config.Port,
			"security": "TLS",
		}).Info("Server is starting and listening")
		log.Fatal(http.ListenAndServeTLS(connAddr, config.TLS.CertFile, config.TLS.KeyFile, nil))
	} else {
		log.WithFields(log.Fields{
			"address":  config.Address,
			"port":     config.Port,
			"security": "none",
		}).Info("Server is starting and listening")
		log.Fatal(http.ListenAndServe(connAddr, nil))
	}
}
