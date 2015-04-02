package commands

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/docker/pkg/units"
	"github.com/ehazlett/freight"
)

var CmdLs = cli.Command{
	Name:   "ls",
	Usage:  "list application containers",
	Action: cmdLs,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "version, v",
			Usage: "app version (overrides config)",
			Value: "",
		},
	},
}

var cmdLs = func(c *cli.Context) {
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

	containers, err := client.ListContainers(true, false, "")
	if err != nil {
		log.Fatal(err)
	}

	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprintln(w, "CONTAINER ID\tNAME\tVERSION\tCREATED\tSTATUS\tPORTS")

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

		cTime := units.HumanDuration(time.Now().UTC().Sub(time.Unix(cnt.Created, 0)))

		exposedPorts := []string{}

		for k, v := range cntInfo.NetworkSettings.Ports {
			for _, port := range v {
				exposedPorts = append(exposedPorts, fmt.Sprintf("%s:%s->%s", port.HostIp, port.HostPort, k))
			}
		}

		portDisplay := strings.Join(exposedPorts, ",")
		if cName != "" && cVersion != "" {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				cId, cName, cVersion, cTime, cnt.Status, portDisplay)
		}
	}

	w.Flush()
}
