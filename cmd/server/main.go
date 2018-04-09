package main

import (
	"golang.org/x/net/webdav"
	"net/http"
)

func main() {
	wdHandler := &webdav.Handler{
		FileSystem: webdav.Dir("/tmp"),
		LockSystem: webdav.NewMemLS(),
	}

	http.ListenAndServe(":8000", wdHandler)
}
