package consumer

import (
	"github.com/IBM/sarama"
	"webhook-listener/internal/config"
)

// KafkaProducer publishes raw event payloads back to the webhook-events topic.
type KafkaProducer struct {
	producer sarama.SyncProducer
	topic    string
}

func NewKafkaProducer(cfg *config.Config) (*KafkaProducer, error) {
	pcfg := sarama.NewConfig()
	pcfg.Version = sarama.MaxVersion
	pcfg.Producer.Return.Successes = true
	pcfg.Producer.Return.Errors = true

	p, err := sarama.NewSyncProducer(cfg.KafkaBrokers, pcfg)
	if err != nil {
		return nil, err
	}
	return &KafkaProducer{producer: p, topic: cfg.KafkaTopic}, nil
}

func (p *KafkaProducer) Publish(payload string) error {
	_, _, err := p.producer.SendMessage(&sarama.ProducerMessage{
		Topic: p.topic,
		Value: sarama.StringEncoder(payload),
	})
	return err
}

func (p *KafkaProducer) Close() error { return p.producer.Close() }
