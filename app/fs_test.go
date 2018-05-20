package app

import (
	"context"
	"testing"
	"os"
	"path/filepath"
	"time"
	"strconv"
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
	tmpDir := filepath.Join(os.TempDir(), "dave__"+strconv.FormatInt(time.Now().Unix(), 10))
	os.Mkdir(tmpDir, 0700)
	defer os.RemoveAll(tmpDir)

	t.Logf("using test dir: %s", tmpDir)
	configTmp := createTestConfig(tmpDir)

	ctx := context.Background()
	admin := context.WithValue(ctx, authInfoKey, &AuthInfo{Username: "admin", Authenticated: true})

	tests := []struct {
		name    string
		ctx     context.Context
		perm    os.FileMode
		wantErr bool
	}{
		{"a", admin, 0700, false},
		{"/a/", admin, 0700, true}, // already exists
		{"ab/c\x00d/ef", admin, 0700, true},
		{"/a/a/a/a", admin, 0700, true},
		{"a/a/a/a", admin, 0700, true},
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
