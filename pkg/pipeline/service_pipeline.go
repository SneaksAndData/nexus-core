package pipeline

import (
	"context"
	"fmt"
	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/SneaksAndData/nexus-core/pkg/telemetry"
	"golang.org/x/time/rate"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"time"
)

type ActorElementProcessor[TIn comparable, TOut comparable] func(element TIn) (TOut, error)
type ActorPostStart func(ctx context.Context) error

func NewActorPostStart(postStart func(ctx context.Context) error) *ActorPostStart {
	var result ActorPostStart
	result = postStart
	return &result
}

type StageActor[TIn comparable, TOut comparable] interface {
	Receive(element TIn)
}

type DefaultPipelineStageActor[TIn comparable, TOut comparable] struct {
	stageName string
	stageTags map[string]string
	queue     workqueue.TypedRateLimitingInterface[TIn]
	workers   int
	processor ActorElementProcessor[TIn, TOut]
	receiver  StageActor[TOut, any]
}

func NewDefaultPipelineStageActor[TIn comparable, TOut comparable](actorName string,
	actorTags map[string]string,
	failureRateBaseDelay time.Duration,
	failureRateMaxDelay time.Duration,
	rateLimitElementsPerSecond int,
	rateLimitElementsBurst int,
	queueWorkers int,
	processor ActorElementProcessor[TIn, TOut],
	receiver StageActor[TOut, any]) *DefaultPipelineStageActor[TIn, TOut] {
	rateLimiter := workqueue.NewTypedMaxOfRateLimiter(
		workqueue.NewTypedItemExponentialFailureRateLimiter[TIn](failureRateBaseDelay, failureRateMaxDelay),
		&workqueue.TypedBucketRateLimiter[TIn]{Limiter: rate.NewLimiter(rate.Limit(rateLimitElementsPerSecond), rateLimitElementsBurst)},
	)

	return &DefaultPipelineStageActor[TIn, TOut]{
		stageName: actorName,
		stageTags: actorTags,
		queue:     workqueue.NewTypedRateLimitingQueue(rateLimiter),
		workers:   queueWorkers,
		processor: processor,
		receiver:  receiver,
	}
}

func (a *DefaultPipelineStageActor[TIn, TOut]) processNextElement(ctx context.Context, logger klog.Logger, metrics *statsd.Client) bool {
	element, shutdown := a.queue.Get()
	logger.V(0).Info("starting processing element", "element", element, "stage", a.stageName)
	elementProcessStart := time.Now()

	if shutdown {
		return false
	}

	// We call Done at the end of this func so the queue knows we have
	// finished processing this item. We also must remember to call Forget
	// if we do not want this work item being re-queued

	defer a.queue.Forget(element)
	defer a.queue.Done(element)
	defer telemetry.GaugeDuration(metrics, fmt.Sprintf("%s_processing_latency", a.stageName), elementProcessStart, a.stageTags, 1)
	defer telemetry.Gauge(metrics, fmt.Sprintf("%s_queue_size", a.stageName), float64(a.queue.Len()), a.stageTags, 1)

	result, err := a.processor(element)
	logger.V(0).Info("finished processing element", "element", element, "stage", a.stageName)
	if err == nil {
		// If no error occurs then we send the result to the receiver, if one is attached
		if a.receiver != nil {
			a.receiver.Receive(result)
		}
		return true
	}
	// there was a failure so be sure to report it.  This method allows for
	// pluggable error handling which can be used for things like cluster-monitoring.
	utilruntime.HandleErrorWithContext(ctx, err, "error processing element", "element", element, "stage", a.stageName)
	logger.V(0).Error(err, "error when processing element", "element", element, "stage", a.stageName)

	// forget this submission to prevent clogging the queue
	a.queue.Forget(element)
	return true
}

func (a *DefaultPipelineStageActor[TIn, TOut]) runActor(ctx context.Context) {
	for a.processNextElement(ctx, klog.FromContext(ctx), telemetry.GetClient(ctx)) {
	}
}

func (a *DefaultPipelineStageActor[TIn, TOut]) Start(ctx context.Context, postStart *ActorPostStart) {
	defer utilruntime.HandleCrash()
	defer a.queue.ShutDown()

	logger := klog.FromContext(ctx)

	logger.V(0).Info("starting workers", "stage", a.stageName)
	for i := 0; i < a.workers; i++ {
		go wait.UntilWithContext(ctx, a.runActor, time.Second)
	}
	logger.V(0).Info("started workers", "workers", a.workers, "stage", a.stageName)

	if postStart != nil { // coverage-ignore
		postStartErr := (*postStart)(ctx)
		if postStartErr != nil {
			logger.V(0).Error(postStartErr, "failed to execute actor post-start", "stage", a.stageName)
			klog.FlushAndExit(klog.ExitFlushTimeout, 1)
		}
	}

	<-ctx.Done()

	logger.V(0).Info("shutting down workers for stage", "stage", a.stageName)
}

func (a *DefaultPipelineStageActor[TIn, TOut]) Receive(element TIn) {
	a.queue.AddRateLimited(element)
}
