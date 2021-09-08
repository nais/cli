local:
	go build -o tool/debuk main/debuk/*.go
debuk:
	go install main/debuk/debuk.go
test:
	go test ./... -count=1 -coverprofile cover.out -short