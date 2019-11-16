//A majority of this file was taken from https://github.com/aws/aws-lambda-go/blob/master/events/s3.go
package model

// Copyright 2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

type S3Event struct {
	Records []S3EventRecord `json:"Records"`
}

type S3EventRecord struct {
	EventVersion      string              `json:"eventVersion"`
	EventSource       string              `json:"eventSource"`
	AWSRegion         string              `json:"awsRegion"`
	EventTime         time.Time           `json:"eventTime"`
	EventName         string              `json:"eventName"`
	PrincipalID       S3UserIdentity      `json:"userIdentity"`
	RequestParameters S3RequestParameters `json:"requestParameters"`
	ResponseElements  map[string]string   `json:"responseElements"`
	S3                S3Entity            `json:"s3"`
}

type S3UserIdentity struct {
	PrincipalID string `json:"principalId"`
}

type S3RequestParameters struct {
	SourceIPAddress string `json:"sourceIPAddress"`
}

type S3Entity struct {
	SchemaVersion   string   `json:"s3SchemaVersion"`
	ConfigurationID string   `json:"configurationId"`
	Bucket          S3Bucket `json:"bucket"`
	Object          S3Object `json:"object"`
}

type S3Bucket struct {
	Name          string         `json:"name"`
	OwnerIdentity S3UserIdentity `json:"ownerIdentity"`
	Arn           string         `json:"arn"`
}

type S3Object struct {
	Key           string `json:"key"`
	Size          int64  `json:"size"`
	URLDecodedKey string `json:"urlDecodedKey"`
	VersionID     string `json:"versionId"`
	ETag          string `json:"eTag"`
	Sequencer     string `json:"sequencer"`
}

type S3TestEvent struct {
	Service   string    `json:"Service"`
	Bucket    string    `json:"Bucket"`
	Event     string    `json:"Event"`
	Time      time.Time `json:"Time"`
	RequestID string    `json:"RequestId"`
	HostID    string    `json:"HostId"`
}

// Helper function written for Lodestone

func (e *S3Event) Create(publisherName string, eventName string, sourceBucket string, sourceBucketKey string, sourceRawPath string) error {

	fileSize := int64(0)
	fileMD5 := ""
	if eventName == "s3:ObjectCreated:Put" {
		fileMetadata, err := os.Stat(sourceRawPath)
		if err != nil {
			return err
		}
		fileSize = fileMetadata.Size()
		fileMD5, err = fileMD5Hash(sourceRawPath)
	}

	record := S3EventRecord{
		EventVersion: "2.0",
		EventSource:  fmt.Sprintf("lodestone:publisher:%s", publisherName),
		AWSRegion:    "",
		EventTime:    time.Now(),
		EventName:    eventName,
		PrincipalID: S3UserIdentity{
			PrincipalID: "lodestone",
		},
		RequestParameters: S3RequestParameters{
			SourceIPAddress: localIP(),
		},
		ResponseElements: make(map[string]string),
		S3: S3Entity{
			SchemaVersion:   "1.0",
			ConfigurationID: "Config",
			Bucket: S3Bucket{
				Name: sourceBucket,
				OwnerIdentity: S3UserIdentity{
					PrincipalID: "lodestone",
				},
				Arn: fmt.Sprintf("arn:aws:s3:::%s", sourceBucket),
			},
			Object: S3Object{
				Key:       sourceBucketKey,
				Size:      fileSize,
				ETag:      fileMD5,
				VersionID: "1",
			},
		},
	}

	e.Records = []S3EventRecord{record}
	return nil
}

func localIP() string {
	host, _ := os.Hostname()
	addrs, _ := net.LookupIP(host)
	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			return ipv4.String()
		}
	}
	return ""
}

func fileMD5Hash(filePath string) (string, error) {
	//Initialize variable returnMD5String now in case an error has to be returned
	var returnMD5String string

	//Open the passed argument and check for any error
	file, err := os.Open(filePath)
	if err != nil {
		return returnMD5String, err
	}

	//Tell the program to call the following function when the current function returns
	defer file.Close()

	//Open a new hash interface to write to
	hash := md5.New()

	//Copy the file in the hash interface and check for any error
	if _, err := io.Copy(hash, file); err != nil {
		return returnMD5String, err
	}

	//Get the 16 bytes hash
	hashInBytes := hash.Sum(nil)[:16]

	//Convert the bytes to a string
	returnMD5String = hex.EncodeToString(hashInBytes)

	return returnMD5String, nil

}
