package document

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/analogj/lodestone-processor/pkg/model"
	"github.com/analogj/lodestone-processor/pkg/processor"
	"github.com/analogj/lodestone-processor/pkg/processor/api"
	"github.com/analogj/lodestone-processor/pkg/version"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/gobuffalo/packr"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)
import "github.com/google/go-tika/tika"

type DocumentProcessor struct {
	processor.CommonProcessor

	apiEndpoint                  *url.URL
	storageThumbnailBucket       string
	tikaEndpoint                 *url.URL
	elasticsearchEndpoint        *url.URL
	elasticsearchIndex           string
	elasticsearchMappingOverride string
	elasticsearchClient          *elasticsearch.Client
	mappings                     *packr.Box
	filter                       *model.Filter
	logger                       *logrus.Entry
}

func CreateDocumentProcessor(logger *logrus.Entry, apiEndpoint string, storageThumbnailBucket string, tikaEndpoint string, elasticsearchEndpoint string, elasticsearchIndex string, elasticsearchMapping string) (DocumentProcessor, error) {

	apiEndpointUrl, err := url.Parse(apiEndpoint)
	if err != nil {
		return DocumentProcessor{}, err
	}

	//retrieve the filters (include/excludes) from the API
	filterData, err := api.GetIncludeExcludeData(apiEndpointUrl)
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

	box := packr.NewBox("../../../static/document-processor")

	dp := DocumentProcessor{
		apiEndpoint:                  apiEndpointUrl,
		storageThumbnailBucket:       storageThumbnailBucket,
		tikaEndpoint:                 tikaEndpointUrl,
		elasticsearchEndpoint:        elasticsearchEndpointUrl,
		elasticsearchIndex:           elasticsearchIndex,
		elasticsearchMappingOverride: elasticsearchMapping,
		mappings:                     &box,
		filter:                       &filterData,
		logger:                       logger,
	}

	//ensure the elastic search index exists (do this once on startup)
	cfg := elasticsearch.Config{
		Addresses: []string{dp.elasticsearchEndpoint.String()},
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return DocumentProcessor{}, err
	}
	dp.elasticsearchClient = es

	dp.logger.Debugln("Connect to ElasticSearch & ensure indicies exist")
	dp.logger.Debugln(elasticsearch.Version)
	dp.logger.Debugln(es.Info())

	err = dp.ensureIndicies()
	if err != nil {
		return DocumentProcessor{}, err
	}

	return dp, nil
}

func (dp *DocumentProcessor) Process(body []byte) error {

	var event model.S3Event
	err := json.Unmarshal(body, &event)
	if err != nil {
		return err
	}

	docBucketName, docBucketPath, err := api.GenerateStoragePath(event)
	if err != nil {
		return err
	}

	//determine if we should even be processing this document
	includeDocument := dp.filter.ValidPath(docBucketPath)
	if !includeDocument {
		dp.logger.Infof("Ignoring document, matches exclude pattern (%s, %s)", docBucketName, docBucketPath)
		return nil
	}

	//make a temporary directory for subsequent processing (original file download, and thumb generation)
	dir, err := ioutil.TempDir("", "doc")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir) // clean up

	if event.Records[0].EventName == "s3:ObjectRemoved:Delete" {
		dp.logger.Debugln("Attempting to delete file")

		//delete document in Elasticsearch
		err = dp.deleteDocument(docBucketName, docBucketPath)
		if err != nil {
			return err
		}
		return nil
	} else {

		filePath, err := api.ReadFile(dp.apiEndpoint, docBucketName, docBucketPath, dir)
		if err != nil {
			return err
		}
		if dp.IsEmptyFile(filePath) {
			dp.logger.Infof("Ignoring document, filesize is 0 (%s, %s)", docBucketName, docBucketPath)
			return nil
		}

		//pass document to TIKA
		doc, err := dp.parseDocument(docBucketName, docBucketPath, filePath)
		if err != nil {
			return err
		}

		//store document in Elasticsearch
		err = dp.storeDocument(doc)
		if err != nil {
			return err
		}
	}

	return nil
}
func (dp *DocumentProcessor) tikaHttpClient() *http.Client {
	client := &http.Client{
		Timeout:   time.Second * 10,
		Transport: TikaRoundTripper{r: http.DefaultTransport},
	}

	return client
}

