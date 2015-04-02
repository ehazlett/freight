package freight

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/samalba/dockerclient"
)

func getReaderFromPath(configPath string) (io.Reader, error) {
	var (
		rdr io.Reader
	)

	isHttp := false

	if strings.Index(configPath, "http") != -1 {
		isHttp = true
	}

	// check for github url
	if strings.Index(configPath, "github.com") != -1 && strings.Index(configPath, "freight.json") == -1 {
		configPath = fmt.Sprintf("https://%s", path.Join(configPath, "raw", "master", "freight.json"))
		isHttp = true
	}

	if isHttp {
		log.Debugf("loading config from http")
		r, err := http.Get(configPath)
		if err != nil {
			return nil, err
		}

		if r.StatusCode != 200 {
			content, err := ioutil.ReadAll(r.Body)
			if err != nil {
				return nil, err
			}

			log.Debug(string(content))
			return nil, fmt.Errorf("cannot load configuration: status=%d", r.StatusCode)
		}

		rdr = r.Body
	} else {
		log.Debugf("loading config from file")
		f, err := os.Open(configPath)
		if err != nil {
			return nil, err
		}

		rdr = f
	}

	return rdr, nil
}

func LoadConfig(path string) (*Config, error) {
	// load config
	r, err := getReaderFromPath(path)
	if err != nil {
		return nil, err
	}

	if r == nil {
		return nil, fmt.Errorf("unable to load config:  unable to get reader from path")
	}

	var config *Config
	if err := json.NewDecoder(r).Decode(&config); err != nil {
		log.Fatal(err)
	}

	return config, nil
}

func GetTLSConfig(caCert, cert, key []byte, allowInsecure bool) (*tls.Config, error) {
	// TLS config
	var tlsConfig tls.Config
	tlsConfig.InsecureSkipVerify = true
	certPool := x509.NewCertPool()

	certPool.AppendCertsFromPEM(caCert)
	tlsConfig.RootCAs = certPool
	keypair, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return &tlsConfig, err
	}
	tlsConfig.Certificates = []tls.Certificate{keypair}
	if allowInsecure {
		tlsConfig.InsecureSkipVerify = true
	}

	return &tlsConfig, nil
}

func GetClient(dockerUrl, tlsCaCert, tlsCert, tlsKey string, allowInsecure bool) (*dockerclient.DockerClient, error) {
	// only load env vars if no args
	// check environment for docker client config
	envDockerHost := os.Getenv("DOCKER_HOST")
	if dockerUrl == "" && envDockerHost != "" {
		dockerUrl = envDockerHost
	}

	// only load env vars if no args
	envDockerCertPath := os.Getenv("DOCKER_CERT_PATH")
	envDockerTlsVerify := os.Getenv("DOCKER_TLS_VERIFY")
	if tlsCaCert == "" && envDockerCertPath != "" && envDockerTlsVerify != "" {
		tlsCaCert = filepath.Join(envDockerCertPath, "ca.pem")
		tlsCert = filepath.Join(envDockerCertPath, "cert.pem")
		tlsKey = filepath.Join(envDockerCertPath, "key.pem")
	}

	// load tlsconfig
	var tlsConfig *tls.Config
	if tlsCaCert != "" && tlsCert != "" && tlsKey != "" {
		log.Debug("using tls for communication with docker")
		caCert, err := ioutil.ReadFile(tlsCaCert)
		if err != nil {
			log.Fatalf("error loading tls ca cert: %s", err)
		}

		cert, err := ioutil.ReadFile(tlsCert)
		if err != nil {
			log.Fatalf("error loading tls cert: %s", err)
		}

		key, err := ioutil.ReadFile(tlsKey)
		if err != nil {
			log.Fatalf("error loading tls key: %s", err)
		}

		cfg, err := GetTLSConfig(caCert, cert, key, allowInsecure)
		if err != nil {
			log.Fatalf("error configuring tls: %s", err)
		}
		tlsConfig = cfg
	}

	log.Debugf("docker client: url=%s", dockerUrl)

	client, err := dockerclient.NewDockerClient(dockerUrl, tlsConfig)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func FromUnixTimestamp(timestamp int64) (*time.Time, error) {
	i, err := strconv.ParseInt("1405544146", 10, 64)
	if err != nil {
		return nil, err
	}

	t := time.Unix(i, 0)
	return &t, nil
}
