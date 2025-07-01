package charm

import (
	"fmt"

	"github.com/canonical/pebble/client"
	"github.com/gruyaume/goops"
	"gopkg.in/yaml.v3"
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

type Config struct {
	Port int `json:"port"`
}

func Configure() error {
	c := Config{}

	err := goops.GetConfig(&c)
	if err != nil {
		return fmt.Errorf("could not get config: %w", err)
	}

	if c.Port <= 0 || c.Port > 65535 {
		goops.SetUnitStatus(goops.StatusBlocked, "invalid config: port must be between 1 and 65535")
		return nil
	}

	err = goops.SetPorts([]*goops.Port{
		{Port: c.Port, Protocol: goops.ProtocolTCP},
	})
	if err != nil {
		return fmt.Errorf("could not set ports: %w", err)
	}

	pebble := goops.Pebble("myapp")

	_, err = pebble.SysInfo()
	if err != nil {
		return fmt.Errorf("could not connect to pebble: %w", err)
	}

	err = syncPebbleService(pebble)
	if err != nil {
		return fmt.Errorf("could not sync pebble service: %w", err)
	}

	goops.SetUnitStatus(goops.StatusActive, "service is running")

	return nil
}

func syncPebbleService(pebble goops.PebbleClient) error {
	if !pebbleLayerCreated(pebble) {
		goops.LogInfof("Pebble layer not created")

		err := addPebbleLayer(pebble)
		if err != nil {
			return fmt.Errorf("could not add pebble layer: %w", err)
		}

		goops.LogInfof("Pebble layer created")
	}

	_, err := pebble.Start(&client.ServiceOptions{
		Names: []string{"myapp"},
	})
	if err != nil {
		return fmt.Errorf("could not start pebble service: %w", err)
	}

	goops.LogInfof("Pebble service started")

	return nil
}

func pebbleLayerCreated(pebble goops.PebbleClient) bool {
	dataBytes, err := pebble.PlanBytes(nil)
	if err != nil {
		return false
	}

	var plan PebblePlan

	err = yaml.Unmarshal(dataBytes, &plan)
	if err != nil {
		return false
	}

	service, exists := plan.Services["myapp"]
	if !exists {
		return false
	}

	if service.Command != "myapp" {
		return false
	}

	return true
}

func addPebbleLayer(pebble goops.PebbleClient) error {
	layerData, err := yaml.Marshal(PebbleLayer{
		Summary:     "MyApp layer",
		Description: "pebble config layer for MyApp",
		Services: map[string]ServiceConfig{
			"myapp": {
				Override: "replace",
				Summary:  "My App Service",
				Command:  "myapp",
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
