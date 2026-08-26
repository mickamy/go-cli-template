// Command init turns a fresh copy of go-cli-template into your own project:
// it rewrites the module path, the binary name, the description, and the
// license holder, strips the template instructions from the README, and
// finally deletes itself.
//
// Run it once from the repository root, after gonew:
//
//	gonew github.com/mickamy/go-cli-template github.com/you/yourtool
//	cd yourtool
//	go run ./tools/init -description "What yourtool does."
package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	templateModule      = "github.com/mickamy/go-cli-template"
	templateName        = "go-cli-template"
	templateAuthor      = "Tetsuro Mikami"
	templateYear        = "2026"
	templateDescription = "A short description of go-cli-template."
)

var readmeTemplateBlock = regexp.MustCompile(`(?s)<!-- template:start -->.*?<!-- template:end -->\n*`)

type options struct {
	module      string
	name        string
	description string
	author      string
	year        string
	dryRun      bool
}

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "init: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	var opts options
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.StringVar(&opts.module, "module", "", "new module path (default: the module in go.mod, as rewritten by gonew)")
	fs.StringVar(&opts.name, "name", "", "binary name (default: last element of the module path)")
	fs.StringVar(&opts.description, "description", "", "one-line description for the README and the release")
	fs.StringVar(&opts.author, "author", "", "license holder (default: git config user.name)")
	fs.StringVar(&opts.year, "year", time.Now().Format("2006"), "license year")
	fs.BoolVar(&opts.dryRun, "dry-run", false, "print what would change without writing")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}

	if _, err := os.Stat("go.mod"); err != nil {
		return errors.New("run from the repository root (go.mod not found)")
	}

	if err := opts.fillDefaults(ctx); err != nil {
		return err
	}

	changed, err := rewriteTree(".", opts)
	if err != nil {
		return err
	}
	for _, f := range changed {
		fmt.Println("rewrote", f)
	}

	if opts.dryRun {
		fmt.Println("dry run: nothing written")

		return nil
	}

	if err := os.RemoveAll("tools"); err != nil {
		return fmt.Errorf("remove tools/: %w", err)
	}
	fmt.Println("removed tools/")
	fmt.Printf("\nNext:\n  git init && git add -A && git commit -m \"chore: initial commit\"\n  go test ./...\n")

	return nil
}

func (o *options) fillDefaults(ctx context.Context) error {
	if o.module == "" {
		m, err := modulePath("go.mod")
		if err != nil {
			return err
		}
		o.module = m
	}
	if o.module == templateModule {
		return errors.New("module path is still the template's; run gonew first or pass -module")
	}
	if o.name == "" {
		o.name = path.Base(o.module)
	}
	if o.author == "" {
		o.author = gitUserName(ctx)
	}
	if o.author == "" {
		return errors.New("-author is required when git config user.name is not set")
	}

	return nil
}

func modulePath(gomod string) (string, error) {
	data, err := os.ReadFile(gomod)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", gomod, err)
	}
	for line := range strings.SplitSeq(string(data), "\n") {
		if m, ok := strings.CutPrefix(strings.TrimSpace(line), "module "); ok {
			return strings.TrimSpace(m), nil
		}
	}

	return "", fmt.Errorf("no module directive in %s", gomod)
}

func gitUserName(ctx context.Context) string {
	out, err := exec.CommandContext(ctx, "git", "config", "user.name").Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(out))
}

func rewriteTree(dir string, opts options) ([]string, error) {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", dir, err)
	}
	defer root.Close()

	var changed []string
	err = fs.WalkDir(root.FS(), ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skipDir(d.Name()) {
				return fs.SkipDir
			}

			return nil
		}

		data, err := root.ReadFile(p)
		if err != nil {
			return fmt.Errorf("read %s: %w", p, err)
		}
		if bytes.IndexByte(data, 0) >= 0 {
			return nil // binary
		}

		out := rewrite(string(data), p, opts)
		if out == string(data) {
			return nil
		}
		changed = append(changed, filepath.Join(dir, filepath.FromSlash(p)))
		if opts.dryRun {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("stat %s: %w", p, err)
		}
		if err := root.WriteFile(p, []byte(out), info.Mode().Perm()); err != nil {
			return fmt.Errorf("write %s: %w", p, err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk %s: %w", dir, err)
	}

	return changed, nil
}

func skipDir(name string) bool {
	switch name {
	case ".git", ".idea", ".vscode", "tools", "node_modules":
		return true
	}

	return false
}

func rewrite(content, rel string, opts options) string {
	if opts.description != "" {
		content = strings.ReplaceAll(content, templateDescription, opts.description)
	}
	content = strings.ReplaceAll(content, templateModule, opts.module)
	content = strings.ReplaceAll(content, templateName, opts.name)

	switch rel {
	case "LICENSE":
		content = strings.ReplaceAll(content, templateYear+" "+templateAuthor, opts.year+" "+opts.author)
	case "README.md":
		content = readmeTemplateBlock.ReplaceAllString(content, "")
	}

	return content
}
