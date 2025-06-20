package bus

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/marcelofabianov/redtogreen/internal/platform/adapter/logger"
	logattr "github.com/marcelofabianov/redtogreen/internal/platform/adapter/logger"
	"github.com/marcelofabianov/redtogreen/internal/platform/config"
	"github.com/marcelofabianov/redtogreen/internal/platform/event"
	"github.com/marcelofabianov/redtogreen/internal/platform/msg"
	platformBus "github.com/marcelofabianov/redtogreen/internal/platform/port/bus"
)

const (
	logFailedToConnect        = "Failed to connect to NATS"
	logFailedToInitJetStream  = "Failed to initialize JetStream"
	logFailedToCreateStream   = "Failed to create stream"
	logStreamCreated          = "Stream created successfully"
	logNatsBusInitialized     = "NATS EventBus successfully initialized"
	logFailedToMarshalEvent   = "Failed to marshal event"
	logFailedToPublishEvent   = "Failed to publish event"
	logEventPublished         = "Event successfully published"
	logNoStreamForSubject     = "Configuration error: could not find a stream for the subject"
	logFoundStreamForSub      = "Found matching stream for subscription"
	logFailedToCreateConsumer = "Failed to create or update consumer"
	logFailedToUnmarshalMsg   = "Failed to unmarshal NATS message into event struct, terminating message"
	logFailedToTermMsg        = "Failed to terminate message"
	logEventHandlerFailed     = "Event handler failed, will allow retry"
	logFailedToAckMsg         = "Failed to acknowledge message"
	logEventProcessed         = "Event successfully processed"
	logFailedToConsume        = "Failed to start consuming messages"
	logSubscribedSuccessfully = "Subscribed successfully to event type"

	consumerNameSuffix = "-processor"
)

type NatsEventBus struct {
	nc     *nats.Conn
	js     jetstream.JetStream
	logger *slog.Logger
	tracer trace.Tracer
}

func NewNatsEventBus(config *config.NATSConfig, sl *slog.Logger) (*NatsEventBus, error) {
	nc, err := nats.Connect(config.URLs)
	if err != nil {
		errMsg := msg.NewInternalError(err, map[string]any{"urls": config.URLs})
		sl.Error(logFailedToConnect,
			logattr.ErrorCode(errMsg.Code),
			slog.String("urls", config.URLs),
			logattr.Err(err),
		)
		return nil, errMsg
	}

	js, err := jetstream.New(nc)
	if err != nil {
		errMsg := msg.NewInternalError(err, nil)
		sl.Error(logFailedToInitJetStream,
			logattr.ErrorCode(errMsg.Code),
			logattr.Err(err),
		)
		return nil, errMsg
	}

	for _, stream := range GetStreamConfigs() {
		_, err := js.Stream(context.Background(), stream.Name)
		if err != nil {
			jsStreamConfig := jetstream.StreamConfig{
				Name:      stream.Config.Name,
				Subjects:  stream.Config.Subjects,
				Storage:   jetstream.StorageType(stream.Config.Storage),
				Replicas:  stream.Config.Replicas,
				MaxMsgs:   stream.Config.MaxMsgs,
				MaxAge:    stream.Config.MaxAge,
				Retention: jetstream.RetentionPolicy(stream.Config.Retention),
			}

			if _, addErr := js.CreateStream(context.Background(), jsStreamConfig); addErr != nil {
				sl.Warn(logFailedToCreateStream,
					slog.String("stream", stream.Name),
					logattr.Err(addErr),
				)
			} else {
				sl.Info(logStreamCreated,
					slog.String("stream", stream.Name),
				)
			}
		}
	}

	sl.Info(logNatsBusInitialized,
		slog.String("urls", config.URLs),
	)

	return &NatsEventBus{
		nc:     nc,
		js:     js,
		logger: sl,
		tracer: otel.Tracer("nats-bus"),
	}, nil
}

