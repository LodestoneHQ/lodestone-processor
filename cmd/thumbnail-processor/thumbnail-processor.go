package main

import (
	"fmt"
	"github.com/analogj/go-util/utils"
	"github.com/analogj/lodestone-processor/pkg/listen"
	"github.com/analogj/lodestone-processor/pkg/processor/thumbnail"
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
		Name:     "lodestone-thumbnail-processor",
		Usage:    "Thumbnail processor for lodestone",
		Version:  version.VERSION,
		Compiled: time.Now(),
		Authors: []cli.Author{
			cli.Author{
				Name:  "Jason Kulatunga",
				Email: "jason@thesparktree.com",
			},
		},
		Before: func(c *cli.Context) error {

			capsuleUrl := "AnalogJ/lodestone-processor:thumbnail"

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
				Usage: "Start the Lodestone thumbnail processor",
				Action: func(c *cli.Context) error {
					processorLogger := logrus.WithFields(logrus.Fields{
						"type": "thumbnail",
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

					thumbnailProcessor, err := thumbnail.CreateThumbnailProcessor(processorLogger, c.String("api-endpoint"))
					if err != nil {
						return err
					}
					return listenClient.Subscribe(thumbnailProcessor.Process)
				},

				Flags: []cli.Flag{

					&cli.StringFlag{
						Name:  "api-endpoint",
						Usage: "The api server endpoint",
						Value: "http://webapp:3000",
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
						Value: "thumbnails",
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
		logrus.Fatal(color.HiRedString("thumbnail ERROR: %v", err))
	}
}
