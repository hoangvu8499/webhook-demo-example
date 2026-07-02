package consumer

import (
	"context"
	"log"
	"time"

	"github.com/IBM/sarama"
	"webhook-listener/internal/config"
	"webhook-listener/internal/service"
)

// KafkaConsumer wraps a sarama ConsumerGroup with graceful shutdown.
type KafkaConsumer struct {
	group   sarama.ConsumerGroup
	topic   string
	handler *groupHandler
}

// NewKafkaConsumer configures and starts a Kafka ConsumerGroup.
//
// Key settings:
//   - BalanceStrategyRoundRobin: partitions assigned to consumers in round-robin order.
//   - OffsetOldest: starts from the beginning if no committed offset exists.
//   - Auto-commit every 1 s to keep offset lag low.
func NewKafkaConsumer(cfg *config.Config, dispatcher *service.Dispatcher) (*KafkaConsumer, error) {
	scfg := sarama.NewConfig()
	scfg.Version = sarama.MaxVersion

	// Round-robin partition assignment across consumer group members
	scfg.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{
		sarama.BalanceStrategyRoundRobin,
	}
	scfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	scfg.Consumer.Offsets.AutoCommit.Enable = true
	scfg.Consumer.Offsets.AutoCommit.Interval = 1 * time.Second
	scfg.Consumer.Return.Errors = true

	scfg.Net.DialTimeout = 10 * time.Second
	scfg.Net.ReadTimeout = 30 * time.Second
	scfg.Net.WriteTimeout = 30 * time.Second

	// Producer-side settings not needed; set consumer fetch sizes for throughput.
	scfg.Consumer.Fetch.Min = 1
	scfg.Consumer.Fetch.Default = 1 << 20 // 1 MB per partition fetch

	group, err := sarama.NewConsumerGroup(cfg.KafkaBrokers, cfg.KafkaGroupID, scfg)
	if err != nil {
		return nil, err
	}

	handler := newGroupHandler(dispatcher, cfg.WorkerCount)
	return &KafkaConsumer{group: group, topic: cfg.KafkaTopic, handler: handler}, nil
}

// Run starts the consumer loop. It re-enters Consume() after each rebalance and
// exits only when ctx is cancelled.
func (c *KafkaConsumer) Run(ctx context.Context) {
	// Log consumer-group errors in background.
	go func() {
		for err := range c.group.Errors() {
			log.Printf("[KAFKA-ERR] %v", err)
		}
	}()

	for {
		if err := c.group.Consume(ctx, []string{c.topic}, c.handler); err != nil {
			log.Printf("[KAFKA] consume error: %v", err)
		}
		if ctx.Err() != nil {
			log.Println("[KAFKA] context cancelled — draining in-flight workers")
			c.handler.drain()
			return
		}
	}
}

// Close shuts down the consumer group connection.
func (c *KafkaConsumer) Close() error { return c.group.Close() }
