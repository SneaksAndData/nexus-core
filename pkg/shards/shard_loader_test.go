package shards

import (
	"testing"

	"k8s.io/klog/v2"
)

func Test_LoadShards(t *testing.T) {
	shards, err := LoadShards(t.Context(), "test", "../../test-resources/kind", "nexus", klog.FromContext(t.Context()))

	if err != nil {
		t.Fatal(err)
	}

	if len(shards) != 1 {
		t.Fatal("expected 1 shard, but got ", len(shards))
	}

	if shards[0].Name != "kind-nexus-shard-0" {
		t.Fatal("expected shard name kind-nexus-shard-0, but got ", shards[0].Name)
	}
}