func (b *NatsEventBus) Publish(ctx context.Context, evt *event.Event) error {
	ctx, span := b.tracer.Start(ctx, fmt.Sprintf("NATS Publish %s", evt.Header.EventType),
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			attribute.String("messaging.system", "nats"),
			attribute.String("messaging.destination", string(evt.Header.EventType)),
			attribute.String("messaging.operation", "publish"),
			attribute.String("event.id", evt.Header.EventID.String()),
			attribute.String("event.type", string(evt.Header.EventType)),
		),
	)
	defer span.End()

	carrier := propagation.HeaderCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	data, err := json.Marshal(evt)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to marshal event")
		errMsg := msg.NewValidationError(err, map[string]any{"event_type": evt.Header.EventType}, "Invalid event data")
		b.logger.Error(logFailedToMarshalEvent,
			logger.ErrorCode(errMsg.Code),
			logger.EventType(string(evt.Header.EventType)),
			logger.Err(err),
		)
		return errMsg
	}

	msgToPublish := nats.Msg{
		Subject: string(evt.Header.EventType),
		Data:    data,
		Header:  make(nats.Header),
	}
	for k, v := range carrier {
		for _, val := range v {
			msgToPublish.Header.Add(k, val)
		}
	}

	_, err = b.js.PublishMsg(ctx, &msgToPublish)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to publish event")
		errMsg := msg.NewInternalError(err, map[string]any{"event_type": evt.Header.EventType})
		b.logger.Error(logFailedToPublishEvent,
			logger.ErrorCode(errMsg.Code),
			logger.EventType(string(evt.Header.EventType)),
			logger.Err(err),
		)
		return errMsg
	}

	span.SetStatus(codes.Ok, "Event published successfully")
	b.logger.Debug(logEventPublished,
		logger.EventType(string(evt.Header.EventType)),
	)

	return nil
}

func (b *NatsEventBus) Subscribe(eventType event.EventType, handler platformBus.EventHandler) error {
	safeConsumerPrefix := strings.ReplaceAll(string(eventType), ".", "-")
	consumerName := safeConsumerPrefix + consumerNameSuffix
	subject := string(eventType)

	streamName, err := findStreamNameForSubject(subject)
	if err != nil {
		b.logger.Error(logNoStreamForSubject,
			slog.String("subject", subject),
			logger.Err(err),
		)
		return err
	}
	b.logger.Info(logFoundStreamForSub,
		slog.String("subject", subject),
		slog.String("stream", streamName),
	)

	consumer, err := b.js.CreateOrUpdateConsumer(context.Background(), streamName, jetstream.ConsumerConfig{
		Durable:       consumerName,
		FilterSubject: subject,
		AckPolicy:     jetstream.AckExplicitPolicy,
		MaxDeliver:    5,
		AckWait:       30 * time.Second,
		BackOff:       []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second, 16 * time.Second},
	})
	if err != nil {
		errMsg := msg.NewInternalError(err, map[string]any{"consumer": consumerName})
		b.logger.Error(logFailedToCreateConsumer,
			logger.ErrorCode(errMsg.Code),
			slog.String("consumer", consumerName),
			slog.String("stream", streamName),
			logger.Err(err),
		)
		return errMsg
	}

	_, err = consumer.Consume(func(natsMsg jetstream.Msg) {

		carrier := propagation.HeaderCarrier(natsMsg.Headers())
		ctx := otel.GetTextMapPropagator().Extract(context.Background(), carrier)

		ctx, span := b.tracer.Start(ctx, fmt.Sprintf("NATS Consume %s", eventType),
			trace.WithSpanKind(trace.SpanKindConsumer),
			trace.WithAttributes(
				attribute.String("messaging.system", "nats"),
				attribute.String("messaging.destination", string(eventType)),
				attribute.String("messaging.operation", "process"),
			),
		)
		defer span.End()

		var evt event.Event
		if err := json.Unmarshal(natsMsg.Data(), &evt); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "Failed to unmarshal NATS message")
			b.logger.Error(logFailedToUnmarshalMsg,
				logger.Err(err),
				slog.String("data", string(natsMsg.Data())),
			)
			if termErr := natsMsg.Term(); termErr != nil {
				b.logger.Error(logFailedToTermMsg, logger.Err(termErr))
			}
			return
		}

		if err := handler(ctx, &evt); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "Event handler failed")
			b.logger.Error(logEventHandlerFailed,
				logger.Err(err),
				logger.EventType(string(eventType)),
			)
			return
		}

		if ackErr := natsMsg.Ack(); ackErr != nil {
			b.logger.Error(logFailedToAckMsg,
				logger.Err(ackErr),
				logger.EventType(string(eventType)),
			)
		}

		span.SetStatus(codes.Ok, "Event processed successfully")
		b.logger.Debug(logEventProcessed,
			logger.EventType(string(eventType)),
		)
	})
	return nil
}
