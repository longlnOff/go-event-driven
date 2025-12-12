package message

import (
	"log/slog"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"github.com/ThreeDotsLabs/watermill/message"
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

			return next(msg)
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

func AddMiddleWare(router *message.Router) {
	router.AddMiddleware(CorrelationIdMiddleware())
	router.AddMiddleware(LoggingMiddleware())
}
