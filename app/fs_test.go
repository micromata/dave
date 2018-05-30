package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestDirResolveUser(t *testing.T) {
	configTmp := createTestConfig("/tmp")

	ctx := context.Background()
	admin := context.WithValue(ctx, authInfoKey, &AuthInfo{Username: "admin", Authenticated: true})
	user1 := context.WithValue(ctx, authInfoKey, &AuthInfo{Username: "user1", Authenticated: true})
	anon := context.WithValue(ctx, authInfoKey, &AuthInfo{Username: "user1", Authenticated: false})

	tests := []struct {
		name string
		ctx  context.Context
		want string
	}{
		{"", admin, "admin"},
		{"", user1, "user1"},
		{"", anon, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Dir{
				Config: configTmp,
			}
			if got := d.resolveUser(tt.ctx); got != tt.want {
				t.Errorf("Dir.resolveUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

// This is a nearly concrete copy of the function TestDirResolve of golang.org/x/net/webdav/file_test.go
// just with prefixes and configuration details.
func TestDirResolve(t *testing.T) {
	configTmp := createTestConfig("/tmp")
	configRoot := createTestConfig("/")
	configCurrentDir := createTestConfig(".")
	configEmpty := createTestConfig("")

	ctx := context.Background()
	admin := context.WithValue(ctx, authInfoKey, &AuthInfo{Username: "admin", Authenticated: true})
	user1 := context.WithValue(ctx, authInfoKey, &AuthInfo{Username: "user1", Authenticated: true})
	user2 := context.WithValue(ctx, authInfoKey, &AuthInfo{Username: "user2", Authenticated: true})

	tests := []struct {
		cfg  *Config
		name string
		ctx  context.Context
		want string
	}{
		{configTmp, "", admin, "/tmp"},
		{configTmp, ".", admin, "/tmp"},
		{configTmp, "/", admin, "/tmp"},
		{configTmp, "./a", admin, "/tmp/a"},
		{configTmp, "..", admin, "/tmp"},
		{configTmp, "../", admin, "/tmp"},
		{configTmp, "../.", admin, "/tmp"},
		{configTmp, "../a", admin, "/tmp/a"},
		{configTmp, "../..", admin, "/tmp"},
		{configTmp, "../bar/a", admin, "/tmp/bar/a"},
		{configTmp, "../baz/a", admin, "/tmp/baz/a"},
		{configTmp, "...", admin, "/tmp/..."},
		{configTmp, ".../a", admin, "/tmp/.../a"},
		{configTmp, ".../..", admin, "/tmp"},
		{configTmp, "a", admin, "/tmp/a"},
		{configTmp, "a/./b", admin, "/tmp/a/b"},
		{configTmp, "a/../../b", admin, "/tmp/b"},
		{configTmp, "a/../b", admin, "/tmp/b"},
		{configTmp, "a/b", admin, "/tmp/a/b"},
		{configTmp, "a/b/c/../../d", admin, "/tmp/a/d"},
		{configTmp, "a/b/c/../../../d", admin, "/tmp/d"},
		{configTmp, "a/b/c/../../../../d", admin, "/tmp/d"},
		{configTmp, "a/b/c/d", admin, "/tmp/a/b/c/d"},
		{configTmp, "/a/b/c/d", admin, "/tmp/a/b/c/d"},

		{configTmp, "ab/c\x00d/ef", admin, ""},

		{configRoot, "", admin, "/"},
		{configRoot, ".", admin, "/"},
		{configRoot, "/", admin, "/"},
		{configRoot, "./a", admin, "/a"},
		{configRoot, "..", admin, "/"},
		{configRoot, "../", admin, "/"},
		{configRoot, "../.", admin, "/"},
		{configRoot, "../a", admin, "/a"},
		{configRoot, "../..", admin, "/"},
		{configRoot, "../bar/a", admin, "/bar/a"},
		{configRoot, "../baz/a", admin, "/baz/a"},
		{configRoot, "...", admin, "/..."},
		{configRoot, ".../a", admin, "/.../a"},
		{configRoot, ".../..", admin, "/"},
		{configRoot, "a", admin, "/a"},
		{configRoot, "a/./b", admin, "/a/b"},
		{configRoot, "a/../../b", admin, "/b"},
		{configRoot, "a/../b", admin, "/b"},
		{configRoot, "a/b", admin, "/a/b"},
		{configRoot, "a/b/c/../../d", admin, "/a/d"},
		{configRoot, "a/b/c/../../../d", admin, "/d"},
		{configRoot, "a/b/c/../../../../d", admin, "/d"},
		{configRoot, "a/b/c/d", admin, "/a/b/c/d"},
		{configRoot, "/a/b/c/d", admin, "/a/b/c/d"},

		{configCurrentDir, "", admin, "."},
		{configCurrentDir, ".", admin, "."},
		{configCurrentDir, "/", admin, "."},
		{configCurrentDir, "./a", admin, "a"},
		{configCurrentDir, "..", admin, "."},
		{configCurrentDir, "../", admin, "."},
		{configCurrentDir, "../.", admin, "."},
		{configCurrentDir, "../a", admin, "a"},
		{configCurrentDir, "../..", admin, "."},
		{configCurrentDir, "../bar/a", admin, "bar/a"},
		{configCurrentDir, "../baz/a", admin, "baz/a"},
		{configCurrentDir, "...", admin, "..."},
		{configCurrentDir, ".../a", admin, ".../a"},
		{configCurrentDir, ".../..", admin, "."},
		{configCurrentDir, "a", admin, "a"},
		{configCurrentDir, "a/./b", admin, "a/b"},
		{configCurrentDir, "a/../../b", admin, "b"},
		{configCurrentDir, "a/../b", admin, "b"},
		{configCurrentDir, "a/b", admin, "a/b"},
		{configCurrentDir, "a/b/c/../../d", admin, "a/d"},
		{configCurrentDir, "a/b/c/../../../d", admin, "d"},
		{configCurrentDir, "a/b/c/../../../../d", admin, "d"},
		{configCurrentDir, "a/b/c/d", admin, "a/b/c/d"},
		{configCurrentDir, "/a/b/c/d", admin, "a/b/c/d"},

		{configEmpty, "", admin, "."},

		{configTmp, "", user1, "/tmp/subdir1"},
		{configTmp, ".", user1, "/tmp/subdir1"},
		{configTmp, "/", user1, "/tmp/subdir1"},
		{configTmp, "./a", user1, "/tmp/subdir1/a"},
		{configTmp, "..", user1, "/tmp/subdir1"},
		{configTmp, "../", user1, "/tmp/subdir1"},
		{configTmp, "../.", user1, "/tmp/subdir1"},
		{configTmp, "../a", user1, "/tmp/subdir1/a"},
		{configTmp, "../..", user1, "/tmp/subdir1"},
		{configTmp, "../bar/a", user1, "/tmp/subdir1/bar/a"},
		{configTmp, "../baz/a", user1, "/tmp/subdir1/baz/a"},
		{configTmp, "...", user1, "/tmp/subdir1/..."},
		{configTmp, ".../a", user1, "/tmp/subdir1/.../a"},
		{configTmp, ".../..", user1, "/tmp/subdir1"},
		{configTmp, "a", user1, "/tmp/subdir1/a"},
		{configTmp, "a/./b", user1, "/tmp/subdir1/a/b"},
		{configTmp, "a/../../b", user1, "/tmp/subdir1/b"},
		{configTmp, "a/../b", user1, "/tmp/subdir1/b"},
		{configTmp, "a/b", user1, "/tmp/subdir1/a/b"},
		{configTmp, "a/b/c/../../d", user1, "/tmp/subdir1/a/d"},
		{configTmp, "a/b/c/../../../d", user1, "/tmp/subdir1/d"},
		{configTmp, "a/b/c/../../../../d", user1, "/tmp/subdir1/d"},
		{configTmp, "a/b/c/d", user1, "/tmp/subdir1/a/b/c/d"},
		{configTmp, "/a/b/c/d", user1, "/tmp/subdir1/a/b/c/d"},
		{configTmp, "", user2, "/tmp/subdir2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Dir{
				Config: tt.cfg,
			}
			if got := d.resolve(tt.ctx, tt.name); got != tt.want {
				t.Errorf("Dir.resolve() = %v, want %v. Base dir is %v", got, tt.want, tt.cfg.Dir)
			}
		})
	}
}

func TestDirMkdir(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "dave__"+strconv.FormatInt(time.Now().UnixNano(), 10))
	os.Mkdir(tmpDir, 0700)
	defer os.RemoveAll(tmpDir)

	t.Logf("using test dir: %s", tmpDir)
	configTmp := createTestConfig(tmpDir)

	ctx := context.Background()
	admin := context.WithValue(ctx, authInfoKey, &AuthInfo{Username: "admin", Authenticated: true})

	tests := []struct {
		name    string
		perm    os.FileMode
		wantErr bool
	}{
		{"a", 0700, false},
		{"/a/", 0700, true}, // already exists
		{"ab/c\x00d/ef", 0700, true},
		{"/a/a/a/a", 0700, true},
		{"a/a/a/a", 0700, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Dir{
				Config: configTmp,
			}
			if err := d.Mkdir(admin, tt.name, tt.perm); (err != nil) != tt.wantErr {
				t.Errorf("Dir.Mkdir() name = %v, error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}

func TestDirOpenFile(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "dave__"+strconv.FormatInt(time.Now().UnixNano(), 10))
	os.Mkdir(tmpDir, 0700)
	defer os.RemoveAll(tmpDir)
	configTmp := createTestConfig(tmpDir)

	ctx := context.Background()
	admin := context.WithValue(ctx, authInfoKey, &AuthInfo{Username: "admin", Authenticated: true})

	tests := []struct {
		name    string
		flag    int
		perm    os.FileMode
		wantErr bool
	}{
		{"foo", os.O_RDWR, 0644, true},
		{"foo", os.O_RDWR | os.O_CREATE, 0644, false},
		{"ab/c\x00d/ef", os.O_RDWR, 0700, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Dir{
				Config: configTmp,
			}
			got, err := d.OpenFile(admin, tt.name, tt.flag, tt.perm)
			if (err != nil) != tt.wantErr {
				t.Errorf("Dir.OpenFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				wantFileInfo, err := os.Stat(filepath.Join(tmpDir, tt.name))
				if err != nil {
					t.Errorf("Dir.OpenFile() error = %v", err)
				}

				gotFileInfo, _ := got.Stat()

				if !reflect.DeepEqual(gotFileInfo, wantFileInfo) {
					t.Errorf("Dir.OpenFile() = %v, want %v", gotFileInfo, wantFileInfo)
				}
			}
		})
	}
}

func TestRemoveDir(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "dave__"+strconv.FormatInt(time.Now().UnixNano(), 10))
	os.Mkdir(tmpDir, 0700)
	defer os.RemoveAll(tmpDir)
	configTmp := createTestConfig(tmpDir)

	ctx := context.Background()
	admin := context.WithValue(ctx, authInfoKey, &AuthInfo{Username: "admin", Authenticated: true})

	tests := []struct {
		name    string
		wantErr bool
	}{
		{"a", false},
		{"a/b/c", false},
		{"/a/b/c", false},
		{"ab/c\x00d/ef", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Dir{
				Config: configTmp,
			}

			file := filepath.Join(tmpDir, tt.name)
			if !tt.wantErr {
				err := os.MkdirAll(file, 0700)
				if err != nil {
					t.Errorf("Dir.RemoveAll() pre condition failed. name = %v, error = %v", tt.name, err)
				}
			}

			if err := d.RemoveAll(admin, tt.name); (err != nil) != tt.wantErr {
				t.Errorf("Dir.RemoveAll() name = %v, error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}

			if !tt.wantErr {
				if _, err := os.Stat(file); err == nil {
					t.Errorf("Dir.RemoveAll() post condition failed. name = %v, error = %v", tt.name, err)
				}
			}
		})
	}
}

func TestDirRemoveAll(t *testing.T) {
	ctx := context.Background()
	admin := context.WithValue(ctx, authInfoKey, &AuthInfo{Username: "admin", Authenticated: true})
	tmpDir := filepath.Join(os.TempDir(), "dave__"+strconv.FormatInt(time.Now().UnixNano(), 10))
	os.Mkdir(tmpDir, 0700)
	defer os.RemoveAll(tmpDir)
	configTmp := createTestConfig(tmpDir)

	tests := []struct {
		name       string
		removeName string
		wantErr    bool
	}{
		{"a/b/c", "a", false},
		{"a/b/c", "a/b", false},
		{"a/b/c", "a/b/c", false},
		{"/a/b/c", "a", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Dir{
				Config: configTmp,
			}

			err := os.MkdirAll(filepath.Join(tmpDir, tt.name), 0700)
			if err != nil {
				t.Errorf("Dir.RemoveAll() error creating dir error = %v, name %v", err, tt.name)
				return
			}

			if err := d.RemoveAll(admin, tt.removeName); (err != nil) != tt.wantErr {
				t.Errorf("Dir.RemoveAll() removeName = %v, error = %v, wantErr %v", tt.removeName, err, tt.wantErr)
				return
			}

			if _, err := os.Stat(filepath.Join(tmpDir, tt.removeName)); err == nil {
				t.Errorf("Dir.RemoveAll() file or directory not deleted = %v, removeName %v", err, tt.removeName)
				return
			}

			if _, err := os.Stat(filepath.Join(tmpDir, tt.removeName, "/..")); err != nil {
				t.Errorf("Dir.RemoveAll() parent directory deleted = %v, removeName %v", err, tt.removeName)
				return
			}
		})
	}
}

func TestRename(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "dave__"+strconv.FormatInt(time.Now().UnixNano(), 10))
	os.Mkdir(tmpDir, 0700)
	defer os.RemoveAll(tmpDir)
	configTmp := createTestConfig(tmpDir)

	ctx := context.Background()
	admin := context.WithValue(ctx, authInfoKey, &AuthInfo{Username: "admin", Authenticated: true})

	tests := []struct {
		name      string
		oldName   string
		newName   string
		create    bool
		wantError bool
	}{
		{"a", "a", "b", false, true},
		{"a", "a", "b", true, false},
		{"\x00d", "\x00da", "foo", false, true},
		{"\x00d", "foo", "\x00da", false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Dir{
				Config: configTmp,
			}

			if tt.create {
				_, err := d.OpenFile(admin, tt.oldName, os.O_RDWR|os.O_CREATE, 0700)
				if err != nil {
					t.Errorf("Dir.Rename() pre condition failed. name = %v, error = %v", tt.name, err)
					return
				}
			}

			if err := d.Rename(admin, tt.oldName, tt.newName); (err != nil) != tt.wantError {
				t.Errorf("Dir.Rename() error = %v, wantErr %v", err, tt.wantError)
				return
			}

			if _, err := os.Stat(filepath.Join(tmpDir, tt.oldName)); err == nil {
				t.Errorf("Dir.Rename() oldName still remained. oldName = %v, newName = %v", tt.oldName, tt.newName)
				return
			}

			join := filepath.Join(tmpDir, tt.newName)
			fmt.Println(join)
			if _, err := os.Stat(join); err != nil {
				if !tt.create {
					// If no file should be created, there can't be anything
					return
				}
				t.Errorf("Dir.Rename() newName not present. oldName = %v, newName = %v", tt.oldName, tt.newName)
				return
			}

		})
	}
}

func TestDirStat(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "dave__"+strconv.FormatInt(time.Now().UnixNano(), 10))
	os.Mkdir(tmpDir, 0700)
	defer os.RemoveAll(tmpDir)
	configTmp := createTestConfig(tmpDir)

	ctx := context.Background()
	admin := context.WithValue(ctx, authInfoKey, &AuthInfo{Username: "admin", Authenticated: true})

	tests := []struct {
		name    string
		kind    string
		wantErr bool
	}{
		{"/a", "dir", false},
		{"/a/b", "file", false},
		{"\x00da", "file", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Dir{
				Config: configTmp,
			}

			fp := filepath.Join(tmpDir, tt.name)
			if tt.kind == "dir" {
				err := os.MkdirAll(fp, 0700)
				if err != nil {
					t.Errorf("Dir.Stat() error creating dir. error = %v", err)
					return
				}
			} else if !tt.wantErr {
				_, err := os.OpenFile(fp, os.O_RDWR|os.O_CREATE, 0644)
				if err != nil {
					t.Errorf("Dir.Stat() error creating file. error = %v", err)
					return
				}
			}

			got, err := d.Stat(admin, tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Dir.Stat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			want, _ := os.Stat(filepath.Join(tmpDir, tt.name))

			if !reflect.DeepEqual(got, want) {
				t.Errorf("Dir.Stat() = %v, want %v", got, want)
			}
		})
	}
}

func createTestConfig(dir string) *Config {
	subdirs := [2]string{"subdir1", "subdir2"}
	userInfos := map[string]*UserInfo{
		"admin": {},
		"user1": {
			Subdir: &subdirs[0],
		},
		"user2": {
			Subdir: &subdirs[1],
		},
	}
	config := &Config{
		Dir:   dir,
		Users: userInfos,
		Log: Logging{
			Error:  true,
			Create: true,
			Read:   true,
			Update: true,
			Delete: true,
		},
	}
	return config
}