func (dp *DocumentProcessor) parseDocument(bucketName string, bucketPath string, localFilePath string) (model.Document, error) {

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
	//trim whitespace/newline characters
	docContent = strings.TrimSpace(docContent)
	dp.logger.Debugf("docContent: '%s'", docContent)

	metaFile, err := os.Open(localFilePath)
	if err != nil {
		return model.Document{}, err
	}
	defer metaFile.Close()

	metaJson, err := client.Meta(context.Background(), metaFile)
	if err != nil {
		return model.Document{}, err
	}
	dp.logger.Debugf("metaJson: %s", metaJson)

	fileStat, err := os.Stat(localFilePath)
	if err != nil {
		return model.Document{}, err
	}

	shaFile, err := os.Open(localFilePath)
	if err != nil {
		return model.Document{}, err
	}
	defer shaFile.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, shaFile); err != nil {
		return model.Document{}, err
	}
	sha256Checksum := hex.EncodeToString(hasher.Sum(nil))

	sysStat := fileStat.Sys().(*syscall.Stat_t)
	AccessedTime := time.Unix(int64(sysStat.Atim.Sec), int64(sysStat.Atim.Nsec))
	CreatedTime := time.Unix(int64(sysStat.Ctim.Sec), int64(sysStat.Ctim.Nsec))

	grp, err := user.LookupGroupId(fmt.Sprintf("%d", sysStat.Gid))
	grpName := ""
	if err == nil {
		grpName = grp.Name
	}

	usr, err := user.LookupGroupId(fmt.Sprintf("%d", sysStat.Uid))
	usrName := ""
	if err == nil {
		usrName = usr.Name
	}

	//convert the filepath into "tags"
	bucketPathDir, _ := filepath.Split(bucketPath)
	bucketPathTags := strings.Split(bucketPathDir, "/")

	doc := model.Document{
		//ID length limit is 512 bytes, cant use path or base64 here. instead we'll use the document checksum value.
		ID:      sha256Checksum,
		Content: docContent, //make sure that empty content is stored as ""
		Lodestone: model.DocLodestone{
			ProcessorVersion: version.VERSION,
			Title:            "",
			Tags:             deleteEmpty(bucketPathTags),
			Bookmark:         false,
		},
		File: model.DocFile{
			FileName:     fileStat.Name(),
			Extension:    strings.ToLower(strings.TrimPrefix(path.Ext(fileStat.Name()), ".")),
			Filesize:     fileStat.Size(),
			IndexedChars: int64(len(docContent)),
			IndexedDate:  time.Now(),

			Created:      CreatedTime,
			LastModified: fileStat.ModTime(),
			LastAccessed: AccessedTime,
			Checksum:     sha256Checksum,

			Group: grpName,
			Owner: usrName,
		},
		Storage: model.DocStorage{
			Path:        bucketPath,
			Bucket:      bucketName,
			ThumbBucket: dp.storageThumbnailBucket,
			ThumbPath:   api.GenerateThumbnailStoragePath(bucketPath),
		},
	}

	err = dp.parseTikaMetadata(metaJson, &doc)
	return doc, err
}

//store document in elasticsearch
func (dp *DocumentProcessor) storeDocument(document model.Document) error {
	// use https://github.com/elastic/go-elasticsearch
	// first migrate capsulecd to support mod

	payload, err := json.Marshal(document)
	if err != nil {
		dp.logger.Printf("An error occurred while json encoding Document: %v", err)
		return err
	}

	dp.logger.Println("Attempting to store new document in elasticsearch")
	esResp, err := dp.elasticsearchClient.Index(dp.elasticsearchIndex, bytes.NewReader(payload), dp.elasticsearchClient.Index.WithDocumentID(document.ID))
	dp.logger.Debugf("DEBUG: ES response: %v", esResp)
	if err != nil {
		dp.logger.Printf("An error occurred while storing document: %v", err)
	}

	return err
}

