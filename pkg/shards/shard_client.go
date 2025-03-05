package shards

import (
	"context"
	clientset "github.com/SneaksAndData/nexus-core/pkg/generated/clientset/versioned"
	informers "github.com/SneaksAndData/nexus-core/pkg/generated/informers/externalversions"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"time"
)

type ShardClient struct {
	name                string
	kubernetesclientset kubernetes.Interface
	nexusclientset      clientset.Interface
}

func NewShardClient(kubernetesClientset kubernetes.Interface, nexusclientsetset clientset.Interface, name string) *ShardClient {
	return &ShardClient{
		name:                name,
		kubernetesclientset: kubernetesClientset,
		nexusclientset:      nexusclientsetset,
	}
}

func (c *ShardClient) getKubeInformerFactory(namespace string) kubeinformers.SharedInformerFactory {
	return kubeinformers.NewSharedInformerFactoryWithOptions(c.kubernetesclientset, time.Second*30, kubeinformers.WithNamespace(namespace))
}

func (c *ShardClient) getNexusInformerFactory(namespace string) informers.SharedInformerFactory {
	return informers.NewSharedInformerFactoryWithOptions(c.nexusclientset, time.Second*30, informers.WithNamespace(namespace))
}

func (c *ShardClient) ToShard(owner string, ctx context.Context) *Shard {
	kubeInformer := c.getKubeInformerFactory(c.name)
	nexusInformer := c.getNexusInformerFactory(c.name)

	defer kubeInformer.Start(ctx.Done())
	defer nexusInformer.Start(ctx.Done())

	return NewShard(
		owner,
		c.name,
		c.kubernetesclientset,
		c.nexusclientset,
		nexusInformer.Science().V1().MachineLearningAlgorithms(),
		kubeInformer.Core().V1().Secrets(),
		kubeInformer.Core().V1().ConfigMaps())
}
