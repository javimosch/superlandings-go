package services

import "testing"

func TestResolveHotifyApp(t *testing.T) {
	apps := []hotifyApp{
		{ID: "mago", Domain: "mago.intrane.fr", Port: 9100},       // someone else's app
		{ID: "mago-go", Domain: "go.mago.intrane.fr", Port: 3099}, // legacy id, ours
		{ID: "sl-docs", Domain: "docs.example.fr", Port: 3099},
		{ID: "sl-other", Domain: "other.example.fr", Port: 4000},
	}
	cases := []struct {
		slug, domain string
		wantID       string
		wantExists   bool
		wantErr      bool
	}{
		{"mago", "new.example.fr", "sl-mago", false, false},       // never the live "mago" app
		{"mago", "mago.intrane.fr", "", false, true},              // domain taken by another app
		{"mago-go", "go.mago.intrane.fr", "mago-go", true, false}, // legacy app on our port is adopted
		{"docs", "docs.example.fr", "sl-docs", true, false},
		{"other", "", "", false, true}, // sl- id but another port
		{"fresh", "", "sl-fresh", false, false},
	}
	for _, c := range cases {
		id, exists, err := resolveHotifyApp(apps, c.slug, c.domain, 3099)
		if (err != nil) != c.wantErr || id != c.wantID || exists != c.wantExists {
			t.Errorf("resolveHotifyApp(%q, %q) = %q, %v, %v; want %q, %v, err=%v",
				c.slug, c.domain, id, exists, err, c.wantID, c.wantExists, c.wantErr)
		}
	}
}
