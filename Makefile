all: binary

binary:
	rm -f nodetree
	GOARCH=amd64 GOOS=linux godep go build -o nodetree
	GOARCH=amd64 GOOS=darwin godep go build -o nodetree

test:
	GOARCH=amd64 GOOS=linux godep go test -v ./...
	GOARCH=amd64 GOOS=darwin godep go test -v ./...
