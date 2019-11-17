package processor

import (
	"context"
	"encoding/json"
	"github.com/analogj/lodestone-processor/pkg/model"
	"io/ioutil"
	"log"
	"os"
)
import "github.com/google/go-tika/tika"

func DocumentProcessor(body []byte, storageEndpoint string) error {

	//make a temporary directory for subsequent processing (original file download, and thumb generation)
	dir, err := ioutil.TempDir("", "doc")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir) // clean up

	var event model.S3Event
	err = json.Unmarshal(body, &event)
	if err != nil {
		return err
	}

	docBucketName, docBucketPath, err := generateStoragePath(event)
	if err != nil {
		return err
	}

	filePath, err := retrieveDocument(storageEndpoint, docBucketName, docBucketPath, dir)
	if err != nil {
		return err
	}

	//TODO pass document to TIKA
	err = parseDocument(filePath)
	if err != nil {
		return err
	}

	return nil
}

func parseDocument(localFilePath string) error {

	f, err := os.Open(localFilePath)
	if err != nil {
		return err
	}

	client := tika.NewClient(nil, "http://tika:9998")
	body, err := client.Parse(context.Background(), f)
	if err != nil {
		return err
	}

	log.Printf("%s", body)

	return nil
}
