package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/webdav"
)

// This file is an extension of golang.org/x/net/webdav/file.go.

// Dir is specialization of webdav.Dir with respect of an authenticated
// user to allow configuration access.
type Dir struct {
	Config *Config
}

func (d Dir) resolveUser(ctx context.Context) string {
	authInfo := AuthFromContext(ctx)
	if authInfo != nil && authInfo.Authenticated {
		return authInfo.Username
	}

	return ""
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

	// Second barrier after basic auth process
	authInfo := AuthFromContext(ctx)
	if authInfo != nil && authInfo.Authenticated {
		userInfo := d.Config.Users[authInfo.Username]
		if userInfo != nil && userInfo.Subdir != nil {
			return filepath.Join(dir, *userInfo.Subdir, filepath.FromSlash(path.Clean("/"+name)))
		}
	}

	return filepath.Join(dir, filepath.FromSlash(path.Clean("/"+name)))
}

// Mkdir resolves the physical file and delegates this to an os.Mkdir execution
func (d Dir) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	if name = d.resolve(ctx, name); name == "" {
		return os.ErrNotExist
	}

	for _, v := range d.Config.Deny.Create.Directory {
		matched, err := filepath.Match(v, filepath.Base(name))
		if err != nil {
			return err
		}
		if matched {
			return errors.New(fmt.Sprintf("mkdir %s, action denied", name))
		}
	}

	err := os.Mkdir(name, perm)
	if err != nil {
		return err
	}

	if d.Config.Log.Create {
		log.WithFields(log.Fields{
			"path": name,
			"user": d.resolveUser(ctx),
		}).Info("Created directory")
	}

	return err
}

// OpenFile resolves the physical file and delegates this to an os.OpenFile execution
func (d Dir) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	if name = d.resolve(ctx, name); name == "" {
		return nil, os.ErrNotExist
	}
	if len(d.Config.Deny.Create.File) > 0 {
		// os.O_RDONLY: 0, os.O_RDWR: 2, os.O_CREATE: 512, O_TRUNC: 1024
		if flag == os.O_RDWR|os.O_CREATE|os.O_TRUNC || flag == os.O_RDWR|os.O_CREATE || flag == os.O_CREATE|os.O_TRUNC || flag == os.O_CREATE {
			for _, v := range d.Config.Deny.Create.File {
				matched, err := filepath.Match(v, filepath.Base(name))
				if err != nil {
					return nil, err
				}
				if matched {
					return nil, errors.New(fmt.Sprintf("create %s, action denied", name))
				}
			}
		}
	}

	f, err := os.OpenFile(name, flag, perm)
	if err != nil {
		return nil, err
	}

	if d.Config.Log.Read {
		log.WithFields(log.Fields{
			"path": name,
			"user": d.resolveUser(ctx),
		}).Info("Opened file")
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

	err := os.RemoveAll(name)
	if err != nil {
		return err
	}

	if d.Config.Log.Delete {
		log.WithFields(log.Fields{
			"path": name,
			"user": d.resolveUser(ctx),
		}).Info("Deleted file or directory")
	}

	return nil
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

	err := os.Rename(oldName, newName)
	if err != nil {
		return err
	}

	if d.Config.Log.Update {
		log.WithFields(log.Fields{
			"oldPath": oldName,
			"newPath": newName,
			"user":    d.resolveUser(ctx),
		}).Info("Renamed file or directory")
	}

	return nil
}

// Stat resolves the physical file and delegates this to an os.Stat execution
func (d Dir) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	if name = d.resolve(ctx, name); name == "" {
		return nil, os.ErrNotExist
	}
	return os.Stat(name)
}
