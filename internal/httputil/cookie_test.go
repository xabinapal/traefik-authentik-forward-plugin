package httputil

import "testing"

func TestParseCookieName(t *testing.T) {
	t.Run("with valid cookie", func(t *testing.T) {
		cookie := "test=value"
		name := ParseCookieName(cookie)

		// check that the name is parsed correctly
		if name != "test" {
			t.Errorf("expected name to be test, got %s", name)
		}
	})

	tests := []struct {
		name  string
		value string
	}{
		{
			name:  "with empty string",
			value: "",
		},
		{
			name:  "with no equal sign",
			value: "value",
		},
		{
			name:  "with empty before equal",
			value: "=value",
		},
		{
			name:  "with empty before and after equal",
			value: "=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name := ParseCookieName(tt.value)

			// check that the name is empty
			if name != "" {
				t.Errorf("expected name to be empty, got %s", name)
			}
		})
	}
}
