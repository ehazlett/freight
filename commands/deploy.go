package commands

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/ehazlett/freight"
	"github.com/samalba/dockerclient"
)

var CmdDeploy = func(c *cli.Context) {
	dockerUrl := c.GlobalString("docker")
	tlsCaCert := c.GlobalString("tls-ca-cert")
	tlsCert := c.GlobalString("tls-cert")
	tlsKey := c.GlobalString("tls-key")
	allowInsecure := c.GlobalBool("allow-insecure")
	client, err := freight.GetClient(dockerUrl, tlsCaCert, tlsCert, tlsKey, allowInsecure)
	if err != nil {
		log.Fatal(err)
	}

	instances := c.Int("instances")

	configPath := c.String("config")

	config, err := freight.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	if instances == -1 {
		instances = config.Instances
	}

	log.Infof("deploying: name=%s version=%s repo=%s", config.Name, config.Version, config.Repo)

	imageName := config.Name

	image := dockerclient.BuildImage{
		Name:           imageName,
		Remote:         config.Repo,
		DockerfilePath: config.DockerfilePath,
	}

	log.Debugf("building image: name=%s", config.Name)
	if err := client.BuildImage(image, nil); err != nil {
		log.Fatal(err)
	}

	hostConfig := &dockerclient.HostConfig{
		PublishAllPorts: config.PublishAllPorts,
	}

	newIds := map[string]bool{}

	cntEnv := config.Environment
	cntEnv = append(cntEnv, fmt.Sprintf("FREIGHT_NAME=%s", config.Name))
	cntEnv = append(cntEnv, fmt.Sprintf("FREIGHT_VERSION=%s", config.Version))

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

		for _, v := range cntInfo.Config.Env {
			// parse the env to get only the freight controlled containers
			parts := strings.Split(v, "=")
			if len(parts) != 2 {
				continue
			}

			k := parts[0]
			v := parts[1]

			cId := cnt.Id[:12]
			// only remove containers of the same name
			if k == "FREIGHT_NAME" && v == config.Name {
				if _, ok := newIds[cnt.Id]; !ok {
					log.Debugf("removing container: id=%s image=%s", cId, cnt.Image)
					if err := client.RemoveContainer(cnt.Id, true, true); err != nil {
						log.Warnf("unable to remove container: id=%s image=%s", cId, cnt.Image)
					}
				}

				break
			}
		}
	}

	log.Infof("successfully deployed name=%s version=%s", config.Name, config.Version)
}
