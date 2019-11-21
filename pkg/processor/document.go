package processor

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/analogj/lodestone-processor/pkg/model"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/gobuffalo/packr"
	"io/ioutil"
	"log"
	"net/url"
	"os"
)
import "github.com/google/go-tika/tika"

type DocumentProcessor struct {
	storageEndpoint       *url.URL
	tikaEndpoint          *url.URL
	elasticsearchEndpoint *url.URL
	elasticsearchIndex    string
	mappings              packr.Box
}

func CreateDocumentProcessor(storageEndpoint string, tikaEndpoint string, elasticsearchEndpoint string, elasticsearchIndex string) (DocumentProcessor, error) {

	storageEndpointUrl, err := url.Parse(storageEndpoint)
	if err != nil {
		return DocumentProcessor{}, err
	}

	tikaEndpointUrl, err := url.Parse(tikaEndpoint)
	if err != nil {
		return DocumentProcessor{}, err
	}

	elasticsearchEndpointUrl, err := url.Parse(elasticsearchEndpoint)
	if err != nil {
		return DocumentProcessor{}, err
	}

	box := packr.NewBox("../static/document-processor")

	dp := DocumentProcessor{
		storageEndpoint:       storageEndpointUrl,
		tikaEndpoint:          tikaEndpointUrl,
		elasticsearchEndpoint: elasticsearchEndpointUrl,
		elasticsearchIndex:    elasticsearchIndex,
		mappings:              box,
	}

	return dp, nil
}

func (dp *DocumentProcessor) Process(body []byte) error {

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

	filePath, err := retrieveDocument(dp.storageEndpoint, docBucketName, docBucketPath, dir)
	if err != nil {
		return err
	}

	//TODO pass document to TIKA
	doc, err := dp.parseDocument(docBucketPath, filePath)
	if err != nil {
		return err
	}

	err = dp.storeDocument(docBucketPath, doc)
	if err != nil {
		return err
	}

	return nil
}

func (dp *DocumentProcessor) parseDocument(bucketPath string, localFilePath string) (model.Document, error) {

	f, err := os.Open(localFilePath)
	if err != nil {
		return model.Document{}, err
	}

	client := tika.NewClient(nil, dp.tikaEndpoint.String())
	body, err := client.Parse(context.Background(), f)
	if err != nil {
		return model.Document{}, err
	}

	meta, err := client.Meta(context.Background(), f)
	if err != nil {
		return model.Document{}, err
	}

	log.Printf("body: %s", body)
	log.Printf("meta: %s", meta)

	doc := model.Document{
		ID:      bucketPath,
		Content: body,
	}

	return doc, nil
}

//store document in elasticsearch
func (dp *DocumentProcessor) storeDocument(docBucketPath string, document model.Document) error {
	// use https://github.com/elastic/go-elasticsearch
	// first migrate capsulecd to support mod

	cfg := elasticsearch.Config{
		Addresses: []string{dp.elasticsearchEndpoint.String()},
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return err
	}

	log.Println(elasticsearch.Version)
	log.Println(es.Info())

	err = dp.ensureIndicies(es)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(document)
	if err != nil {
		return err
	}

	_, err = es.Create(dp.elasticsearchIndex, docBucketPath, bytes.NewReader(payload))
	return err
}

func (dp *DocumentProcessor) ensureIndicies(es *elasticsearch.Client) error {
	//we cant be sure that the elasticsearch index (& mappings) already exist, so we have to check if they exist on every document insertion.

	_, err := es.Indices.Exists([]string{dp.elasticsearchIndex})
	if err == nil {
		//index exists, do nothing

		return nil
	}

	//index does not exist, lets create it
	mappings, err := dp.mappings.Find("settings.json")
	if err != nil {
		return err
	}
	mappingsReader := bytes.NewReader(mappings)

	_, err = es.Indices.Create(dp.elasticsearchIndex, es.Indices.Create.WithBody(mappingsReader))
	return err
}
