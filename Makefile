local:
	go build -o nais ./

test:
	go test ./... -count=1 -coverprofile cover.out -short

check:
	go run honnef.co/go/tools/cmd/staticcheck ./...
	go run golang.org/x/vuln/cmd/govulncheck ./...