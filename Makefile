mac:
	GOOS=darwin GOARCH=amd64 go build -o ./bin/cnvrg-dep-tool-amd64
	GOOS=darwin GOARCH=arm64 go build -o ./bin/cnvrg-dep-tool-arm64
linux:
	GOOS=linux GOARCH=amd64 go build -o ./bin/cnvrg-dep-tool-linux


