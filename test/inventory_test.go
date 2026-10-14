package test

import (
	"net"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type inventoryHost struct {
	AnsibleHost string `yaml:"ansible_host"`
}

// inventory vars are deliberately untyped: control_plane_endpoint and
// load_balancer_ip_pool may reference other vars in this same block via
// Ansible/Jinja templates (e.g. "{{ metal_subnet_prefix }}.239"), so they're
// resolved with resolveTemplate before being parsed as addresses.
type inventory struct {
	All struct {
		Vars map[string]interface{} `yaml:"vars"`
	} `yaml:"all"`
	Metal struct {
		Children struct {
			Masters struct {
				Hosts map[string]inventoryHost `yaml:"hosts"`
			} `yaml:"masters"`
			Workers struct {
				Hosts map[string]inventoryHost `yaml:"hosts"`
			} `yaml:"workers"`
		} `yaml:"children"`
	} `yaml:"metal"`
}

// subnet24 returns the /24 network an IP belongs to, as a string like "192.168.0".
func subnet24(ip net.IP) string {
	ip4 := ip.To4()
	return net.IPv4(ip4[0], ip4[1], ip4[2], 0).String()
}

var templateVarRe = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_]+)\s*\}\}`)

// resolveTemplate expands simple "{{ var_name }}" references against other
// scalar vars declared in the same inventory's `all.vars` block. It doesn't
// implement Jinja -- just enough to let values like control_plane_endpoint
// derive from a single shared prefix var instead of repeating it.
func resolveTemplate(s string, vars map[string]string) string {
	return templateVarRe.ReplaceAllStringFunc(s, func(match string) string {
		key := templateVarRe.FindStringSubmatch(match)[1]
		if v, ok := vars[key]; ok {
			return v
		}
		return match
	})
}

// TestInventorySubnetConsistency makes sure every address configured for a given
// metal inventory (host addresses, the kube-vip control plane endpoint, and the
// load balancer IP pool) lives on the same /24 subnet. kube-vip advertises the
// control plane endpoint and LB pool over ARP, which only reaches nodes on the
// same L2 segment -- a mismatch here means the cluster will fail to come up
// even though every value individually looks like a valid IP/CIDR.
func TestInventorySubnetConsistency(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob("../metal/inventories/*.yml")
	require.NoError(t, err)
	require.NotEmpty(t, files, "expected to find at least one inventory file")

	for _, file := range files {
		file := file
		t.Run(filepath.Base(file), func(t *testing.T) {
			t.Parallel()

			var inv inventory
			raw, err := os.ReadFile(file)
			require.NoError(t, err)
			require.NoError(t, yaml.Unmarshal(raw, &inv))

			var hostSubnets []string
			var hostDescriptions []string
			for name, host := range inv.Metal.Children.Masters.Hosts {
				ip := net.ParseIP(host.AnsibleHost)
				require.NotNilf(t, ip, "master %q has an invalid ansible_host %q", name, host.AnsibleHost)
				hostSubnets = append(hostSubnets, subnet24(ip))
				hostDescriptions = append(hostDescriptions, name+"="+host.AnsibleHost)
			}
			for name, host := range inv.Metal.Children.Workers.Hosts {
				ip := net.ParseIP(host.AnsibleHost)
				require.NotNilf(t, ip, "worker %q has an invalid ansible_host %q", name, host.AnsibleHost)
				hostSubnets = append(hostSubnets, subnet24(ip))
				hostDescriptions = append(hostDescriptions, name+"="+host.AnsibleHost)
			}

			if len(hostSubnets) == 0 {
				t.Skip("no hosts defined, nothing to check")
			}

			expectedSubnet := hostSubnets[0]
			for i, s := range hostSubnets {
				require.Equalf(t, expectedSubnet, s,
					"all metal hosts must be on the same /24 subnet, got: %v", hostDescriptions[i])
			}

			scalarVars := map[string]string{}
			for k, v := range inv.All.Vars {
				if s, ok := v.(string); ok {
					scalarVars[k] = s
				}
			}

			if raw, ok := inv.All.Vars["control_plane_endpoint"].(string); ok && raw != "" {
				endpoint := resolveTemplate(raw, scalarVars)
				ip := net.ParseIP(endpoint)
				require.NotNilf(t, ip, "control_plane_endpoint %q (raw: %q) is not a valid IP", endpoint, raw)
				require.Equalf(t, expectedSubnet, subnet24(ip),
					"control_plane_endpoint %s must be on the same /24 subnet as the metal hosts (%s.0/24)", endpoint, expectedSubnet)
			}

			if rawPool, ok := inv.All.Vars["load_balancer_ip_pool"].([]interface{}); ok {
				for _, entry := range rawPool {
					raw, _ := entry.(string)
					cidr := resolveTemplate(raw, scalarVars)
					ip, _, err := net.ParseCIDR(cidr)
					require.NoErrorf(t, err, "load_balancer_ip_pool entry %q (raw: %q) is not a valid CIDR", cidr, raw)
					require.Equalf(t, expectedSubnet, subnet24(ip),
						"load_balancer_ip_pool entry %s must be on the same /24 subnet as the metal hosts (%s.0/24)", cidr, expectedSubnet)
				}
			}
		})
	}
}
