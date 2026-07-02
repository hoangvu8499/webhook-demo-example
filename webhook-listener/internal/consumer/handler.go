package consumer

import (
	"context"
	"sync"

	"github.com/IBM/sarama"
	"webhook-listener/internal/service"
)

// groupHandler implements sarama.ConsumerGroupHandler.
//
// Architecture:
//   - ConsumeClaim is invoked once per assigned Kafka partition (in its own goroutine).
//   - Messages from all partitions are fed into a shared semaphore-bounded worker pool,
//     so the total concurrent HTTP calls never exceeds workerCount.
//   - Offsets are marked immediately after dispatching (at-least-once semantics).
//     Duplicate delivery is harmless because webhook_events uses dedup_key.
type groupHandler struct {
	dispatcher *service.Dispatcher
	sem        chan struct{} // bounded concurrency semaphore
	wg         sync.WaitGroup
}

func newGroupHandler(dispatcher *service.Dispatcher, workerCount int) *groupHandler {
	return &groupHandler{
		dispatcher: dispatcher,
		sem:        make(chan struct{}, workerCount),
	}
}

func (h *groupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *groupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim processes one partition's message stream concurrently via the worker pool.
func (h *groupHandler) ConsumeClaim(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim,
) error {
	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				return nil
			}

			// Mark offset right away; actual delivery happens in a goroutine.
			// On crash/restart Kafka replays from the last committed offset,
			// and the dedup_key in webhook_events prevents duplicate inserts.
			session.MarkMessage(msg, "")

			// Block here until a worker slot is free, then fire off the goroutine.
			h.sem <- struct{}{}
			h.wg.Add(1)

			payload := make([]byte, len(msg.Value))
			copy(payload, msg.Value)

			go func() {
				defer func() {
					<-h.sem
					h.wg.Done()
				}()
				h.dispatcher.Process(context.Background(), payload)
			}()

		case <-session.Context().Done():
			return nil
		}
	}
}

// drain waits for all in-flight goroutines to finish before shutdown.
func (h *groupHandler) drain() { h.wg.Wait() }
