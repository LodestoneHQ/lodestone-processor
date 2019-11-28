#!/bin/bash -e

# download the packr CLI
go get -u github.com/gobuffalo/packr/packr

go mod vendor
packr build -o lodestone-document-processor-linux-amd64 ./cmd/document-processor/document-processor.go
go build -o lodestone-thumbnail-processor-linux-amd64 ./cmd/thumbnail-processor/thumbnail-processor.go

./lodestone-document-processor-linux-amd64 --help
./lodestone-thumbnail-processor-linux-amd64 --help
