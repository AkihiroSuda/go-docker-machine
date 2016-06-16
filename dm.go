// Package dm provides a binding for Docker Machine
package dm

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	"github.com/docker/engine-api/client"
	"github.com/docker/go-connections/tlsconfig"
)

const (
	// DefaultDockerMachinePath is the default path to `docker-machine`
	DefaultDockerMachinePath = "docker-machine"
)

var (
	// ExepctedMachineConfigVersions are expected, but not mandatory
	ExepctedMachineConfigVersions = []int{3}
)

// commandExecutor is used to run *exec.Cmd
type commandExecutor func(*exec.Cmd) ([]byte, error)

// DockerMachine is an instance of `docker-machine`
type DockerMachine struct {
	Path            string
	commandExecutor commandExecutor
}

// MachineState is a state of a machine
type MachineState string

// Machine is a result of `docker-machine ls`
type Machine struct {
	Name  string       `json:"Name"`
	State MachineState `json:"State"`
	URL   string       `json:"URL"`
}

// EngineOptions is included in a result of `docker-machine inspect`
type EngineOptions struct {
	TLSVerify bool `json:"TlsVerify"`
}

// AuthOptions is included in a result of `docker-machine inspect`
type AuthOptions struct {
	CaCertPath     string `json:"CaCertPath"`
	ClientCertPath string `json:"ClientCertPath"`
	ClientKeyPath  string `json:"ClientKeyPath"`
	StorePath      string `json:"StorePath"`
}

// HostOptions is included in a result of `docker-machine inspect`
type HostOptions struct {
	EngineOptions EngineOptions `json:"EngineOptions"`
	AuthOptions   AuthOptions   `json:"AuthOptions"`
}

// MachineConfig is a result of `docker-machine inspect`
type MachineConfig struct {
	ConfigVersion int         `json:"ConfigVersion"`
	HostOptions   HostOptions `json:"HostOptions"`
}

// NewDockerMachine instantiates DockerMachine
func NewDockerMachine() *DockerMachine {
	return &DockerMachine{
		Path: DefaultDockerMachinePath,
	}
}

func defaultCommandExecutor(cmd *exec.Cmd) ([]byte, error) {
	return cmd.CombinedOutput()
}

func (dm *DockerMachine) rawCommandResult(command string, arg ...string) ([]byte, error) {
	cmd := exec.Command(command, arg...)
	if dm.commandExecutor != nil {
		return dm.commandExecutor(cmd)
	}
	return defaultCommandExecutor(cmd)
}

// RawCommandResult returns a raw `docker-machine` command result with arg
func (dm *DockerMachine) RawCommandResult(arg ...string) ([]byte, error) {
	return dm.rawCommandResult(dm.Path, arg...)
}

// Machine returns the machine (`docker-machine ls`)
func (dm *DockerMachine) Machine(name string) (Machine, error) {
	machines, err := dm.Machines()
	if err != nil {
		return Machine{}, err
	}
	for _, machine := range machines {
		if machine.Name == name {
			return machine, nil
		}
	}
	return Machine{}, fmt.Errorf("machine %s not found", name)
}

// Machines returns the list of machines (`docker-machine ls`)
func (dm *DockerMachine) Machines() ([]Machine, error) {
	filter := `{"Name":"{{.Name}}","State":"{{.State}}","URL":"{{.URL}}"}`
	bytes, err := dm.RawCommandResult("ls", "-f", filter)
	if err != nil {
		return nil, err
	}
	var machines []Machine
	for _, s := range strings.Split(string(bytes), "\n") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		var machine Machine
		if err = json.Unmarshal([]byte(s), &machine); err != nil {
			return nil, fmt.Errorf("error while parsing \"%s\": %v", s, err)
		}
		machines = append(machines, machine)
	}
	return machines, nil
}

// Inspect inspects the machine (`docker-machine inspect`)
func (dm *DockerMachine) Inspect(name string) (MachineConfig, error) {
	var config MachineConfig
	bytes, err := dm.RawCommandResult("inspect", name)
	if err != nil {
		return config, err
	}
	if err = json.Unmarshal(bytes, &config); err != nil {
		return config, err
	}
	return config, nil
}

func (dm *DockerMachine) annotateError(err error, config MachineConfig, while string) error {
	for _, ver := range ExepctedMachineConfigVersions {
		if ver == config.ConfigVersion {
			return fmt.Errorf("error while %s: %v", while, err)
		}
	}
	return fmt.Errorf("error while %s, maybe due to unsupported config version %d (expected %v): %v",
		while, config.ConfigVersion, ExepctedMachineConfigVersions, err)
}

func (dm *DockerMachine) tlsConfig(config MachineConfig) (*tls.Config, error) {
	eopt := config.HostOptions.EngineOptions
	aopt := config.HostOptions.AuthOptions
	options := tlsconfig.Options{
		CAFile:             aopt.CaCertPath,     // ca.pem
		CertFile:           aopt.ClientCertPath, // cert.pem
		KeyFile:            aopt.ClientKeyPath,  // key.pem
		InsecureSkipVerify: !eopt.TLSVerify,
	}
	return tlsconfig.Client(options)
}

func (dm *DockerMachine) client(machine Machine, config MachineConfig, apiVersion string) (*client.Client, error) {
	tlsc, err := dm.tlsConfig(config)
	if err != nil {
		return nil, dm.annotateError(err, config, "tlsConfig")
	}
	httpc := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsc,
		},
	}
	host := machine.URL
	return client.NewClient(host, apiVersion, httpc, nil)
}

// Client returns docker/engine-api/client.(*Client) for the machine.
// The API version is set to client.DefaultVersion.
// You can update the version later by calling (*Client).UpdateClientVersion.
func (dm *DockerMachine) Client(name string) (*client.Client, error) {
	machine, err := dm.Machine(name)
	if err != nil {
		return nil, err
	}
	config, err := dm.Inspect(name)
	if err != nil {
		return nil, err
	}
	apiVersion := client.DefaultVersion
	return dm.client(machine, config, apiVersion)
}

// Env returns string map that contains DOCKER_{TLS_VERIFY,HOST,CERT_PATH,MACHINE_NAME}
func (dm *DockerMachine) Env(name string) (map[string]string, error) {
	env := make(map[string]string, 0)
	machine, err := dm.Machine(name)
	if err != nil {
		return nil, err
	}
	config, err := dm.Inspect(name)
	if err != nil {
		return nil, err
	}
	eopt := config.HostOptions.EngineOptions
	aopt := config.HostOptions.AuthOptions
	tlsVerify := "0"
	if eopt.TLSVerify {
		tlsVerify = "1"
	}
	env["DOCKER_TLS_VERIFY"] = tlsVerify
	env["DOCKER_HOST"] = machine.URL
	env["DOCKER_CERT_PATH"] = aopt.StorePath
	env["DOCKER_MACHINE_NAME"] = name
	return env, nil
}
