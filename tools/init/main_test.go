package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestRewrite(t *testing.T) {
	t.Parallel()

	opts := options{
		module:      "github.com/you/yourtool",
		name:        "yourtool",
		description: "Does the thing.",
		author:      "You",
		year:        "2030",
	}

	tests := map[string]struct {
		rel     string
		content string
		want    string
	}{
		"go import path": {
			rel:     "main.go",
			content: `import "github.com/mickamy/go-cli-template/internal/cli"`,
			want:    `import "github.com/you/yourtool/internal/cli"`,
		},
		"binary name": {
			rel:     "internal/cli/cli.go",
			content: `const Name = "go-cli-template"`,
			want:    `const Name = "yourtool"`,
		},
		"description before name": {
			rel:     "README.md",
			content: "# go-cli-template\n\nA short description of go-cli-template.\n",
			want:    "# yourtool\n\nDoes the thing.\n",
		},
		"license holder": {
			rel:     "LICENSE",
			content: "Copyright (c) 2026 Tetsuro Mikami\n",
			want:    "Copyright (c) 2030 You\n",
		},
		"author elsewhere is untouched": {
			rel:     "README.md",
			content: "Maintained by Tetsuro Mikami.\n",
			want:    "Maintained by Tetsuro Mikami.\n",
		},
		"readme template block": {
			rel: "README.md",
			content: "# go-cli-template\n\n" +
				"<!-- template:start -->\n## Using this template\n\nsteps\n<!-- template:end -->\n\n" +
				"## Installation\n",
			want: "# yourtool\n\n## Installation\n",
		},
		"untouched": {
			rel:     "Makefile",
			content: "test:\n\tgo test ./...\n",
			want:    "test:\n\tgo test ./...\n",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := rewrite(tt.content, tt.rel, opts); got != tt.want {
				t.Errorf("rewrite:\n got %q\nwant %q", got, tt.want)
			}
		})
	}
}

func TestRewriteKeepsDescriptionPlaceholderWhenEmpty(t *testing.T) {
	t.Parallel()

	opts := options{module: "github.com/you/yourtool", name: "yourtool", author: "You", year: "2030"}
	got := rewrite("A short description of go-cli-template.\n", "README.md", opts)
	if want := "A short description of yourtool.\n"; got != want {
		t.Errorf("rewrite: got %q, want %q", got, want)
	}
}

func TestRewriteTree(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	write := func(rel, content string) {
		t.Helper()
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("go.mod", "module github.com/you/yourtool\n")
	write("main.go", `import "github.com/mickamy/go-cli-template/internal/cli"`)
	write("tools/init/main.go", "const templateName = \"go-cli-template\"\n")
	write(".git/config", "url = github.com/mickamy/go-cli-template\n")
	write("bin/blob", "go-cli-template\x00binary")
	write("Makefile", "build:\n\tgo build -o bin/go-cli-template .\n")

	opts := options{module: "github.com/you/yourtool", name: "yourtool", author: "You", year: "2030"}
	changed, err := rewriteTree(root, opts)
	if err != nil {
		t.Fatal(err)
	}

	rel := make([]string, 0, len(changed))
	for _, c := range changed {
		r, err := filepath.Rel(root, c)
		if err != nil {
			t.Fatal(err)
		}
		rel = append(rel, filepath.ToSlash(r))
	}
	slices.Sort(rel)
	if want := []string{"Makefile", "main.go"}; !slices.Equal(rel, want) {
		t.Errorf("changed files: got %v, want %v", rel, want)
	}

	got, err := os.ReadFile(filepath.Join(root, "Makefile"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "bin/yourtool") {
		t.Errorf("Makefile not rewritten: %q", got)
	}
	untouched, err := os.ReadFile(filepath.Join(root, "tools/init/main.go"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(untouched), "yourtool") {
		t.Errorf("tools/ must be skipped, got %q", untouched)
	}
}

func TestModulePath(t *testing.T) {
	t.Parallel()

	p := filepath.Join(t.TempDir(), "go.mod")
	if err := os.WriteFile(p, []byte("module github.com/you/yourtool\n\ngo 1.25.0\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := modulePath(p)
	if err != nil {
		t.Fatal(err)
	}
	if got != "github.com/you/yourtool" {
		t.Errorf("modulePath: got %q", got)
	}
}
