package message

import (
	"log/slog"
	"time"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/lithammer/shortuuid/v3"
)

func LoggingMiddleware() func(h message.HandlerFunc) message.HandlerFunc {
	return func(next message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			logger := log.FromContext(msg.Context())
			logger.With(
				"payload", string(msg.Payload),
				"metadata", msg.Metadata,
				"handler", message.HandlerNameFromCtx(msg.Context()),
			)
			logger.Info("Handling a message, message_id=", msg.UUID)

			msgs, err := next(msg)
			if err != nil {
				logger.With(
					"error", err,
					"message_id", msg.UUID,
				).Error("Error while handling a message")

			}

			return msgs, err
		}
	}
}

func CorrelationIdMiddleware() func(h message.HandlerFunc) message.HandlerFunc {
	return func(next message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			correlationID := msg.Metadata.Get("correlation_id")
			if correlationID == "" {
				correlationID = shortuuid.New()
			}

			ctx := log.ContextWithCorrelationID(msg.Context(), correlationID)
			ctx = log.ToContext(ctx, slog.With("correlation_id", correlationID))

			msg.SetContext(ctx)
			return next(msg)
		}
	}
}

func RetryMiddleware(watermillLogger watermill.LoggerAdapter) func(h message.HandlerFunc) message.HandlerFunc {
	retry := middleware.Retry{
		MaxRetries:      10,
		InitialInterval: time.Millisecond * 100,
		MaxInterval:     time.Second,
		Multiplier:      2,
		Logger:          watermillLogger,
	}

	return retry.Middleware
}

func AddMiddleWare(router *message.Router, watermillLogger watermill.LoggerAdapter) {
	router.AddMiddleware(CorrelationIdMiddleware())
	router.AddMiddleware(LoggingMiddleware())
	router.AddMiddleware(RetryMiddleware(watermillLogger))
}
