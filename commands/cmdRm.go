package commands

import (
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/ehazlett/freight"
)

var CmdRemove = cli.Command{
	Name:   "rm",
	Usage:  "remove an application",
	Action: cmdRemove,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "version, v",
			Usage: "app version (overrides config)",
			Value: "",
		},
	},
}

var cmdRemove = func(c *cli.Context) {
	client, err := getClient(c)
	if err != nil {
		log.Fatal(err)
	}

	version := c.String("version")

	configPath := c.GlobalString("config")

	config, err := freight.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	if version == "" {
		version = config.Version
	}

	// remove old containers
	containers, err := client.ListContainers(true, false, "")
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("remove: name=%s version=%s", config.Name, version)

	for _, cnt := range containers {
		cntInfo, err := client.InspectContainer(cnt.Id)
		if err != nil {
			log.Fatal(err)
		}

		cId := cnt.Id[:12]

		cName := ""
		cVersion := ""

		for _, v := range cntInfo.Config.Env {
			// parse the env to get only the freight controlled containers
			parts := strings.Split(v, "=")
			if len(parts) != 2 {
				continue
			}

			k := parts[0]
			v := parts[1]

			if k == "FREIGHT_NAME" {
				cName = v
				continue
			}

			if k == "FREIGHT_VERSION" {
				cVersion = v
				continue
			}

		}

		// only remove containers of the same name
		if cName == config.Name && cVersion == version {
			log.Debugf("stopping container: id=%s image=%s", cId, cnt.Image)
			if err := client.StopContainer(cId, 1); err != nil {
				log.Warnf("unable to stop container: %s", err)
				continue
			}

			time.Sleep(10 * time.Millisecond)

			log.Debugf("removing container: id=%s image=%s", cId, cnt.Image)
			if err := client.RemoveContainer(cnt.Id, true, true); err != nil {
				log.Warnf("unable to remove container: id=%s image=%s", cId, cnt.Image)
			}
		}
	}
}
