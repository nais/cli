
test:
	go run github.com/onsi/ginkgo/v2/ginkgo -r --race --randomize-all --randomize-suites --fail-on-pending --fail-on-empty

test-ci:
	go run github.com/onsi/ginkgo/v2/ginkgo -r --randomize-all --randomize-suites --fail-on-pending --fail-on-empty --keep-going --cover --coverprofile=cover.profile --race --trace --junit-report=report.xml --github-output

nais-cli:
	go build -installsuffix cgo -o bin/nais main.go

check:
	go run honnef.co/go/tools/cmd/staticcheck ./...
	go run golang.org/x/vuln/cmd/govulncheck ./...