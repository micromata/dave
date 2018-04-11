package app

import (
	"context"
	"github.com/abbot/go-http-auth"
	"golang.org/x/net/webdav"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// This file is an extension of golang.org/x/net/webdav/file.go.

// Dir is specialization of webdav.Dir with respect of an authenticated
// user to allow configuration access.
type Dir struct {
	Config *Config
}

// resolve tries to gain authentication information and suffixes the BaseDir with the
// username of the authentication information. If none authentication information can
// achieved during the process, the BaseDir is used
func (d Dir) resolve(ctx context.Context, name string) string {
	// This implementation is based on Dir.Open's code in the standard net/http package.
	if filepath.Separator != '/' && strings.IndexRune(name, filepath.Separator) >= 0 ||
		strings.Contains(name, "\x00") {
		return ""
	}
	dir := string(d.Config.Dir)
	if dir == "" {
		dir = "."
	}

	authInfo := auth.FromContext(ctx)
	if authInfo != nil && authInfo.Authenticated {
		userInfo := d.Config.Users[authInfo.Username]
		if userInfo != nil && userInfo.Subdir != nil {
			return filepath.Join(dir, authInfo.Username, filepath.FromSlash(path.Clean("/"+name)))
		}
	}

	return filepath.Join(dir, filepath.FromSlash(path.Clean("/"+name)))
}

// Mkdir resolves the physical file and delegates this to an os.Mkdir execution
func (d Dir) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	if name = d.resolve(ctx, name); name == "" {
		return os.ErrNotExist
	}
	return os.Mkdir(name, perm)
}

// OpenFile resolves the physical file and delegates this to an os.OpenFile execution
func (d Dir) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	if name = d.resolve(ctx, name); name == "" {
		return nil, os.ErrNotExist
	}
	f, err := os.OpenFile(name, flag, perm)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// RemoveAll resolves the physical file and delegates this to an os.RemoveAll execution
func (d Dir) RemoveAll(ctx context.Context, name string) error {
	if name = d.resolve(ctx, name); name == "" {
		return os.ErrNotExist
	}
	if name == filepath.Clean(string(d.Config.Dir)) {
		// Prohibit removing the virtual root directory.
		return os.ErrInvalid
	}
	return os.RemoveAll(name)
}

// Rename resolves the physical file and delegates this to an os.Rename execution
func (d Dir) Rename(ctx context.Context, oldName, newName string) error {
	if oldName = d.resolve(ctx, oldName); oldName == "" {
		return os.ErrNotExist
	}
	if newName = d.resolve(ctx, newName); newName == "" {
		return os.ErrNotExist
	}
	if root := filepath.Clean(string(d.Config.Dir)); root == oldName || root == newName {
		// Prohibit renaming from or to the virtual root directory.
		return os.ErrInvalid
	}
	return os.Rename(oldName, newName)
}

// Stat resolves the physical file and delegates this to an os.Stat execution
func (d Dir) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	if name = d.resolve(ctx, name); name == "" {
		return nil, os.ErrNotExist
	}
	return os.Stat(name)
}
