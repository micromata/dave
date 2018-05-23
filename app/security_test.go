package app

import (
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
				config:   &Config{},
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
				config:   &Config{},
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
