testIgnorePattern := "/interfaces|/constants|/fixtures"

ut:
  go test $(go list ./... | grep -v -E '{{testIgnorePattern}}') -cover -gcflags=all=-l -coverprofile=coverage.out -tags=exclude_fixture

lint:
  golangci-lint run -c .golangci.yaml

fmt:
  go fmt ./...

bench:
  go test -bench=. -benchmem -run='^$' ./...

ci: ut lint fmt
