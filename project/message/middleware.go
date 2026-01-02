package message

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/lithammer/shortuuid/v3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var (
	messagesProcessingDuration = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  "messages",
			Name:       "processing_duration_seconds",
			Help:       "The total time spent processing messages",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"topic", "handler"},
	)

	messagesProcessingFailedCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "messages",
			Name:      "processing_failed_count",
		},
		[]string{"topic", "handler"},
	)

	messagesProcessingTotalCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "messages",
			Name:      "processed_total",
		},
		[]string{"topic", "handler"},
	)
)

func MetricsMiddleware() func(h message.HandlerFunc) message.HandlerFunc {
	return func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) (events []*message.Message, err error) {
			start := time.Now()
			topic := message.SubscribeTopicFromCtx(msg.Context())
			handler := message.HandlerNameFromCtx(msg.Context())
			msgs, err := h(msg)
			messagesProcessingDuration.WithLabelValues(topic, handler).Observe(time.Since(start).Seconds())
			messagesProcessingTotalCounter.With(prometheus.Labels{"topic": topic, "handler": handler}).Inc()
			if err != nil {
				messagesProcessingFailedCounter.With(prometheus.Labels{"topic": topic, "handler": handler}).Inc()
			}
			return msgs, err
		}
	}
}

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

func DistributedTracingMiddleware() func(h message.HandlerFunc) message.HandlerFunc {
	return func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) (events []*message.Message, err error) {
			topic := message.SubscribeTopicFromCtx(msg.Context())
			handler := message.HandlerNameFromCtx(msg.Context())

			// get span information from context
			ctx := otel.GetTextMapPropagator().Extract(msg.Context(), propagation.MapCarrier(msg.Metadata))

			spanName := fmt.Sprintf("topic: %s, handler: %s", topic, handler)
			ctx, span := otel.Tracer("").Start(
				ctx,
				spanName,
				trace.WithAttributes(
					attribute.String("topic", topic),
					attribute.String("handler", handler),
				),
			)
			defer span.End()

			msg.SetContext(ctx)
			msgs, err := h(msg)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
			return msgs, err
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
	router.AddMiddleware(MetricsMiddleware())
	router.AddMiddleware(DistributedTracingMiddleware())
}
