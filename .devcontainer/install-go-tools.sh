#!/usr/bin/env bash

set -euo pipefail

if ! command -v go >/dev/null 2>&1; then
	echo "Go is required to install tooling." >&2
	exit 1
fi

export GOBIN="${GOBIN:-$(go env GOPATH)/bin}"
mkdir -p "$GOBIN"

curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b "$GOBIN" v2.10.1
go install golang.org/x/vuln/cmd/govulncheck@latest
go install github.com/securego/gosec/v2/cmd/gosec@latest
go install github.com/fe3dback/go-arch-lint@latest
go install github.com/go-gremlins/gremlins/cmd/gremlins@v0.6.0
