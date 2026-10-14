package test

import (
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type inventoryHost struct {
	AnsibleHost string `yaml:"ansible_host"`
}

type inventory struct {
	All struct {
		Vars struct {
			ControlPlaneEndpoint string   `yaml:"control_plane_endpoint"`
			LoadBalancerIPPool   []string `yaml:"load_balancer_ip_pool"`
		} `yaml:"vars"`
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

			if endpoint := inv.All.Vars.ControlPlaneEndpoint; endpoint != "" {
				ip := net.ParseIP(endpoint)
				require.NotNilf(t, ip, "control_plane_endpoint %q is not a valid IP", endpoint)
				require.Equalf(t, expectedSubnet, subnet24(ip),
					"control_plane_endpoint %s must be on the same /24 subnet as the metal hosts (%s.0/24)", endpoint, expectedSubnet)
			}

			for _, cidr := range inv.All.Vars.LoadBalancerIPPool {
				ip, _, err := net.ParseCIDR(cidr)
				require.NoErrorf(t, err, "load_balancer_ip_pool entry %q is not a valid CIDR", cidr)
				require.Equalf(t, expectedSubnet, subnet24(ip),
					"load_balancer_ip_pool entry %s must be on the same /24 subnet as the metal hosts (%s.0/24)", cidr, expectedSubnet)
			}
		})
	}
}
