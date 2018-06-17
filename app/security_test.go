package app

import (
	"context"
	"golang.org/x/net/webdav"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestAuthenticate(t *testing.T) {
	type args struct {
		config   *Config
		username string
		password string
	}
	tests := []struct {
		name    string
		args    args
		want    *AuthInfo
		wantErr bool
	}{
		{
			"empty username",
			args{
				config: &Config{Users: map[string]*UserInfo{
					"foo": {
						Password: GenHash([]byte("password")),
					},
				}},
				username: "",
				password: "password",
			},
			&AuthInfo{
				Username:      "",
				Authenticated: false,
			},
			true,
		},
		{
			"empty password",
			args{
				config: &Config{Users: map[string]*UserInfo{
					"foo": {
						Password: GenHash([]byte("password")),
					},
				}},
				username: "foo",
				password: "",
			},
			&AuthInfo{
				Username:      "foo",
				Authenticated: false,
			},
			true,
		},
		{
			"empty username without users",
			args{
				config:   &Config{},
				username: "",
				password: "password",
			},
			&AuthInfo{
				Username:      "",
				Authenticated: false,
			},
			false,
		},
		{
			"empty password without users",
			args{
				config:   &Config{},
				username: "foo",
				password: "",
			},
			&AuthInfo{
				Username:      "",
				Authenticated: false,
			},
			false,
		},
		{
			"user not found",
			args{
				config: &Config{Users: map[string]*UserInfo{
					"bar": nil,
				}},
				username: "foo",
				password: "password",
			},
			&AuthInfo{
				Username:      "foo",
				Authenticated: false,
			},
			true,
		},
		{
			"password doesn't match",
			args{
				config: &Config{Users: map[string]*UserInfo{
					"foo": {
						Password: GenHash([]byte("not-my-password")),
					},
				}},
				username: "foo",
				password: "password",
			},
			&AuthInfo{
				Username:      "foo",
				Authenticated: false,
			},
			true,
		},
		{
			"all fine",
			args{
				config: &Config{Users: map[string]*UserInfo{
					"foo": {
						Password: GenHash([]byte("password")),
					},
				}},
				username: "foo",
				password: "password",
			},
			&AuthInfo{
				Username:      "foo",
				Authenticated: true,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := authenticate(tt.args.config, tt.args.username, tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("authenticate() name = %v, error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("authenticate() name = %v, got = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestAuthFromContext(t *testing.T) {
	type fakeKey int
	var fakeKeyValue fakeKey

	baseCtx := context.Background()
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		args args
		want *AuthInfo
	}{
		{
			"success",
			args{
				ctx: context.WithValue(baseCtx, authInfoKey, &AuthInfo{"username", true}),
			},
			&AuthInfo{"username", true},
		},
		{
			"failure",
			args{
				ctx: context.WithValue(baseCtx, fakeKeyValue, &AuthInfo{"username", true}),
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AuthFromContext(tt.args.ctx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AuthFromContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandle(t *testing.T) {
	type args struct {
		ctx      context.Context
		w        *httptest.ResponseRecorder
		r        *http.Request
		username []byte
		password []byte
		a        *App
	}
	tests := []struct {
		name       string
		args       args
		statusCode int
	}{
		{
			"basic auth error",
			args{
				context.Background(),
				httptest.NewRecorder(),
				httptest.NewRequest("PROPFIND", "/", nil),
				nil,
				nil,
				&App{Config: &Config{Users: map[string]*UserInfo{
					"foo": {
						Password: GenHash([]byte("password")),
					},
				}}},
			},
			401,
		},
		{
			"unauthorized error",
			args{
				context.Background(),
				httptest.NewRecorder(),
				httptest.NewRequest("PROPFIND", "/", nil),
				[]byte("u"),
				[]byte("p"),
				&App{Config: &Config{Users: map[string]*UserInfo{
					"foo": {
						Password: GenHash([]byte("password")),
					},
				}}},
			},
			401,
		},
		{
			"ok",
			args{
				context.Background(),
				httptest.NewRecorder(),
				httptest.NewRequest("PROPFIND", "/", nil),
				[]byte("foo"),
				[]byte("password"),
				&App{
					Config: &Config{Users: map[string]*UserInfo{
						"foo": {
							Password: GenHash([]byte("password")),
						},
					}},
					Handler: &webdav.Handler{
						FileSystem: webdav.NewMemFS(),
						LockSystem: webdav.NewMemLS(),
					},
				},
			},
			207,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.username != nil || tt.args.password != nil {
				tt.args.r.SetBasicAuth(string(tt.args.username), string(tt.args.password))
			}

			handle(tt.args.ctx, tt.args.w, tt.args.r, tt.args.a)
			resp := tt.args.w.Result()

			if resp.StatusCode != tt.statusCode {
				t.Errorf("TestHandle() = %v, want %v", resp.StatusCode, tt.statusCode)
			}
		})
	}
}
