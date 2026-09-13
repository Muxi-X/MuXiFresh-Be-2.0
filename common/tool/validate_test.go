package tool

import "testing"

func TestValidateAvatarURL(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty allowed", "", false},
		{"blank allowed", "   ", false},
		{"valid https", "https://ossfresh-test.muxixyz.com/avatar/abc/1--a.png", false},
		{"valid http", "http://example.com/a.png", false},
		{"valid with query", "https://ossfresh-test.muxixyz.com/avatar/abc/1--a.png?x=1", false},

		{"undefined segment", "https://ossfresh-test.muxixyz.com/undefined", true},
		{"null segment", "https://ossfresh-test.muxixyz.com/avatar/null", true},
		{"NaN segment", "https://ossfresh-test.muxixyz.com/avatar/NaN", true},
		{"object object", "https://ossfresh-test.muxixyz.com/[object Object]", true},
		{"javascript scheme", "javascript:alert(1)", true},
		{"missing scheme", "ossfresh-test.muxixyz.com/a.png", true},
		{"missing host", "https:///a.png", true},
		{"relative path", "/avatar/a.png", true},
		{"not a url", "not a url", true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := ValidateAvatarURL(c.input)
			if c.wantErr && err == nil {
				t.Fatalf("ValidateAvatarURL(%q) expected error, got nil", c.input)
			}
			if !c.wantErr && err != nil {
				t.Fatalf("ValidateAvatarURL(%q) unexpected error: %v", c.input, err)
			}
		})
	}
}
