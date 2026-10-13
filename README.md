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

`inf` 3.4 talks to core 3.4. Rules, alarms, audit, plugins, namespace OTA, and quotas need core `v3.4.0` or newer. The 3.2 CLI cannot log in to a 3.4 core. Each core release also releases `inf` with the same version, and a final release updates the Homebrew formula. Older cores still serve devices, accounts, and namespaces.

A quota of zero means unlimited. Set one with `inf ns create` or `inf ns update` (`--max-devices`, `--device-mps`, `--namespace-mps`).

## Login

```shell
inf login api.localhost:443 --insecure-skip-verify
inf login --root api.localhost:443 --insecure-skip-verify
```

The API on Compose is `api.localhost:443`. A host with no port is dialed on 443. The default Traefik certificate does not cover `api.localhost`, so a local login passes `--insecure-skip-verify`. `--ca-file` is the PEM of CAs to trust when the certificate matches the host. Login stores that file as an absolute path. A later `--ca-file` login checks the certificate again. `--insecure` is plain HTTP/2 and is a different switch. Login stores the host, the token, and those TLS settings.

`--root` asks for a root token. `inf plugins create` and `inf plugins delete` need one. A login without `--root` is an account token. `inf accounts list` with a root token lists every account.

`inf plugins create` prints the addon bearer once and does not store it. Pass that bearer on the next command with `--token`. `--token` does not replace the stored login.

```shell
inf --token "$ADDON" ota assign --device <device> --version 1.2.3
```

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

Each core process serves Prometheus on `METRICS_ADDR` (default `:2112`) inside that process. Compose publishes host port 2112 on the mqtt bridge only, so `inf metrics` requires `--addr`. `infinimesh_up` is 1 while that process is serving `/metrics`.

`inf ota assign` needs an account token or an addon token, plus `--version` and one of `--namespace`, `--tag`, or `--device`. A device token is rejected. A refused device is still a successful call. The error is that device's result.

A full walkthrough, including the OTA addon order (root login, `inf plugins create --kind ota`, then `inf ns update --ota --provider`), is in the [CLI guide](https://github.com/infinimesh/docs/blob/main/guides/cli-getting-started.md).

## Notes

### Login

> CLI currently supports only `standard` type of Authorization

### Auto-Completion

> CLI is based on [spf13/cobra](https://github.com/spf13/cobra) and supports auto-completion,
> you can see register scripts by running `inf completion [your shell]`
> To check how to register completions, add `-h` flag
