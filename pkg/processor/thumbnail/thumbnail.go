package thumbnail

import (
	"encoding/json"
	"fmt"
	"github.com/analogj/lodestone-processor/pkg/model"
	"github.com/analogj/lodestone-processor/pkg/processor"
	"github.com/minio/minio-go/v6"
	"gopkg.in/gographics/imagick.v2/imagick"
	"io/ioutil"
	"math"
	"net/url"
	"os"
	"path"
	"path/filepath"
)

type ThumbnailProcessor struct {
	storageEndpoint *url.URL
}

func CreateThumbnailProcessor(storageEndpoint string) (ThumbnailProcessor, error) {

	storageEndpointUrl, err := url.Parse(storageEndpoint)
	if err != nil {
		return ThumbnailProcessor{}, err
	}

	tp := ThumbnailProcessor{
		storageEndpoint: storageEndpointUrl,
	}

	return tp, nil
}

func (tp *ThumbnailProcessor) Process(body []byte) error {
	//make a temporary directory for subsequent processing (original file download, and thumb generation)
	dir, err := ioutil.TempDir("", "thumb")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir) // clean up

	var event model.S3Event
	err = json.Unmarshal(body, &event)
	if err != nil {
		return err
	}

	docBucketName, docBucketPath, err := processor.GenerateStoragePath(event)
	if err != nil {
		return err
	}

	filePath, err := processor.RetrieveDocument(tp.storageEndpoint, docBucketName, docBucketPath, dir)
	if err != nil {
		return err
	}

	thumbFilePath, err := generateThumbnail(filePath, dir)
	if err != nil {
		return err
	}

	err = uploadThumbnail(tp.storageEndpoint, "thumbnails", docBucketPath, thumbFilePath)

	return err

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

	//get base filename and change the file extension.
	fileName := filepath.Base(docFilePath)
	ext := path.Ext(fileName)
	fileName = fileName[0:len(fileName)-len(ext)] + ".jpg"

	outputFilePath := filepath.Join(outputDirectory, fileName)
	err = mw.WriteImage(outputFilePath)

	return outputFilePath, err
}

func uploadThumbnail(storageEndpoint *url.URL, storageBucket string, storagePath string, thumbFilePath string) error {

	//convert extension to jpg before uploading
	thumbStoragePath := processor.GenerateThumbnailStoragePath(storagePath)

	secureProtocol := storageEndpoint.Scheme == "https"

	s3Client, err := minio.New(storageEndpoint.Host, os.Getenv("MINIO_ACCESS_KEY"), os.Getenv("MINIO_SECRET_KEY"), secureProtocol)
	if err != nil {
		return err
	}

	if _, err := s3Client.FPutObject(storageBucket, thumbStoragePath, thumbFilePath, minio.PutObjectOptions{
		ContentType: "image/jpeg",
	}); err != nil {
		return err
	}

	return nil
}
