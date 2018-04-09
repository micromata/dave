package main

import (
	"golang.org/x/net/webdav"
	"net/http"
	"github.com/micromata/swd/app"
	"fmt"
)

func main() {
	config := app.ParseConfig()

	wdHandler := &webdav.Handler{
		FileSystem: webdav.Dir("/tmp"),
		LockSystem: webdav.NewMemLS(),
	}

	connAddr := fmt.Sprintf("%s:%s", config.Address, config.Port)
	fmt.Printf("Server is starting and listening at: %s \n", connAddr)
	http.ListenAndServe(connAddr, wdHandler)
}
