#!/bin/bash
just gen
go run github.com/syumai/workers/cmd/workers-assets-gen -mode=go
GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o ./build/app.wasm .
cp -r ./public/* ./build/public/