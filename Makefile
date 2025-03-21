.PHONY: build check staticcheck vulncheck deadcode fmt test test-ci nais-cli vet

build: check fmt
	go build

check: staticcheck vulncheck deadcode

staticcheck:
	go run honnef.co/go/tools/cmd/staticcheck@latest ./...

vulncheck:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

deadcode:
	go run golang.org/x/tools/cmd/deadcode@latest -test ./...

vet:
	go vet ./...

fmt:
	go run mvdan.cc/gofumpt@latest -w ./

test: fmt vet
	go run github.com/onsi/ginkgo/v2/ginkgo -r --race --randomize-all --randomize-suites --fail-on-pending --fail-on-empty

test-ci: vet
	go run github.com/onsi/ginkgo/v2/ginkgo -r --randomize-all --randomize-suites --fail-on-pending --fail-on-empty --keep-going --cover --coverprofile=cover.out --race --trace --junit-report=report.xml --github-output

nais-cli:
	go build -installsuffix cgo -o bin/nais main.go
