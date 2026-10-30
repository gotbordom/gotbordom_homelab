package test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestGrafanaRoleMappingConfigured guards against platform/grafana/values.yaml
// losing its role_attribute_path -- without it, every user who can log in
// gets the same Grafana org role regardless of Kanidm tier, silently undoing
// the point of having separate admin/editor/viewer groups.
func TestGrafanaRoleMappingConfigured(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile("../platform/grafana/values.yaml")
	require.NoError(t, err)

	var values struct {
		Grafana struct {
			GrafanaIni struct {
				AuthGenericOauth struct {
					RoleAttributePath string `yaml:"role_attribute_path"`
				} `yaml:"auth.generic_oauth"`
			} `yaml:"grafana.ini"`
		} `yaml:"grafana"`
	}
	require.NoError(t, yaml.Unmarshal(raw, &values))

	require.NotEmptyf(t, values.Grafana.GrafanaIni.AuthGenericOauth.RoleAttributePath,
		"platform/grafana/values.yaml must set auth.generic_oauth.role_attribute_path, "+
			"otherwise Grafana can't tell admin/editor/viewer apart")
}

// TestKanidmPermissionTiersDefined is a light text-level guard that
// scripts/hacks still provisions all three access tiers (homelab_admins,
// homelab_editors, homelab_viewers) as Kanidm groups scope-mapped onto the dex oauth2 client. It's not
// a real parse of the Python source -- just enough to catch someone
// accidentally dropping a tier back to the old single-"editor" setup.
func TestKanidmPermissionTiersDefined(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile("../scripts/hacks")
	require.NoError(t, err)
	content := string(raw)

	for _, tier := range []string{"homelab_admins", "homelab_editors", "homelab_viewers"} {
		require.Containsf(t, content, "'"+tier+"'",
			"scripts/hacks no longer appears to define the %q tier", tier)
	}
}
