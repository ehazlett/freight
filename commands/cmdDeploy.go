package commands

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/ehazlett/freight"
	"github.com/samalba/dockerclient"
)

var CmdDeploy = cli.Command{
	Name:   "deploy",
	Usage:  "deploy an application",
	Action: cmdDeploy,
	Flags: []cli.Flag{
		cli.IntFlag{
			Name:  "instances, i",
			Usage: "number of instances to deploy (overrides config)",
			Value: -1,
		},
		cli.StringFlag{
			Name:  "version, v",
			Usage: "app version (overrides config)",
			Value: "",
		},
	},
}

var cmdDeploy = func(c *cli.Context) {
	client, err := getClient(c)
	if err != nil {
		log.Fatal(err)
	}

	instances := c.Int("instances")
	version := c.String("version")

	configPath := c.GlobalString("config")

	config, err := freight.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	// config overrides
	if instances == -1 {
		instances = config.Instances
	}

	if version == "" {
		version = config.Version
	}

	log.Infof("deploy: name=%s version=%s repo=%s", config.Name, version, config.Repo)

	imageName := fmt.Sprintf("%s:%s", config.Name, version)

	image := dockerclient.BuildImage{
		Name:           imageName,
		Remote:         config.Repo,
		DockerfilePath: config.DockerfilePath,
	}

	log.Debugf("building image: name=%s version=%s", config.Name, version)

	if err := client.BuildImage(image, nil); err != nil {
		log.Fatal(err)
	}

	hostConfig := &dockerclient.HostConfig{
		PublishAllPorts: config.PublishAllPorts,
	}

	newIds := map[string]bool{}

	cntEnv := config.Environment
	cntEnv = append(cntEnv, fmt.Sprintf("FREIGHT_NAME=%s", config.Name))
	cntEnv = append(cntEnv, fmt.Sprintf("FREIGHT_VERSION=%s", version))

	for i := 0; i < instances; i++ {
		log.Debugf("starting instance: image=%s instance=%d", imageName, i)
		// inject env vars
		containerConfig := &dockerclient.ContainerConfig{
			Image:  imageName,
			Env:    cntEnv,
			Labels: config.Labels,
		}

		id, err := client.CreateContainer(containerConfig, "")
		if err != nil {
			log.Fatal(err)
		}

		if err := client.StartContainer(id, hostConfig); err != nil {
			log.Fatal(err)
		}

		newIds[id] = true
	}

	// remove old containers
	containers, err := client.ListContainers(true, false, "")
	if err != nil {
		log.Fatal(err)
	}

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
			if _, ok := newIds[cnt.Id]; !ok {
				log.Debugf("removing container: id=%s image=%s", cId, cnt.Image)
				if err := client.RemoveContainer(cnt.Id, true, true); err != nil {
					log.Warnf("unable to remove container: id=%s image=%s", cId, cnt.Image)
				}
			}

		}
	}

	log.Infof("successfully deployed name=%s version=%s", config.Name, version)
}
