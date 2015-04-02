package commands

import (
	"github.com/codegangsta/cli"
	"github.com/ehazlett/freight"
	"github.com/samalba/dockerclient"
)

func getClient(c *cli.Context) (*dockerclient.DockerClient, error) {
	dockerUrl := c.GlobalString("docker")
	tlsCaCert := c.GlobalString("tls-ca-cert")
	tlsCert := c.GlobalString("tls-cert")
	tlsKey := c.GlobalString("tls-key")
	allowInsecure := c.GlobalBool("allow-insecure")
	return freight.GetClient(dockerUrl, tlsCaCert, tlsCert, tlsKey, allowInsecure)
}
