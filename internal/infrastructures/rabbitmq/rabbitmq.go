package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"

	"email-job/internal/domains"
	"email-job/internal/emails"
)

const (
	mainExchange   = "email.main.exchange"
	retryExchange  = "email.retry.exchange"
	queueName      = "email.queue"
	retryQueueName = "email.retry.queue"
)

type RabbitMQ struct {
	conn   *amqp.Connection
	ch     *amqp.Channel
	logger *slog.Logger
}

func NewRabbitMQClient(dsn string, logger *slog.Logger) (*RabbitMQ, error) {
	conn, err := amqp.Dial(dsn)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq: dial %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("rabbitmq: channel %w", err)
	}

	r := &RabbitMQ{conn: conn, ch: ch, logger: logger}
	if err := r.setup(); err != nil {
		r.Close()
		return nil, err
	}

	return r, nil
}

func (r *RabbitMQ) Channel() *amqp.Channel { 
	return r.ch 
}

func (r *RabbitMQ) Close() {
	if r.ch != nil {
		r.ch.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}
}

func (r *RabbitMQ) setup() error {
	if err := r.ch.ExchangeDeclare(mainExchange, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("rabbitmq: declare main exchange %w", err)
	}

	if err := r.ch.ExchangeDeclare(retryExchange, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("rabbitmq: declare retry exchange %w", err)
	}

	if _, err := r.ch.QueueDeclare(queueName, true, false, false, false, nil); err != nil {
		return fmt.Errorf("rabbitmq: declare main queue %w", err)
	}
	if err := r.ch.QueueBind(queueName, queueName, mainExchange, false, nil); err != nil {
		return fmt.Errorf("rabbitmq: bind main queue %w", err)
	}

	retryArgs := amqp.Table{
		"x-dead-letter-exchange":    mainExchange,
		"x-dead-letter-routing-key": queueName,
	}
	if _, err := r.ch.QueueDeclare(retryQueueName, true, false, false, false, retryArgs); err != nil {
		return fmt.Errorf("rabbitmq: declare retry queue %w", err)
	}
	if err := r.ch.QueueBind(retryQueueName, retryQueueName, retryExchange, false, nil); err != nil {
		return fmt.Errorf("rabbitmq: bind retry queue %w", err)
	}

	return nil
}

func (r *RabbitMQ) StartWorker(ctx context.Context, svc emails.EmailService) error {
	if err := r.ch.Qos(1, 0, false); err != nil {
		return fmt.Errorf("rabbitmq: qos %w", err)
	}

	msgs, err := r.ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("rabbitmq: consume %w", err)
	}

	r.logger.Info("rabbitmq: email worker started", slog.String("queue", queueName))

	for {
		select {
		case <-ctx.Done():
			r.logger.Info("rabbitmq: email worker stopping")
			return nil

		case msg, ok := <-msgs:
			if !ok {
				r.logger.Warn("rabbitmq: channel closed")
				return fmt.Errorf("rabbitmq: channel closed unexpectedly")
			}

			var payload domains.EmailQueueEntity
			if err := json.Unmarshal(msg.Body, &payload); err != nil {
				r.logger.Error("rabbitmq: bad message, discarding", slog.Any("error", err))
				msg.Nack(false, false)
				continue
			}

			if err := svc.DeliverQueued(ctx, payload); err != nil {
				r.logger.Error("rabbitmq: failed to deliver email, nacking",
					slog.String("type", string(payload.Type)),
					slog.Any("error", err),
				)
				msg.Nack(false, false)
				continue
			}

			r.logger.Info("rabbitmq: email delivered",
				slog.String("type", string(payload.Type)),
				slog.String("to", payload.To),
			)

			msg.Ack(false)
		}
	}
}