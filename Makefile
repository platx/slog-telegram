test:
	go test -v ./...

cover:
	go test -coverprofile=coverage.txt

lint:
	golangci-lint run
