package test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type cephPool struct {
	Spec struct {
		Replicated struct {
			Size int `yaml:"size"`
		} `yaml:"replicated"`
	} `yaml:"spec"`
}

type cephFileSystem struct {
	Spec struct {
		MetadataPool struct {
			Replicated struct {
				Size int `yaml:"size"`
			} `yaml:"replicated"`
		} `yaml:"metadataPool"`
		DataPools []struct {
			Replicated struct {
				Size int `yaml:"size"`
			} `yaml:"replicated"`
		} `yaml:"dataPools"`
	} `yaml:"spec"`
}

type rookCephValues struct {
	RookCephCluster struct {
		CephClusterSpec struct {
			Mon struct {
				Count int `yaml:"count"`
			} `yaml:"mon"`
		} `yaml:"cephClusterSpec"`
		CephBlockPools  []cephPool       `yaml:"cephBlockPools"`
		CephFileSystems []cephFileSystem `yaml:"cephFileSystems"`
	} `yaml:"rook-ceph-cluster"`
}

// TestCephReplicationFitsNodeCount checks that every declared Ceph replication
// factor (block pools and filesystem pools) is achievable with the number of
// nodes actually defined in the production metal inventory. This is a static
// check only -- it can't see whether those nodes actually have free disks or
// are healthy, only that the desired-state numbers aren't self-contradictory.
// This is exactly the gap that let a single-active-node inventory sit next to
// a replicated-size-2 pool with nowhere to put the second copy.
//
// Checked against prod.yml specifically, not every inventory: stag.yml is
// intentionally a single-node environment (see
// docs/how-to-guides/single-node-cluster-adjustments.md, which tells users to
// manually drop replication to 1 for that case), so this invariant doesn't
// apply there.
func TestCephReplicationFitsNodeCount(t *testing.T) {
	t.Parallel()

	inv, err := loadInventory("../metal/inventories/prod.yml")
	require.NoError(t, err)
	nodeCount := inv.metalHostCount()
	require.Greaterf(t, nodeCount, 0, "expected metal/inventories/prod.yml to define at least one host")

	raw, err := os.ReadFile("../system/rook-ceph/values.yaml")
	require.NoError(t, err)
	var values rookCephValues
	require.NoError(t, yaml.Unmarshal(raw, &values))

	for _, pool := range values.RookCephCluster.CephBlockPools {
		require.LessOrEqualf(t, pool.Spec.Replicated.Size, nodeCount,
			"a cephBlockPool asks for replicated.size %d, but prod.yml only defines %d node(s)",
			pool.Spec.Replicated.Size, nodeCount)
	}

	for _, fs := range values.RookCephCluster.CephFileSystems {
		require.LessOrEqualf(t, fs.Spec.MetadataPool.Replicated.Size, nodeCount,
			"a cephFileSystem metadataPool asks for replicated.size %d, but prod.yml only defines %d node(s)",
			fs.Spec.MetadataPool.Replicated.Size, nodeCount)
		for _, dataPool := range fs.Spec.DataPools {
			require.LessOrEqualf(t, dataPool.Replicated.Size, nodeCount,
				"a cephFileSystem dataPool asks for replicated.size %d, but prod.yml only defines %d node(s)",
				dataPool.Replicated.Size, nodeCount)
		}
	}
}

// TestCephMonCountFitsNodeCount checks that the desired mon count doesn't
// exceed the number of nodes available to place monitors on -- Ceph mons need
// one-per-node for quorum, so asking for more mons than you have nodes is a
// desired-state value that can never actually be satisfied.
func TestCephMonCountFitsNodeCount(t *testing.T) {
	t.Parallel()

	inv, err := loadInventory("../metal/inventories/prod.yml")
	require.NoError(t, err)
	nodeCount := inv.metalHostCount()

	raw, err := os.ReadFile("../system/rook-ceph/values.yaml")
	require.NoError(t, err)
	var values rookCephValues
	require.NoError(t, yaml.Unmarshal(raw, &values))

	monCount := values.RookCephCluster.CephClusterSpec.Mon.Count
	require.LessOrEqualf(t, monCount, nodeCount,
		"cephClusterSpec.mon.count is %d, but prod.yml only defines %d node(s)", monCount, nodeCount)
}
