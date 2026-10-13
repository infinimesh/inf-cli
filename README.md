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

Download the archive for your system from the [releases page](https://github.com/infinimesh/inf-cli/releases), unpack it and put `inf` somewhere on your `PATH`. A `vX.Y.Z-rc.N` tag is a GitHub pre-release and does not update the Homebrew formula.

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

## Versions

`inf` 3.4 talks to core 3.4. Rules, alarms, audit, plugins, namespace OTA, and quotas need a core built from `v3.4.0-rc.9` or newer. Older cores still serve devices, accounts, and namespaces.

A quota of zero means unlimited.

## Login

```shell
inf login api.localhost:8000 --ca-file ./ca.pem
inf login api.localhost:8000 --insecure-skip-verify
```

`--ca-file` is the PEM of CAs to trust. `--insecure-skip-verify` accepts any certificate. `--insecure` is plain HTTP/2 and is a different switch. Login stores the host, the token, and those TLS settings.

Compose with `BASE_DOMAIN=localhost` uses `console.localhost`, `api.localhost`, and `media.localhost`. Each host has its own certificate. The CLI trusts `--ca-file` for both the API and `inf media`.

`inf login` prints a confirmation. Add `--print-token` to show the bearer.

## Device template

`inf devices create` reads a JSON or YAML file and `--crt`.

```yaml
title: pump
enabled: true
tags:
  - plant
```

```shell
inf devices create pump.yaml --crt pump.pem --namespace <namespace>
```

## Media and metrics

```shell
inf media --url https://media.localhost list <namespace>
inf metrics --addr 127.0.0.1:2112
```

`inf media` defaults to `https://media.localhost`. List, upload, and delete use the stored account token. A downloaded file is public.

Core serves Prometheus on `METRICS_ADDR` (default `:2112`) inside the process. Compose does not publish that port, so pass `--addr`.

`inf ota assign` needs an account token or an addon token. A device token is rejected. `inf plugins create` needs a root token (`POST /v1/token` with `"inf": true`) and prints the addon bearer once.

## Notes

### Login

> CLI currently supports only `standard` type of Authorization

### Auto-Completion

> CLI is based on [spf13/cobra](https://github.com/spf13/cobra) and supports auto-completion,
> you can see register scripts by running `inf completion [your shell]`
> To check how to register completions, add `-h` flag
