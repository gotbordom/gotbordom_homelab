package test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestGPUNodeSelectorConsistency guards the one string ("gpu=true" / gpu:
// "true") that k3s's node-label, the nvidia-device-plugin chart, and
// apps/ollama all have to independently agree on for GPU scheduling to work
// at all. There's no schema tying these together -- a typo or a later rename
// in just one of them would silently leave a pod Pending forever.
func TestGPUNodeSelectorConsistency(t *testing.T) {
	t.Parallel()

	k3sTemplate, err := os.ReadFile("../metal/roles/k3s/templates/config.yaml.j2")
	require.NoError(t, err)
	require.Containsf(t, string(k3sTemplate), `"gpu=true"`,
		"metal/roles/k3s/templates/config.yaml.j2 no longer sets the gpu=true node-label")

	type nodeSelectorValues struct {
		NvidiaDevicePlugin struct {
			NodeSelector map[string]string `yaml:"nodeSelector"`
		} `yaml:"nvidia-device-plugin"`
	}
	raw, err := os.ReadFile("../system/nvidia-device-plugin/values.yaml")
	require.NoError(t, err)
	var ndp nodeSelectorValues
	require.NoError(t, yaml.Unmarshal(raw, &ndp))
	require.Equalf(t, "true", ndp.NvidiaDevicePlugin.NodeSelector["gpu"],
		"system/nvidia-device-plugin/values.yaml's nodeSelector no longer matches the gpu=true node-label")

	type ollamaValues struct {
		AppTemplate struct {
			Controllers struct {
				Main struct {
					Pod struct {
						NodeSelector map[string]string `yaml:"nodeSelector"`
					} `yaml:"pod"`
				} `yaml:"main"`
			} `yaml:"controllers"`
		} `yaml:"app-template"`
	}
	raw, err = os.ReadFile("../apps/ollama/values.yaml")
	require.NoError(t, err)
	var ollama ollamaValues
	require.NoError(t, yaml.Unmarshal(raw, &ollama))
	require.Equalf(t, "true", ollama.AppTemplate.Controllers.Main.Pod.NodeSelector["gpu"],
		"apps/ollama/values.yaml's main controller nodeSelector no longer matches the gpu=true node-label")
}
