.PHONY: all
all: fmt check test test-ci build

.PHONY: build
build:
	go build -installsuffix cgo -o bin/nais ./cmd/cli

.PHONY: check
check: staticcheck vulncheck deadcode vet

.PHONY: staticcheck
staticcheck:
	go tool honnef.co/go/tools/cmd/staticcheck ./...

.PHONY: vulncheck
vulncheck:
	go tool golang.org/x/vuln/cmd/govulncheck ./...

.PHONY: deadcode
deadcode:
	go tool golang.org/x/tools/cmd/deadcode -test ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: gosec
gosec:
	go tool github.com/securego/gosec/v2/cmd/gosec --exclude-generated -terse ./...

.PHONY: fmt
fmt:
	go tool mvdan.cc/gofumpt -w ./

.PHONY: test
test:
	go tool github.com/onsi/ginkgo/v2/ginkgo -r --race --randomize-all --randomize-suites --fail-on-pending --fail-on-empty

.PHONY: test-ci
test-ci:
	go tool github.com/onsi/ginkgo/v2/ginkgo -r --randomize-all --randomize-suites --fail-on-pending --fail-on-empty --keep-going --cover --coverprofile=cover.out --race --trace --junit-report=report.xml --github-output
