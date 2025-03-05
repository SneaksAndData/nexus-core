package shards

import (
	"context"
	clientset "github.com/SneaksAndData/nexus-core/pkg/generated/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"os"
	"path"
	"strings"
)

func LoadClients(shardConfigPath string, logger klog.Logger) ([]*ShardClient, error) {
	files, err := os.ReadDir(shardConfigPath)
	if err != nil {
		logger.Error(err, "Error opening kubeconfig files for Shards")
		return nil, err
	}
	shardClients := make([]*ShardClient, 0, len(files))

	// only load kubeconfig files in the provided location
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".kubeconfig") {
			logger.Info("Loading Shard kubeconfig file", "file", file.Name())

			cfg, err := clientcmd.BuildConfigFromFlags("", path.Join(shardConfigPath, file.Name()))
			if err != nil {
				logger.Error(err, "Error building kubeconfig for shard {shard}", file.Name())
				return nil, err
			}

			kubeClient, err := kubernetes.NewForConfig(cfg)
			if err != nil {
				logger.Error(err, "Error building kubernetes clientset for shard {shard}", file.Name())
				return nil, err
			}

			nexusClient, err := clientset.NewForConfig(cfg)
			if err != nil {
				logger.Error(err, "Error building kubernetes clientset for MachineLearningAlgorithm API for shard {shard}", file.Name())
				return nil, err
			}

			shardClients = append(shardClients, NewShardClient(kubeClient, nexusClient, strings.Split(file.Name(), ".")[0]))
		}
	}

	return shardClients, nil
}

func LoadShards(ctx context.Context, owner string, shardConfigPath string, logger klog.Logger) ([]*Shard, error) {
	shardClients, err := LoadClients(shardConfigPath, logger)
	connectedShards := []*Shard{}

	if err != nil {
		return nil, err
	}

	for _, shardClient := range shardClients {
		connectedShards = append(connectedShards, shardClient.ToShard(owner, ctx))
	}

	return connectedShards, nil
}
