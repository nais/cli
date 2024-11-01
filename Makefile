.PHONY: build test test-ci nais-cli check fmt vet

build:
	go build

test: fmt vet
	go run github.com/onsi/ginkgo/v2/ginkgo -r --race --randomize-all --randomize-suites --fail-on-pending --fail-on-empty

fmt:
	go run mvdan.cc/gofumpt -w ./
vet:
	go vet ./...

test-ci:
	go run github.com/onsi/ginkgo/v2/ginkgo -r --randomize-all --randomize-suites --fail-on-pending --fail-on-empty --keep-going --cover --coverprofile=cover.out --race --trace --junit-report=report.xml --github-output

nais-cli:
	go build -installsuffix cgo -o bin/nais main.go

check:
	go run honnef.co/go/tools/cmd/staticcheck ./...
	go run golang.org/x/vuln/cmd/govulncheck ./...
