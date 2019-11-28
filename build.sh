go mod vendor
go build -o lodestone-document-processor-linux-amd64 ./cmd/document-processor/document-processor.go
go build -o lodestone-thumbnail-processor-linux-amd64 ./cmd/thumbnail-processor/thumbnail-processor.go

./lodestone-document-processor-linux-amd64 --help
./lodestone-thumbnail-processor-linux-amd64 --help
