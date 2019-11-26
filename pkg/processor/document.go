package processor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/analogj/lodestone-processor/pkg/model"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/gobuffalo/packr"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"time"
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

type TikaRoundTripper struct {
	r http.RoundTripper
}

// https://cwiki.apache.org/confluence/display/tika/TikaJAXRS#TikaJAXRS-MultipartSupport TIKA must have an Accept header to return JSON responses.
func (mrt TikaRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.URL.Path == "/meta" {
		r.Header.Add("Accept", "application/json")
	}

	return mrt.r.RoundTrip(r)
}

func (dp *DocumentProcessor) tikaHttpClient() *http.Client {
	client := &http.Client{
		Timeout:   time.Second * 10,
		Transport: TikaRoundTripper{r: http.DefaultTransport},
	}

	return client
}

func (dp *DocumentProcessor) parseDocument(bucketPath string, localFilePath string) (model.Document, error) {

	docFile, err := os.Open(localFilePath)
	if err != nil {
		return model.Document{}, err
	}
	defer docFile.Close()

	client := tika.NewClient(dp.tikaHttpClient(), dp.tikaEndpoint.String())
	docContent, err := client.Parse(context.Background(), docFile)
	if err != nil {
		return model.Document{}, err
	}
	log.Printf("docContent: %s", docContent)

	metaFile, err := os.Open(localFilePath)
	if err != nil {
		return model.Document{}, err
	}
	defer metaFile.Close()

	metaJson, err := client.Meta(context.Background(), metaFile)
	if err != nil {
		return model.Document{}, err
	}
	log.Printf("metaJson: %s", metaJson)

	doc := model.Document{
		ID:      bucketPath,
		Content: docContent,
	}

	err = dp.parseTikaMetadata(metaJson, &doc)

	return doc, err
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

func (dp *DocumentProcessor) parseTikaMetadata(metaJson string, doc *model.Document) error {
	var parsedMeta map[string]interface{}
	err := json.Unmarshal([]byte(metaJson), &parsedMeta)
	if err != nil {
		return err
	}

	doc.File.ContentType = dp.findString(parsedMeta, "Content-Type", "content-type")
	doc.Meta = model.DocMeta{
		Author:      dp.findString(parsedMeta, "Author", "meta:author"),
		Date:        dp.findString(parsedMeta, "Date"),
		Keywords:    dp.findStringArray(parsedMeta, "Keywords", "meta:keyword", "pdf:docinfo:keywords"),
		Title:       dp.findString(parsedMeta, "title", "dc:title", "cp:subject", "pdf:docinfo:title"),
		Language:    dp.findString(parsedMeta, "language"),
		Format:      dp.findString(parsedMeta, "dc:format"),
		Identifier:  dp.findString(parsedMeta, "identifier"),
		Contributor: dp.findString(parsedMeta, "contributor"),
		Modifier:    dp.findString(parsedMeta, "modifier"),
		CreatorTool: dp.findString(parsedMeta, "pdf:docinfo:creator_tool", "xmp:CreatorTool"),
		Publisher:   dp.findString(parsedMeta, "publisher"),
		Relation:    dp.findString(parsedMeta, "relation"),
		Rights:      dp.findString(parsedMeta, "rights"),
		Source:      dp.findString(parsedMeta, "source"),
		Type:        dp.findString(parsedMeta, "type"),
		Description: dp.findString(parsedMeta, "description", "subject", "dc:description", "cp:subject", "pdf:docinfo:subject"),
		Created:     dp.findString(parsedMeta, "created", "Creation-Date"),
		//PrintDate    time.Time `json:"print_date"`
		//MetadataDate time.Time `json:"metadata_date"`
		Latitude:  dp.findString(parsedMeta, "latitude", "Latitude"),
		Longitude: dp.findString(parsedMeta, "longitude", "Longitude"),
		Altitude:  dp.findString(parsedMeta, "altitude"),
		//Rating       byte      `json:"rating"`
		Comments: dp.findString(parsedMeta, "comments"),
	}
	return nil
}

func (dp *DocumentProcessor) findStringArray(dict map[string]interface{}, keys ...string) []string {
	for _, v := range keys {
		val := castToStringArray(dict[v])
		if val != nil {
			return val
		}
	}

	return []string{}
}

func (dp *DocumentProcessor) findString(dict map[string]interface{}, keys ...string) string {
	for _, v := range keys {
		val := castToString(dict[v])
		if val != "" {
			return val
		}
	}

	return ""
}

func castToString(val interface{}) string {
	rt := reflect.TypeOf(val)
	switch rt.Kind() {
	case reflect.Slice:
		return fmt.Sprintf(strings.Join(val.([]string), ", "))
	case reflect.Array:
		return fmt.Sprintf(strings.Join(val.([]string), ", "))
	case reflect.String:
		return val.(string)
	default:
		return ""
	}
}

func castToStringArray(val interface{}) []string {
	rt := reflect.TypeOf(val)
	switch rt.Kind() {
	case reflect.Slice:
		return val.([]string)
	case reflect.Array:
		return val.([]string)
	case reflect.String:
		return []string{val.(string)}
	default:
		return nil
	}
}

//func stringHasValue(str string) bool {
//	return str != ""
//}
