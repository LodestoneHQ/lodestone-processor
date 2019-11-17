package main

import (
	"fmt"
	"github.com/analogj/go-util/utils"
	"github.com/analogj/lodestone-processor/pkg/listen"
	"github.com/analogj/lodestone-processor/pkg/processor"
	"github.com/analogj/lodestone-processor/pkg/version"
	"github.com/fatih/color"
	"github.com/urfave/cli"
	"log"
	"os"
	"time"
)

var goos string
var goarch string

func main() {
	app := &cli.App{
		Name:     "lodestone-document-processor",
		Usage:    "Notification processor for lodestone",
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

					var listenClient listen.Interface

					listenClient = new(listen.AmqpListen)
					err := listenClient.Init(map[string]string{
						"amqp-url":         c.String("amqp-url"),
						"exchange":         c.String("amqp-exchange"),
						"queue":            c.String("amqp-queue"),
						"storage-endpoint": c.String("storage-endpoint"),
						"tika-url":         c.String("tika-url"),
					})
					if err != nil {
						return err
					}
					defer listenClient.Close()

					return listenClient.Subscribe(processor.DocumentProcessor)
				},

				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "storage-endpoint",
						Usage: "The storage server endpoint",
						Value: "storage:9000",
					},

					&cli.StringFlag{
						Name:  "tika-url",
						Usage: "The tika server url",
						Value: "http://tika:9998",
					},

					&cli.StringFlag{
						Name:  "amqp-url",
						Usage: "The amqp connection string",
						Value: "amqp://guest:guest@localhost:5672",
					},

					&cli.StringFlag{
						Name:  "amqp-exchange",
						Usage: "The amqp exchange",
						Value: "storageevents",
					},

					&cli.StringFlag{
						Name:  "amqp-queue",
						Usage: "The amqp queue",
						Value: "documents",
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(color.HiRedString("ERROR: %v", err))
	}
}
