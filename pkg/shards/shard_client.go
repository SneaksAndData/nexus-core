package shards

import (
	"context"
	clientset "github.com/SneaksAndData/nexus-core/pkg/generated/clientset/versioned"
	informers "github.com/SneaksAndData/nexus-core/pkg/generated/informers/externalversions"
	"github.com/aws/smithy-go/ptr"
	"github.com/hashicorp/golang-lru/v2"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"time"
)

type ShardClient struct {
	Name                string
	Namespace           string
	kubernetesClientSet kubernetes.Interface
	nexusClientSet      clientset.Interface
	parentJobCache      *lru.Cache[string, batchv1.Job]
}

func NewShardClient(kubernetesClientSet kubernetes.Interface, nexusClientSet clientset.Interface, name string, namespace string, logger klog.Logger) *ShardClient {
	lruCache, err := lru.New[string, batchv1.Job](1000)

	if err != nil {
		logger.V(0).Error(err, "failed to provision a new LRU cache for shard client '%s'", name)
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	return &ShardClient{
		Name:                name,
		Namespace:           namespace,
		kubernetesClientSet: kubernetesClientSet,
		nexusClientSet:      nexusClientSet,
		parentJobCache:      lruCache,
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

func (c *ShardClient) DeleteJob(namespace string, jobName string, policy v1.DeletionPropagation) error {
	return c.kubernetesClientSet.BatchV1().Jobs(namespace).Delete(context.TODO(), jobName, v1.DeleteOptions{
		PropagationPolicy:  &policy,
		GracePeriodSeconds: ptr.Int64(0),
	})
}

func (c *ShardClient) FindJob(jobName string, namespace string) (*batchv1.Job, error) {
	if job, ok := c.parentJobCache.Get(jobName); !ok {
		jobInfo, err := c.kubernetesClientSet.BatchV1().Jobs(namespace).Get(context.TODO(), jobName, v1.GetOptions{})

		if err != nil {
			return nil, err
		}

		c.parentJobCache.Add(jobName, *jobInfo)

		return jobInfo, nil
	} else {
		return &job, nil
	}
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
		kubeInformerFactory.Core().V1().ConfigMaps())
}
