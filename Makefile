.PHONY: build
build: check fmt
	go build

.PHONY: check
check: staticcheck vulncheck deadcode

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

.PHONY: fmt
fmt:
	go tool mvdan.cc/gofumpt -w ./

.PHONY: test
test: fmt vet
	go tool github.com/onsi/ginkgo/v2/ginkgo -r --race --randomize-all --randomize-suites --fail-on-pending --fail-on-empty

.PHONY: test-ci
test-ci: vet
	go tool github.com/onsi/ginkgo/v2/ginkgo -r --randomize-all --randomize-suites --fail-on-pending --fail-on-empty --keep-going --cover --coverprofile=cover.out --race --trace --junit-report=report.xml --github-output

.PHONY: nais-cli
nais-cli:
	go build -installsuffix cgo -o bin/nais main.go
