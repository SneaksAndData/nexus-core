package pipeline

import (
	"golang.org/x/time/rate"
	"k8s.io/client-go/util/workqueue"
	"time"
)

type PipelineStageActor interface {
	Receive()
	Reply()
	Send()
}

type DefaultPipelineStageActor[T comparable] struct {
	queue workqueue.TypedRateLimitingInterface[T]
}

func NewDefaultPipelineStageActor[T comparable](failureRateBaseDelay time.Duration,
	failureRateMaxDelay time.Duration,
	rateLimitElementsPerSecond int,
	rateLimitElementsBurst int) *DefaultPipelineStageActor[T] {
	ratelimiter := workqueue.NewTypedMaxOfRateLimiter(
		workqueue.NewTypedItemExponentialFailureRateLimiter[T](failureRateBaseDelay, failureRateMaxDelay),
		&workqueue.TypedBucketRateLimiter[T]{Limiter: rate.NewLimiter(rate.Limit(rateLimitElementsPerSecond), rateLimitElementsBurst)},
	)

	return &DefaultPipelineStageActor[T]{
		queue: workqueue.NewTypedRateLimitingQueue(ratelimiter),
	}
}

func (a *DefaultPipelineStageActor[T]) Receive() {}
func (a *DefaultPipelineStageActor[T]) Reply()   {}
func (a *DefaultPipelineStageActor[T]) Send()    {}
