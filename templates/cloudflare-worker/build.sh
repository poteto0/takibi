#!/bin/bash
set -e
go run github.com/syumai/workers/cmd/workers-assets-gen -mode=go
GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o ./build/app.wasm .
