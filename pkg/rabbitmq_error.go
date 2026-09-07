package pkg

import (
    "errors"
    "fmt"
    "net"

    amqp "github.com/rabbitmq/amqp091-go"
)

// Domain errors
var (
    ErrRabbitConnectionFailed = errors.New("rabbitmq connection failed")
    ErrRabbitTimeout          = errors.New("rabbitmq operation timed out")
    ErrRabbitChannelClosed    = errors.New("rabbitmq channel is closed")
    ErrRabbitQueueNotFound    = errors.New("rabbitmq queue not found")
    ErrRabbitPublishFailed    = errors.New("failed to publish message")
    ErrRabbitConsumeFailed    = errors.New("failed to consume message")
    ErrRabbitInternal         = errors.New("internal rabbitmq error")
)

// HandleRabbitError map raw RabbitMQ error to domain error to prevent it from leaking out.
func HandleRabbitError(err error) error {
    if err == nil {
        return nil
    }

    if errors.Is(err, amqp.ErrClosed) {
        return ErrRabbitChannelClosed
    }

    var netErr *net.OpError
    if errors.As(err, &netErr) {
        return fmt.Errorf("%w: %s", ErrRabbitConnectionFailed, netErr.Op)
    }

    if isRabbitTimeoutError(err) {
        return ErrRabbitTimeout
    }

    msg := err.Error()
    if contains(msg, "NOT_FOUND") {
        return ErrRabbitQueueNotFound
    }

    return fmt.Errorf("%w: %s", ErrRabbitInternal, err.Error())
}

func isRabbitTimeoutError(err error) bool {
    if err == nil {
        return false
    }
    msg := err.Error()
    return contains(msg, "timeout") || contains(msg, "deadline exceeded") || contains(msg, "context canceled")
}