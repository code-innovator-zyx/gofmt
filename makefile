windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build

ubuntu:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build