#!/bin/bash
GOOS=darwin GOARCH=amd64 go build -o ./bin/cnvrg-dep-tool-amd64
GOOS=darwin GOARCH=arm64 go build -o ./bin/cnvrg-dep-tool-arm64


