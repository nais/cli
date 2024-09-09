
test:
	go test ./... -count=1 -coverprofile cover.out -short

nais-cli:
	go build -installsuffix cgo -o bin/nais main.go

check:
	go run honnef.co/go/tools/cmd/staticcheck ./...
	go run golang.org/x/vuln/cmd/govulncheck ./...