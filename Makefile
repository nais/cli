.PHONY: check gosec staticcheck vulncheck deadcode vet fmt test build

check: staticcheck vulncheck deadcode gosec

gosec:
	go tool github.com/securego/gosec/v2/cmd/gosec --exclude-generated -terse ./...

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

test:
	go test --race ./...
	go tool github.com/onsi/ginkgo/v2/ginkgo -r --randomize-all --randomize-suites --fail-on-pending --fail-on-empty --keep-going --cover --coverprofile=cover.out --race --trace --junit-report=report.xml --github-output

build:
	go build -installsuffix cgo -o bin/nais main.go
