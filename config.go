package freight

type Config struct {
	Name            string            `json:"name,omitempty"`
	Version         string            `json:"version,omitempty"`
	Repo            string            `json:"repo,omitempty"`
	Labels          map[string]string `json:"labels,omitempty"`
	Instances       int               `json:"instances,omitempty"`
	Environment     []string          `json:"environment,omitempty"`
	RestartPolicy   string            `json:"restart_policy,omitempty"`
	DockerfilePath  string            `json:"dockerfile_path,omitempty"`
	PublishAllPorts bool              `json:"publish_all_ports,omitempty"`
	Hostname        string            `json:"hostname,omitempty"`
	Domainname      string            `json:"domain,omitempty"`
}
