package processor

import (
	"encoding/json"
	"fmt"
	"github.com/analogj/lodestone-processor/pkg/model"
	"gopkg.in/gographics/imagick.v2/imagick"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"path/filepath"
)

func ThumbnailProcessor(body []byte, storageUrl string) error {
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

	storageDocPath, storageThumbPath, err := generateStoragePath(event)
	if err != nil {
		return err
	}

	filePath, err := retrieveDocument(storageUrl, storageDocPath, dir)
	if err != nil {
		return err
	}

	thumbFilePath, err := generateThumbnail(filePath, dir)
	if err != nil {
		return err
	}

	err = uploadThumbnail(storageUrl, thumbFilePath, storageThumbPath)

	return err

}

func generateStoragePath(event model.S3Event) (string, string, error) {
	/*
		{
			"Records": [{
				"eventVersion": "2.0",
				"eventSource": "lodestone:publisher:fs",
				"awsRegion": "",
				"eventTime": "2019-11-16T23:46:21.1467633Z",
				"eventName": "s3:ObjectRemoved:Delete",
				"userIdentity": {
					"principalId": "lodestone"
				},
				"requestParameters": {
					"sourceIPAddress": "172.19.0.5"
				},
				"responseElements": {},
				"s3": {
					"s3SchemaVersion": "1.0",
					"configurationId": "Config",
					"bucket": {
						"name": "documents",
						"ownerIdentity": {
							"principalId": "lodestone"
						},
						"arn": "arn:aws:s3:::documents"
					},
					"object": {
						"key": "filetypes/fIoiDm",
						"size": 0,
						"urlDecodedKey": "",
						"versionId": "1",
						"eTag": "",
						"sequencer": ""
					}
				}
			}]
		}
	*/
	bucketName := event.Records[0].S3.Bucket.Name
	documentPath := event.Records[0].S3.Object.Key

	return fmt.Sprintf("%s/%s", bucketName, documentPath), fmt.Sprintf("thumbnails/%s", documentPath), nil
}

func retrieveDocument(storageUrl string, storagePath string, outputDirectory string) (string, error) {
	u, err := url.Parse(storageUrl)
	if err != nil {
		return "", err
	}
	storageFileUrl := path.Join(u.Path, storagePath)

	// Get the data
	resp, err := http.Get(storageFileUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Create the file
	outputFilePath := filepath.Join(outputDirectory, storagePath)
	out, err := os.Create(outputFilePath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return outputFilePath, err
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

	//thumbnailImage := mw.GetImageBlob()
	fileName := filepath.Base(docFilePath)
	outputFilePath := filepath.Join(outputDirectory, fileName)
	err = mw.WriteImage(outputFilePath)

	return outputFilePath, err
}

func uploadThumbnail(storageUrl string, thumbFilePath string, storageThumbPath string) error {

	u, err := url.Parse(storageUrl)
	if err != nil {
		return err
	}
	storageThumbFileUrl := path.Join(u.Path, storageThumbPath)

	// Open a file.
	f, err := os.Open(thumbFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Post a file to URL.
	resp, err := http.Post(storageThumbFileUrl, "application/data", f)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Dump response to debug.
	b, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return err
	}
	fmt.Println(b)
	return nil
}
