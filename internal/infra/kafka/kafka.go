package kafka

import (
	"context"
	"errors"
	"fmt"
	"time"

	kafkago "github.com/segmentio/kafka-go"

	"github.com/alonespig/go-start/config"
)

// Message 是生产/消费统一使用的消息结构，屏蔽底层库类型
type Message struct {
	Topic     string
	Key       []byte
	Value     []byte
	Time      time.Time
	Partition int
	Offset    int64
}

// Producer 消息生产者抽象，便于替换实现或单测 mock
type Producer interface {
	Publish(ctx context.Context, topic string, key, value []byte) error
	Close() error
}

// Consumer 消息消费者抽象；handler 返回 error 时不提交 offset，便于重试
type Consumer interface {
	Consume(ctx context.Context, handler func(ctx context.Context, msg Message) error) error
	Close() error
}

type producer struct {
	writer *kafkago.Writer
}

// NewProducer 创建生产者。kafka-go 的连接是惰性的，启动时不会阻塞。
func NewProducer(cfg config.KafkaConfig) (Producer, error) {
	if len(cfg.Brokers) == 0 {
		return nil, fmt.Errorf("kafka.brokers 不能为空")
	}
	return &producer{writer: &kafkago.Writer{
		Addr:         kafkago.TCP(cfg.Brokers...),
		Balancer:     &kafkago.LeastBytes{},
		RequiredAcks: kafkago.RequireOne,
		BatchTimeout: 10 * time.Millisecond,
	}}, nil
}

func (p *producer) Publish(ctx context.Context, topic string, key, value []byte) error {
	if topic == "" {
		return fmt.Errorf("topic 不能为空")
	}
	return p.writer.WriteMessages(ctx, kafkago.Message{
		Topic: topic,
		Key:   key,
		Value: value,
	})
}

func (p *producer) Close() error {
	return p.writer.Close()
}

type consumer struct {
	reader *kafkago.Reader
}

// NewConsumer 为指定 topic 创建消费者；消费逻辑见 Consume。
func NewConsumer(cfg config.KafkaConfig, topic string) (Consumer, error) {
	if len(cfg.Brokers) == 0 {
		return nil, fmt.Errorf("kafka.brokers 不能为空")
	}
	if topic == "" {
		return nil, fmt.Errorf("topic 不能为空")
	}
	return &consumer{reader: kafkago.NewReader(kafkago.ReaderConfig{
		Brokers: cfg.Brokers,
		GroupID: cfg.GroupID,
		Topic:   topic,
	})}, nil
}

// Consume 阻塞消费直到 ctx 取消；只有 handler 成功才提交 offset。
func (c *consumer) Consume(ctx context.Context, handler func(ctx context.Context, msg Message) error) error {
	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil
			}
			return err
		}

		msg := Message{
			Topic:     m.Topic,
			Key:       m.Key,
			Value:     m.Value,
			Time:      m.Time,
			Partition: m.Partition,
			Offset:    m.Offset,
		}
		if err := handler(ctx, msg); err != nil {
			return err
		}
		if err := c.reader.CommitMessages(ctx, m); err != nil {
			return err
		}
	}
}

func (c *consumer) Close() error {
	return c.reader.Close()
}
