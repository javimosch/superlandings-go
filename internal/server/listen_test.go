package server

import "testing"

func TestListenAddr(t *testing.T) {
	cases := []struct {
		bind, token, want string
		wantErr           bool
	}{
		{"", "", "127.0.0.1:3099", false},
		{"127.0.0.1", "", "127.0.0.1:3099", false},
		{"localhost", "", "localhost:3099", false},
		{"::1", "", "[::1]:3099", false},
		{"0.0.0.0", "", "", true},
		{"92.113.145.178", "", "", true},
		{"0.0.0.0", "secret", "0.0.0.0:3099", false},
	}
	for _, c := range cases {
		got, err := listenAddr(c.bind, 3099, c.token)
		if (err != nil) != c.wantErr || got != c.want {
			t.Errorf("listenAddr(%q, token=%q) = %q, %v; want %q, err=%v", c.bind, c.token, got, err, c.want, c.wantErr)
		}
	}
}
