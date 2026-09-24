package server

import "testing"

func TestContentTypeFor(t *testing.T) {
	cases := map[string]string{
		"":                "text/html; charset=utf-8",
		"about":           "text/html; charset=utf-8",
		"index.html":      "text/html; charset=utf-8",
		"static/site.css": "text/css; charset=utf-8",
		"favicon.svg":     "image/svg+xml",
		"app.js":          "text/javascript; charset=utf-8",
		"data.json":       "application/json",
	}
	for in, want := range cases {
		if got := contentTypeFor(in); got != want {
			t.Errorf("contentTypeFor(%q) = %q, want %q", in, got, want)
		}
	}
}
