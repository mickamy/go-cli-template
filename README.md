# go-cli-template

A short description of go-cli-template.

[![CI](https://github.com/mickamy/go-cli-template/actions/workflows/ci.yaml/badge.svg)](https://github.com/mickamy/go-cli-template/actions/workflows/ci.yaml)
[![Release](https://img.shields.io/github/v/release/mickamy/go-cli-template)](https://github.com/mickamy/go-cli-template/releases/latest)

<!-- template:start -->
## Using this template

This repository is a starting point for a Go command-line tool. It has no dependencies beyond the standard library and comes with:

- A `main.go` that handles Ctrl-C and SIGTERM through `context`, and an `internal/cli` package with subcommand dispatch, per-command flag sets, usage output, and exit codes (`internal/exit`), all testable through injected streams.
- A `version` command backed by `debug.ReadBuildInfo`, so `go install` and tagged builds report the right version without linker flags.
- `golangci-lint` configured with every linter on, minus a documented list of exceptions.
- CI on Linux, macOS, and Windows against the Go version in `go.mod` and the latest stable release, with actions pinned by commit.
- Releases through goreleaser on `v*` tags: archives for six OS/arch pairs, checksums, a grouped changelog, and a Homebrew cask pushed to a tap.
- Dependabot for Go modules and GitHub Actions, an `.editorconfig`, and an MIT license.

### Create a project

```sh
go install golang.org/x/tools/cmd/gonew@latest
gonew github.com/mickamy/go-cli-template github.com/you/yourtool
cd yourtool
go run ./tools/init -description "What yourtool does."
```

`gonew` copies the template and rewrites the module path. `tools/init` then renames the binary, sets the description and the license holder, removes this section from the README, and deletes itself. Run `go run ./tools/init -h` for the flags; `-dry-run` shows what would change.

Then:

```sh
git init && git add -A && git commit -m "chore: initial commit"
gh repo create you/yourtool --public --source . --push
```

### Make it yours

- Replace the `hello` command in `internal/cli/hello.go` with your own, and register it in the `commands` table in `internal/cli/cli.go`.
- Update the sections below.
- To publish to Homebrew, add a `TAP_GITHUB_TOKEN` repository secret: a token with write access to the tap repository named in `.goreleaser.yaml` (`homebrew_casks.repository`). Without it, the release job fails at the Homebrew step; remove the `homebrew_casks` section if you do not need it.
- Cut a release by pushing a tag: `git tag -a v0.1.0 -m v0.1.0 && git push origin v0.1.0`.
<!-- template:end -->

## Installation

```sh
go install github.com/mickamy/go-cli-template@latest
```

With Homebrew:

```sh
brew install --cask mickamy/tap/go-cli-template
```

Prebuilt binaries are on the [releases page](https://github.com/mickamy/go-cli-template/releases).

## Usage

```sh
go-cli-template hello -name Go
go-cli-template help
go-cli-template version
```

## Development

```sh
make test   # go test -race ./...
make lint   # golangci-lint run
make build  # bin/go-cli-template
```

## License

[MIT](./LICENSE)
