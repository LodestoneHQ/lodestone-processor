package processor

import (
	"github.com/analogj/lodestone-processor/pkg/model"
	"github.com/minio/minio-go/v6"
	"net/url"
	"os"
	"path/filepath"
)

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

	return bucketName, documentPath, nil
}

func retrieveDocument(storageEndpoint *url.URL, storageBucket string, storagePath string, outputDirectory string) (string, error) {

	secureProtocol := storageEndpoint.Scheme == "https"

	s3Client, err := minio.New(storageEndpoint.Host, os.Getenv("MINIO_ACCESS_KEY"), os.Getenv("MINIO_SECRET_KEY"), secureProtocol)
	if err != nil {
		return "", err
	}

	fileName := filepath.Base(storagePath)
	localFilepath := filepath.Join(outputDirectory, fileName)

	err = s3Client.FGetObject(storageBucket, storagePath, localFilepath, minio.GetObjectOptions{})

	return localFilepath, err
}
