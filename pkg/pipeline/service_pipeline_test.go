package pipeline

import (
	"context"
	"fmt"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/ktesting"
	"testing"
	"time"
)

var (
	counter = make(chan int)
)

type testElement struct {
	prop1 string
	prop2 int
}

func Test_DefaultPipelineStageActor(t *testing.T) {
	_, ctx := ktesting.NewTestContext(t)

	receiver := NewDefaultPipelineStageActor(
		"receiver",
		map[string]string{},
		time.Second,
		2*time.Second,
		10,
		10,
		2,
		func(element int) (string, error) {
			return countElement(element)
		},
		nil,
	)
	sender := NewDefaultPipelineStageActor(
		"sender",
		map[string]string{},
		time.Second,
		2*time.Second,
		10,
		10,
		2,
		func(element *testElement) (int, error) {
			return sendElement(ctx, element)
		},
		receiver)

	go sender.Start(ctx)
	go receiver.Start(ctx)

	total := 10
	expected := 0
	counted := 0

	for i := 0; i < total; i++ {
		sender.Receive(&testElement{
			prop1: fmt.Sprintf("test-%d", i),
			prop2: i,
		})
		expected += i
	}

	// ensure we finish processing before we close the channel by sending a termination
	time.Sleep(1 * time.Second)

	// send a terminating element
	sender.Receive(&testElement{
		prop1: fmt.Sprintf("test-%d", total),
		prop2: -1,
	})

	for index := range counter {
		counted += index
	}

	ctx.Done()

	if expected != counted {
		t.Errorf("Incorrect behaviour observed for the pipeline: expected %d for a cumulative sum, observed %d as a result", expected, counted)
	}
}

func countElement(index int) (string, error) {
	if index < 0 {
		close(counter)
	} else {
		counter <- index
	}

	return "done", nil
}

func sendElement(ctx context.Context, element *testElement) (int, error) {
	logger := klog.FromContext(ctx)
	logger.Info("Processing element with prop1=%s and prop2=%d", element.prop1, element.prop2)

	return element.prop2, nil
}
