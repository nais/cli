local:
	go build -o nais ./cmd

test:
	go test ./... -count=1 -coverprofile cover.out -short
