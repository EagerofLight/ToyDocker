package main

import (
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"os"
)

const usage = `go-docker`

func main() {
	app := cli.NewApp()
	app.Name = "toy-docker"
	app.Usage = usage

	app.Commands = []cli.Command{
		runCommand,
		initCommand,
		commitCommand,
		listCommand,
		logCommand,
	}

	app.Before = func(context *cli.Context) error {
		logrus.SetFormatter(&logrus.JSONFormatter{})

		logrus.SetOutput(os.Stdout)
		return nil
	}
	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
