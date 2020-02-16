package thumbnail

import (
	"encoding/json"
	"fmt"
	"github.com/analogj/lodestone-processor/pkg/model"
	"github.com/analogj/lodestone-processor/pkg/processor"
	"github.com/analogj/lodestone-processor/pkg/processor/api"
	log "github.com/sirupsen/logrus"
	"gopkg.in/gographics/imagick.v2/imagick"
	"io/ioutil"
	"math"
	"net/url"
	"os"
	"path/filepath"
)

type ThumbnailProcessor struct {
	processor.CommonProcessor

	apiEndpoint *url.URL
	filter      *model.Filter
}

func CreateThumbnailProcessor(apiEndpoint string) (ThumbnailProcessor, error) {

	apiEndpointUrl, err := url.Parse(apiEndpoint)
	if err != nil {
		return ThumbnailProcessor{}, err
	}

	//retrieve the filters (include/excludes) from the API
	filterData, err := api.GetIncludeExcludeData(apiEndpointUrl)
	if err != nil {
		return ThumbnailProcessor{}, err
	}

	tp := ThumbnailProcessor{
		apiEndpoint: apiEndpointUrl,
		filter:      &filterData,
	}

	return tp, nil
}

func (tp *ThumbnailProcessor) Process(body []byte) error {

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
	includeDocument := tp.filter.ValidPath(docBucketPath)
	if !includeDocument {
		log.Infof("Ignoring document, matches exclude pattern (%s, %s)", docBucketName, docBucketPath)
		return nil
	}

	//make a temporary directory for subsequent processing (original file download, and thumb generation)
	dir, err := ioutil.TempDir("", "thumb")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir) // clean up

	if event.Records[0].EventName == "s3:ObjectRemoved:Delete" {
		log.Debugln("Attempting to delete thumbnail file")

		thumbStoragePath := api.GenerateThumbnailStoragePath(docBucketPath)
		err = api.DeleteFile(tp.apiEndpoint, "thumbnails", thumbStoragePath)
		if err != nil {
			return err
		}
		return nil

	} else {
		filePath, err := api.ReadFile(tp.apiEndpoint, docBucketName, docBucketPath, dir)
		if err != nil {
			return err
		}
		if tp.IsEmptyFile(filePath) {
			log.Infof("Ignoring document, filesize is 0 (%s, %s)", docBucketName, docBucketPath)
			return nil
		}

		thumbFilePath, err := generateThumbnail(filePath, dir)
		if err != nil {
			return err
		}

		//convert extension to jpg before uploading
		thumbStoragePath := api.GenerateThumbnailStoragePath(docBucketPath)
		err = api.CreateFile(tp.apiEndpoint, "thumbnails", thumbStoragePath, thumbFilePath)

		return err
	}
}

func generateThumbnail(docFilePath string, outputDirectory string) (string, error) {
	maxThumbWidth := 500
	maxThumbHeight := 800

	imagick.Initialize()
	defer imagick.Terminate()

	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	//code from https://github.com/gographics/imagick/issues/170

	fmt.Println("----> reading the original document...")

	// load the file blob as image.
	dat, err := ioutil.ReadFile(docFilePath)

	err = mw.ReadImageBlob(dat)

	fmt.Println("----> finished reading original document")

	if err != nil {
		return "", err
	}

	// Go to page one, if it's an PDF file.
	mw.SetIteratorIndex(0)

	// Get original size
	width := mw.GetImageWidth()
	height := mw.GetImageHeight()

	scaler := math.Max(float64(maxThumbWidth)/float64(width), float64(maxThumbHeight)/float64(height))

	if scaler < 1 {
		err = mw.ThumbnailImage(uint(float64(width)*scaler), uint(float64(height)*scaler))
		if err != nil {
			return "", err
		}
	}
	err = mw.SetImageFormat("jpg")
	if err != nil {
		return "", err
	}

	fmt.Println("----> set to jpg...")

	err = mw.SetImageAlphaChannel(imagick.ALPHA_CHANNEL_REMOVE)
	if err != nil {
		return "", err
	}

	pw := imagick.NewPixelWand()
	pw.SetColor("rgb(255,255,255)")
	err = mw.SetImageBackgroundColor(pw)
	if err != nil {
		return "", err
	}

	err = mw.SetImageCompressionQuality(95)
	if err != nil {
		return "", err
	}

	//get base filename and append the jpg file extension.
	fileName := filepath.Base(docFilePath)
	fileName = fileName + ".jpg"

	outputFilePath := filepath.Join(outputDirectory, fileName)
	err = mw.WriteImage(outputFilePath)

	return outputFilePath, err
}
