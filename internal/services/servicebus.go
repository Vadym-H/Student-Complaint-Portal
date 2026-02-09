package services

import (
	"context"
	"log/slog"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
)

type ServiceBusService struct {
	client *azservicebus.Client
	log    *slog.Logger
}

// NewServiceBusService creates a new ServiceBusService with the given connection string
func NewServiceBusService(connectionString string, log *slog.Logger) (*ServiceBusService, error) {
	client, err := azservicebus.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		return nil, err
	}

	log.Info("service bus service initialized")

	return &ServiceBusService{
		client: client,
		log:    log,
	}, nil
}

// SendMessage sends a message to the specified queue
func (s *ServiceBusService) SendMessage(ctx context.Context, queueName, messageBody string) error {
	sender, err := s.client.NewSender(queueName, nil)
	if err != nil {
		s.log.Error("failed to create service bus sender", slog.String("queue", queueName), slog.String("error", err.Error()))
		return err
	}
	defer func(sender *azservicebus.Sender, ctx context.Context) {
		err := sender.Close(ctx)
		if err != nil {
			s.log.Error("failed to close service bus sender", slog.String("error", err.Error()))
		}
	}(sender, ctx)

	message := &azservicebus.Message{
		Body: []byte(messageBody),
	}

	err = sender.SendMessage(ctx, message, nil)
	if err != nil {
		s.log.Error("failed to send message to service bus", slog.String("queue", queueName), slog.String("error", err.Error()))
		return err
	}

	s.log.Info("message sent to service bus", slog.String("queue", queueName))
	return nil
}
