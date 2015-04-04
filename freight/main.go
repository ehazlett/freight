package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/ehazlett/freight/commands"
	"github.com/ehazlett/freight/version"
)

func main() {
	app := cli.NewApp()
	app.Name = "freight"
	app.Usage = "app deployment"
	app.Version = version.Version + " (" + version.Gitcommit + ")"
	app.Author = "@ehazlett"
	app.Email = ""
	app.Before = func(c *cli.Context) error {
		if c.GlobalBool("debug") {
			log.SetLevel(log.DebugLevel)
		}
		return nil
	}
	app.Commands = []cli.Command{
		commands.CmdDeploy,
		commands.CmdLs,
		commands.CmdRemove,
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Usage: "freight config url / path / github repo",
			Value: "freight.json",
		},
		cli.StringFlag{
			Name:   "docker, d",
			Value:  "unix:///var/run/docker.sock",
			Usage:  "docker swarm addr",
			EnvVar: "DOCKER_HOST",
		},
		cli.StringFlag{
			Name:  "tls-ca-cert",
			Value: "",
			Usage: "tls ca certificate",
		},
		cli.StringFlag{
			Name:  "tls-cert",
			Value: "",
			Usage: "tls certificate",
		},
		cli.StringFlag{
			Name:  "tls-key",
			Value: "",
			Usage: "tls key",
		},
		cli.BoolFlag{
			Name:  "allow-insecure",
			Usage: "enable insecure tls communication",
		},
		cli.BoolFlag{
			Name:  "debug, D",
			Usage: "enable debug",
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
