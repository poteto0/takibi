testIgnorePattern := "/interfaces|/constants|/fixtures"

ut:
  go test $(go list ./... | grep -v -E '{{testIgnorePattern}}') -cover -gcflags=all=-l -coverprofile=coverage.out -tags=exclude_fixture

lint:
  golangci-lint run -c .golangci.yaml

build-wasm:
  GOOS=js GOARCH=wasm go build ./...

lint-wasm:
  GOOS=js GOARCH=wasm GOLANGCI_LINT_CACHE="$HOME/.cache/golangci-lint-wasm" golangci-lint run -c .golangci.yaml

fmt:
  go fmt ./...
  templ fmt .

bench:
  go test -bench=. -benchmem -run='^$' ./...

ci: ut lint lint-wasm build-wasm fmt

gen:
  templ generate

[working-directory("docs-tool")]
gen-code +input:
  go run main.go {{input}}

[working-directory("docs")]
doc:
  @pnpm run dev

[working-directory("docs")]
deploy-doc:
  @pnpm run deploy