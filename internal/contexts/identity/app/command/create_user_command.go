package command

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/marcelofabianov/redtogreen/internal/contexts/identity/domain/user"
	"github.com/marcelofabianov/redtogreen/internal/platform/adapter/logger"
)

type createUserCommand struct {
	useCase   user.CreateUserUseCase
	publisher user.UserCreatedEventPublisher
	logger    *slog.Logger
	tracer    trace.Tracer
}

func NewCreateUserCommand(
	uc user.CreateUserUseCase,
	pub user.UserCreatedEventPublisher,
	logger *slog.Logger,
) user.CreateUserCommand {
	return &createUserCommand{
		useCase:   uc,
		publisher: pub,
		logger:    logger,
		tracer:    otel.Tracer("identity-command"),
	}
}

func (c *createUserCommand) Execute(
	ctx context.Context,
	input user.CreateUserCommandInput,
) (user.CreateUserOutput, error) {
	ctx, span := c.tracer.Start(ctx, "CreateUserCommand.Execute",
		trace.WithAttributes(
			attribute.String("user.email", input.NewUserInput.Email),
			attribute.String("command.type", "CreateUser"),
		),
	)
	defer span.End()

	loggerWithTrace := c.logger.With(logger.TraceID(input.TraceID.String()))
	loggerWithTrace.Info("starting create user command", "email", input.NewUserInput.Email)

	output, err := c.useCase.Execute(ctx, input.NewUserInput)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to execute use case")
		loggerWithTrace.Error("failed to execute create user use case", "error", err)
		return user.CreateUserOutput{}, err
	}

	payload := user.UserCreatedPayload{
		UserID: output.User.ID,
		Name:   output.User.Name,
		Email:  output.User.Email.String(),
		Phone:  output.User.Phone.String(),
	}

	eventInput := user.USerCreatedEventInput{
		CorrelationID:   input.CorrelationID,
		UserID:          input.UserAuthorID,
		TraceID:         input.TraceID,
		PreviousEventID: input.PreviousEventID,
		CausationID:     input.CausationID,
		Payload:         payload,
	}

	publishErr := c.publisher.PublishUserCreatedEvent(ctx, eventInput)
	if publishErr != nil {
		span.RecordError(publishErr)
		span.SetStatus(codes.Error, "Failed to publish event")
		loggerWithTrace.Error("failed to publish user created event", "error", publishErr)
	}

	span.SetStatus(codes.Ok, "Command finished successfully")
	loggerWithTrace.Info("create user command finished successfully", "user_id", output.User.ID.String())

	return output, nil
}
