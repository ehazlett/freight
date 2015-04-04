package commands

import (
	"fmt"

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
		cli.StringFlag{
			Name:  "hostname",
			Usage: "hostname (overrides config)",
			Value: "",
		},
		cli.StringFlag{
			Name:  "domainname",
			Usage: "domain name (overrides config)",
			Value: "",
		},
		cli.StringFlag{
			Name:   "auth-username",
			Usage:  "Registry auth username",
			Value:  "",
			EnvVar: "FREIGHT_AUTH_USERNAME",
		},
		cli.StringFlag{
			Name:   "auth-password",
			Usage:  "Registry auth password",
			Value:  "",
			EnvVar: "FREIGHT_AUTH_PASSWORD",
		},
		cli.StringFlag{
			Name:   "auth-email",
			Usage:  "Registry auth email",
			Value:  "",
			EnvVar: "FREIGHT_AUTH_EMAIL",
		},
		cli.BoolFlag{
			Name:  "force-build, b",
			Usage: "force building image from repo",
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
	hostname := c.String("hostname")
	domainname := c.String("domainname")
	authUsername := c.String("auth-username")
	authPassword := c.String("auth-password")
	authEmail := c.String("auth-email")
	forceBuild := c.Bool("force-build")

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

	if hostname == "" {
		hostname = config.Hostname
	}

	if domainname == "" {
		domainname = config.Domainname
	}

	log.Infof("deploy: name=%s image=%s version=%s repo=%s", config.Name, config.Image, version, config.Repo)

	imageName := fmt.Sprintf("%s:%s", config.Image, version)

	if config.BuildImage || forceBuild {
		image := dockerclient.BuildImage{
			Name:           imageName,
			Remote:         config.Repo,
			DockerfilePath: config.DockerfilePath,
		}

		log.Debugf("building: image=%s", imageName)

		if err := client.BuildImage(image, nil); err != nil {
			log.Fatal(err)
		}
	} else {
		var authConfig *dockerclient.AuthConfig

		if authUsername != "" && authPassword != "" {

			authConfig = &dockerclient.AuthConfig{
				Username: authUsername,
				Password: authPassword,
				Email:    authEmail,
			}
		}

		log.Debugf("pulling: image=%s", imageName)

		if err := client.PullImage(imageName, authConfig); err != nil {
			log.Fatal(err)
		}
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
			Hostname:   hostname,
			Domainname: domainname,
			Image:      imageName,
			Env:        cntEnv,
			Labels:     config.Labels,
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

	log.Infof("successfully deployed name=%s version=%s", config.Name, version)
}
