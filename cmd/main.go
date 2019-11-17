package main

import (
	"github.com/analogj/lodestone-processor/pkg/processor"
	"log"
)

func main() {
	log.Printf("%s", processor.ThumbnailProcessor([]byte{}))
}
