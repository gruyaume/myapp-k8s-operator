package charm

import (
	"fmt"
	"strings"

	"github.com/canonical/pebble/client"
	"github.com/gruyaume/goops"
	"gopkg.in/yaml.v3"
)

const (
	Port       = 8080
	ConfigPath = "/etc/myapp/config.yaml"
)

type ServiceConfig struct {
	Override string `yaml:"override"`
	Summary  string `yaml:"summary"`
	Command  string `yaml:"command"`
	Startup  string `yaml:"startup"`
}

type PebbleLayer struct {
	Summary     string                   `yaml:"summary"`
	Description string                   `yaml:"description"`
	Services    map[string]ServiceConfig `yaml:"services"`
}

type PebblePlan struct {
	Services map[string]ServiceConfig `yaml:"services"`
}

func Configure() error {
	pebble := goops.Pebble("myapp")

	err := goops.SetPorts([]*goops.Port{
		{Port: Port, Protocol: goops.ProtocolTCP},
	})
	if err != nil {
		return fmt.Errorf("could not set ports: %w", err)
	}

	err = syncConfig(pebble)
	if err != nil {
		return fmt.Errorf("could not sync config: %w", err)
	}

	_, err = pebble.SysInfo()
	if err != nil {
		return fmt.Errorf("could not connect to pebble: %w", err)
	}

	err = syncPebbleService(pebble)
	if err != nil {
		return fmt.Errorf("could not sync pebble service: %w", err)
	}

	_ = goops.SetUnitStatus(goops.StatusActive, "service is running")

	return nil
}

type MyAppConfig struct {
	Port int `yaml:"port"`
}

func getExpectedConfig() ([]byte, error) {
	myappConfig := MyAppConfig{
		Port: Port,
	}

	b, err := yaml.Marshal(myappConfig)
	if err != nil {
		return nil, fmt.Errorf("could not marshal config to YAML: %w", err)
	}

	return b, nil
}

func syncConfig(pebble goops.PebbleClient) error {
	content, err := getExpectedConfig()
	if err != nil {
		return fmt.Errorf("could not get expected config: %w", err)
	}

	source := strings.NewReader(string(content))

	err = pebble.Push(&client.PushOptions{
		Source: source,
		Path:   ConfigPath,
	})

	goops.LogInfof("Config file pushed to %s", ConfigPath)

	return nil
}

func syncPebbleService(pebble goops.PebbleClient) error {
	err := addPebbleLayer(pebble)
	if err != nil {
		return fmt.Errorf("could not add pebble layer: %w", err)
	}

	goops.LogInfof("Pebble layer created")

	_, err = pebble.Start(&client.ServiceOptions{
		Names: []string{"myapp"},
	})
	if err != nil {
		return fmt.Errorf("could not start pebble service: %w", err)
	}

	goops.LogInfof("Pebble service started")

	return nil
}

func addPebbleLayer(pebble goops.PebbleClient) error {
	layerData, err := yaml.Marshal(PebbleLayer{
		Summary:     "MyApp layer",
		Description: "pebble config layer for MyApp",
		Services: map[string]ServiceConfig{
			"myapp": {
				Override: "replace",
				Summary:  "My App Service",
				Command:  "myapp -config /etc/myapp/config.yaml",
				Startup:  "enabled",
			},
		},
	})
	if err != nil {
		return fmt.Errorf("could not marshal layer data to YAML: %w", err)
	}

	err = pebble.AddLayer(&client.AddLayerOptions{
		Combine:   true,
		Label:     "myapp",
		LayerData: layerData,
	})
	if err != nil {
		return fmt.Errorf("could not add pebble layer: %w", err)
	}

	return nil
}
