package main

import (
	"golang.org/x/net/webdav"
	"net/http"
	"github.com/micromata/swd/app"
	"fmt"
	"log"
)

func main() {
	config := app.ParseConfig()

	wdHandler := &webdav.Handler{
		FileSystem: webdav.Dir("/Users/cclaus"),
		LockSystem: webdav.NewMemLS(),
	}

	connAddr := fmt.Sprintf("%s:%s", config.Address, config.Port)

	if config.TLS != nil {
		fmt.Printf("TLS Server is starting and listening at: %s\n", connAddr)
		log.Fatal(http.ListenAndServeTLS(connAddr, config.TLS.CertFile, config.TLS.KeyFile, wdHandler))
	} else {
		fmt.Printf("Server is starting and listening at: %s\n", connAddr)
		log.Fatal(http.ListenAndServe(connAddr, wdHandler))
	}
}
