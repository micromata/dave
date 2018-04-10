// Package app provides all app related stuff like config parsing, serving, etc.
package app

import "golang.org/x/net/webdav"

// App holds configuration information and the webdav handler.
type App struct {
	Config  *Config
	Handler *webdav.Handler
}
