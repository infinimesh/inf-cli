# infinimesh CLI

Official CLI for [infinimesh Platform](https://github.com/infinimesh/core)

## Table of Contents

* [Usage](#usage)
* [Installation](#installation)
* [Notes](#notes)

## Usage

Check out `inf -h` for list of available commands.
You can also use `inf help [command]` to see help notes for particular command.

You must authorize first, to use other commands, see `inf help login`.

## Installation

### Homebrew (macOS and Linux)

```shell
brew tap infinimesh/inf-cli https://github.com/infinimesh/inf-cli
brew install inf
```

If brew says the tap is untrusted, run `brew trust infinimesh/inf-cli` once and install again.

### Prebuilt binaries

Download the archive for your system from the [releases page](https://github.com/infinimesh/inf-cli/releases), unpack it and put `inf` somewhere on your `PATH`.

### Go install

```shell
go install github.com/infinimesh/inf-cli@latest
```

This installs a binary named `inf-cli`. Rename it to `inf` so the commands in the docs work.

### Build from source

```shell
git clone https://github.com/infinimesh/inf-cli
cd inf-cli
go build -o inf .
```

## Checks

`make commit-check` runs `fmt`, `race`, `test-fuzz`, and `code-ql-gate`. `fmt` writes Go files that `gofmt` would change. `code-ql` runs the CodeQL CLI and writes SARIF under `.tmp/codeql/`. The gate fails when a result is level `error`, ignoring paths under `vendor`. `jq` is required. GitHub Actions runs the same race and fuzz steps, and CodeQL on pull requests.

## Notes

### Login

> CLI currently supports only `standard` type of Authorization

### Auto-Completion

> CLI is based on [spf13/cobra](https://github.com/spf13/cobra) and supports auto-completion,
> you can see register scripts by running `inf completion [your shell]`
> To check how to register completions, add `-h` flag
