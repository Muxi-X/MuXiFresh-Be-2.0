package tool

import "testing"

func TestValidateAvatarURL(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"empty allowed", "", "", false},
		{"blank normalized to empty", "   ", "", false},
		{"padded url trimmed", "  https://example.com/a.png  ", "https://example.com/a.png", false},
		{"valid https", "https://ossfresh-test.muxixyz.com/avatar/abc/1--a.png", "https://ossfresh-test.muxixyz.com/avatar/abc/1--a.png", false},
		{"valid http", "http://example.com/a.png", "http://example.com/a.png", false},
		{"valid with query", "https://ossfresh-test.muxixyz.com/avatar/abc/1--a.png?x=1", "https://ossfresh-test.muxixyz.com/avatar/abc/1--a.png?x=1", false},
		{"filename containing undefined", "https://example.com/undefined.png", "https://example.com/undefined.png", false},

		{"undefined segment", "https://ossfresh-test.muxixyz.com/undefined", "", true},
		{"null segment", "https://ossfresh-test.muxixyz.com/avatar/null", "", true},
		{"NaN segment", "https://ossfresh-test.muxixyz.com/avatar/NaN", "", true},
		{"undefined in query", "https://example.com/avatar.png?key=undefined", "", true},
		{"undefined in fragment", "https://example.com/avatar.png#undefined", "", true},
		{"object object", "https://ossfresh-test.muxixyz.com/[object Object]", "", true},
		{"javascript scheme", "javascript:alert(1)", "", true},
		{"missing scheme", "ossfresh-test.muxixyz.com/a.png", "", true},
		{"missing host", "https:///a.png", "", true},
		{"host without hostname", "https://:8080/a.png", "", true},
		{"relative path", "/avatar/a.png", "", true},
		{"not a url", "not a url", "", true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ValidateAvatarURL(c.input)
			if c.wantErr {
				if err == nil {
					t.Fatalf("ValidateAvatarURL(%q) expected error, got %q", c.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ValidateAvatarURL(%q) unexpected error: %v", c.input, err)
			}
			if got != c.want {
				t.Fatalf("ValidateAvatarURL(%q) = %q, want %q", c.input, got, c.want)
			}
		})
	}
}
