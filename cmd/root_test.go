package cmd

import (
	"os"
	"path"
	"testing"
)

func TestRealPath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home directory: %v", err)
	}

	cases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "tilda",
			input: "~/.bqiam-cache-file.toml",
			want:  path.Join(homeDir, ".bqiam-cache-file.toml"),
		},
		{
			name:  "relative",
			input: ".bqiam-cache-file.toml",
			want:  ".bqiam-cache-file.toml",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := realPath(c.input)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != c.want {
				t.Errorf("got: %s, want: %s", got, c.want)
			}
		})
	}
}