//deletee document in elasticsearch
func (dp *DocumentProcessor) deleteDocument(docBucketName string, docBucketPath string) error {
	// use https://github.com/elastic/go-elasticsearch

	dp.logger.Println("Attempting to delete document by query in elasticsearch")
	esResp, err := dp.elasticsearchClient.DeleteByQuery([]string{dp.elasticsearchIndex}, strings.NewReader(
		fmt.Sprintf(`
			{
				"query": {
					"bool": {
						"must": [
							{ "match": { "storage.bucket": "%v" }},
							{ "match": { "storage.path": "%v"}}
						]
					}
				}
			}`, docBucketName, docBucketPath)))
	dp.logger.Debugf("DEBUG: ES response: %v", esResp)
	if err != nil {
		dp.logger.Printf("An error occurred while deleting document: %v", err)
	}
	return err
}

func (dp *DocumentProcessor) ensureIndicies() error {
	//we cant be sure that the elasticsearch index (& mappings) already exist, so we have to check if they exist on every document insertion.

	dp.logger.Printf("Attempting to create %s index, if it does not exist", dp.elasticsearchIndex)
	resp, err := dp.elasticsearchClient.Indices.Exists([]string{dp.elasticsearchIndex})
	dp.logger.Debugf("%v \n %v", resp, err)
	if err == nil && resp.StatusCode == 200 {
		//index exists, do nothing
		dp.logger.Println("Index already exists, skipping.")
		return nil
	}

	//index does not exist, lets create it
	var mappingReader io.Reader
	if len(dp.elasticsearchMappingOverride) > 0 {
		f, err := os.Open(dp.elasticsearchMappingOverride)
		if err != nil {
			dp.logger.Printf("COULD NOT OPEN MAPPING OVERRIDE FILE: %v, %v", dp.elasticsearchMappingOverride, err)
			return err
		}
		mappingReader = bufio.NewReader(f)
	} else {
		dp.logger.Debugf("looking for settings.json in mappings....")
		mappings, err := dp.mappings.FindString("settings.json")
		if err != nil {
			dp.logger.Printf("COULD NOT FIND MAPPING FOR settings.json: %v, %v", mappings, err)
			return err
		}
		dp.logger.Debugf("Found settings.json: %v", mappings)
		mappingReader = strings.NewReader(mappings)
	}

	_, err = dp.elasticsearchClient.Indices.Create(dp.elasticsearchIndex, dp.elasticsearchClient.Indices.Create.WithBody(mappingReader))
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
		Date:        dp.findTime(parsedMeta, "Date"),
		CreatedDate: dp.findTime(parsedMeta, "created", "Creation-Date", "meta:creation-date", "dcterms:created", "pdf:docinfo:created"),
		SavedDate:   dp.findTime(parsedMeta, "Last-Save-Date", "meta:save-date"),

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
		Latitude:    dp.findString(parsedMeta, "latitude", "Latitude"),
		Longitude:   dp.findString(parsedMeta, "longitude", "Longitude"),
		Altitude:    dp.findString(parsedMeta, "altitude"),
		//Rating       byte      `json:"rating"`
		Comments: dp.findString(parsedMeta, "comments"),
		Pages:    dp.findString(parsedMeta, "xmpTPg:NPages"),
	}
	return nil
}

func (dp *DocumentProcessor) findTime(dict map[string]interface{}, keys ...string) time.Time {
	for _, v := range keys {
		val := castToTime(dict[v])
		if !val.IsZero() {
			return val
		}
	}

	return time.Time{}
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
