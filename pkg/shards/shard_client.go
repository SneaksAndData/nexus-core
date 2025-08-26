package shards

import (
	"context"
	clientset "github.com/SneaksAndData/nexus-core/pkg/generated/clientset/versioned"
	informers "github.com/SneaksAndData/nexus-core/pkg/generated/informers/externalversions"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"time"
)

type ShardClient struct {
	Name                string
	Namespace           string
	kubernetesClientSet kubernetes.Interface
	nexusClientSet      clientset.Interface
}

func NewShardClient(kubernetesClientSet kubernetes.Interface, nexusClientsetSet clientset.Interface, name string, namespace string) *ShardClient {
	return &ShardClient{
		Name:                name,
		Namespace:           namespace,
		kubernetesClientSet: kubernetesClientSet,
		nexusClientSet:      nexusClientsetSet,
	}
}

func (c *ShardClient) getKubeInformerFactory(namespace string) kubeinformers.SharedInformerFactory {
	return kubeinformers.NewSharedInformerFactoryWithOptions(c.kubernetesClientSet, time.Second*30, kubeinformers.WithNamespace(namespace))
}

func (c *ShardClient) getNexusInformerFactory(namespace string) informers.SharedInformerFactory {
	return informers.NewSharedInformerFactoryWithOptions(c.nexusClientSet, time.Second*30, informers.WithNamespace(namespace))
}

func (c *ShardClient) SendJob(namespace string, job *batchv1.Job) (*batchv1.Job, error) {
	return c.kubernetesClientSet.BatchV1().Jobs(namespace).Create(context.TODO(), job, v1.CreateOptions{})
}

func (c *ShardClient) ToShard(owner string, ctx context.Context) *Shard {
	kubeInformerFactory := c.getKubeInformerFactory(c.Name)
	nexusInformerFactory := c.getNexusInformerFactory(c.Name)

	defer kubeInformerFactory.Start(ctx.Done())
	defer nexusInformerFactory.Start(ctx.Done())

	return NewShard(
		owner,
		c.Name,
		c.kubernetesClientSet,
		c.nexusClientSet,
		nexusInformerFactory.Science().V1().NexusAlgorithmTemplates(),
		nexusInformerFactory.Science().V1().NexusAlgorithmWorkgroups(),
		kubeInformerFactory.Core().V1().Secrets(),
		kubeInformerFactory.Core().V1().ConfigMaps(),
		kubeInformerFactory.Batch().V1().Jobs().Informer())
}
