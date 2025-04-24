.PHONY: build check staticcheck vulncheck deadcode fmt test test-ci nais-cli vet

build: check fmt
	go build

check: staticcheck vulncheck deadcode

staticcheck:
	go tool honnef.co/go/tools/cmd/staticcheck ./...

vulncheck:
	go tool golang.org/x/vuln/cmd/govulncheck ./...

deadcode:
	go tool golang.org/x/tools/cmd/deadcode -test ./...

vet:
	go vet ./...

fmt:
	go tool mvdan.cc/gofumpt -w ./

test: fmt vet
	go tool github.com/onsi/ginkgo/v2/ginkgo -r --race --randomize-all --randomize-suites --fail-on-pending --fail-on-empty

test-ci: vet
	go tool github.com/onsi/ginkgo/v2/ginkgo -r --randomize-all --randomize-suites --fail-on-pending --fail-on-empty --keep-going --cover --coverprofile=cover.out --race --trace --junit-report=report.xml --github-output

nais-cli:
	go build -installsuffix cgo -o bin/nais main.go
