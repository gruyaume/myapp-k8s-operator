package charm_test

import (
	"myapp-k8s-operator/internal/charm"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/gruyaume/goops/goopstest"
)

func TestGivenBadPortConfigWhenAnyEventThenStatusBlocked(t *testing.T) {
	ctx := goopstest.NewContext(
		charm.Configure,
	)
	stateIn := goopstest.State{
		Config: map[string]any{
			"port": 0, // Invalid port
		},
		Containers: []goopstest.Container{
			{
				Name:       "myapp",
				CanConnect: true,
			},
		},
	}

	stateOut := ctx.Run("update-status", stateIn)

	expectedStatus := goopstest.Status{
		Name:    goopstest.StatusBlocked,
		Message: "invalid config: port must be between 1 and 65535",
	}
	if stateOut.UnitStatus != expectedStatus {
		t.Errorf("expected status %v, got %v", expectedStatus, stateOut.UnitStatus)
	}
}

func TestGivenValidConfigWhenAnyEventThenStatusActive(t *testing.T) {
	ctx := goopstest.NewContext(
		charm.Configure,
	)
	stateIn := goopstest.State{
		Config: map[string]any{
			"port": 8080, // Valid port
		},
		Containers: []goopstest.Container{
			{
				Name:       "myapp",
				CanConnect: true,
			},
		},
	}

	stateOut := ctx.Run("update-status", stateIn)

	expectedStatus := goopstest.Status{
		Name:    goopstest.StatusActive,
		Message: "service is running on port 8080",
	}
	if stateOut.UnitStatus != expectedStatus {
		t.Errorf("expected status %v, got %v", expectedStatus, stateOut.UnitStatus)
	}
}

func TestGivenValidConfigWhenAnyEventThenPebbleLayerIsAdded(t *testing.T) {
	ctx := goopstest.NewContext(
		charm.Configure,
	)
	stateIn := goopstest.State{
		Config: map[string]any{
			"port": 8080, // Valid port
		},
		Containers: []goopstest.Container{
			{
				Name:       "myapp",
				CanConnect: true,
			},
		},
	}

	stateOut := ctx.Run("update-status", stateIn)

	got := stateOut.Containers[0].Layers["myapp"]

	want := goopstest.Layer{
		Summary:     "MyApp layer",
		Description: "pebble config layer for MyApp",
		Services: map[string]goopstest.Service{
			"myapp": {
				Summary:  "My App Service",
				Command:  "myapp -config /etc/myapp/config.yaml",
				Startup:  "enabled",
				Override: "replace",
			},
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected pebble layer (-want +got):\n%s", diff)
	}
}
