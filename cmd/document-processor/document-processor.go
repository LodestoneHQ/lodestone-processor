package main

import (
	"fmt"
	"github.com/analogj/go-util/utils"
	"github.com/analogj/lodestone-processor/pkg/listen"
	"github.com/analogj/lodestone-processor/pkg/processor/document"
	"github.com/analogj/lodestone-processor/pkg/version"
	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"os"
	"time"
)

var goos string
var goarch string

func main() {
	app := &cli.App{
		Name:     "lodestone-document-processor",
		Usage:    "Document processor for lodestone",
		Version:  version.VERSION,
		Compiled: time.Now(),
		Authors: []cli.Author{
			cli.Author{
				Name:  "Jason Kulatunga",
				Email: "jason@thesparktree.com",
			},
		},
		Before: func(c *cli.Context) error {

			capsuleUrl := "AnalogJ/lodestone-processor:document"

			versionInfo := fmt.Sprintf("%s.%s-%s", goos, goarch, version.VERSION)

			subtitle := capsuleUrl + utils.LeftPad2Len(versionInfo, " ", 53-len(capsuleUrl))

			fmt.Fprintf(c.App.Writer, fmt.Sprintf(utils.StripIndent(
				`
			 __    _____  ____  ____  ___  ____  _____  _  _  ____ 
			(  )  (  _  )(  _ \( ___)/ __)(_  _)(  _  )( \( )( ___)
			 )(__  )(_)(  )(_) ))__) \__ \  )(   )(_)(  )  (  )__) 
			(____)(_____)(____/(____)(___/ (__) (_____)(_)\_)(____)
			%s
			`), subtitle))
			return nil
		},

		Commands: []cli.Command{
			{
				Name:  "start",
				Usage: "Start the Lodestone document processor",
				Action: func(c *cli.Context) error {

					processorLogger := logrus.WithFields(logrus.Fields{
						"type": "document",
					})

					if c.Bool("debug") {
						logrus.SetLevel(logrus.DebugLevel)
					} else {
						logrus.SetLevel(logrus.InfoLevel)
					}

					var listenClient listen.Interface

					listenClient = new(listen.AmqpListen)
					err := listenClient.Init(processorLogger, map[string]string{
						"amqp-url": c.String("amqp-url"),
						"exchange": c.String("amqp-exchange"),
						"queue":    c.String("amqp-queue"),
					})
					if err != nil {
						return err
					}
					defer listenClient.Close()

					documentProcessor, err := document.CreateDocumentProcessor(
						processorLogger,
						c.String("api-endpoint"),
						c.String("storage-thumbnail-bucket"),
						c.String("tika-endpoint"),
						c.String("elasticsearch-endpoint"),
						c.String("elasticsearch-index"),
						c.String("elasticsearch-mapping"),
						c.String("ocr-language"),
					)

					if err != nil {
						return err
					}

					return listenClient.Subscribe(documentProcessor.Process)
				},

				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "api-endpoint",
						Usage: "The api server endpoint",
						Value: "http://webapp:3000",
					},
					&cli.StringFlag{
						Name:  "storage-thumbnail-bucket",
						Usage: "The thumbnail bucket on the storage server",
						Value: "thumbnails",
					},

					&cli.StringFlag{
						Name:  "tika-endpoint",
						Usage: "The tika server endpoint",
						Value: "http://tika:9998",
					},
					&cli.StringFlag{
						Name:  "elasticsearch-endpoint",
						Usage: "The elasticsearch server endpoint",
						Value: "http://elasticsearch:9200",
					},
					&cli.StringFlag{
						Name:  "elasticsearch-index",
						Usage: "The elasticsearch index to store documents in",
						Value: "lodestone",
					},
					&cli.StringFlag{
						Name:  "elasticsearch-mapping",
						Usage: "Path to elasticsearch mapping file. Can be used to override static/document-processor/settings.json",
						Value: "",
					},

					&cli.StringFlag{
						Name:  "ocr-language",
						Usage: "OCR language override for Tika requests",
						Value: "",
					},

					&cli.StringFlag{
						Name:  "amqp-url",
						Usage: "The amqp connection string",
						Value: "amqp://guest:guest@localhost:5672",
					},

					&cli.StringFlag{
						Name:  "amqp-exchange",
						Usage: "The amqp exchange",
						Value: "lodestone",
					},

					&cli.StringFlag{
						Name:  "amqp-queue",
						Usage: "The amqp queue",
						Value: "documents",
					},

					&cli.BoolFlag{
						Name:  "debug",
						Usage: "Enable debug logging",
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		logrus.Fatal(color.HiRedString("document ERROR: %v", err))
	}
}
