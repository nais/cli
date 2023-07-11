local:
	go build -o nais ./

test:
	go test ./... -count=1 -coverprofile cover.out -short
