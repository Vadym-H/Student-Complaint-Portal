package services

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
)

type ServiceBusService struct {
	client *azservicebus.Client
}

// NewServiceBusService creates a new ServiceBusService with the given connection string
func NewServiceBusService(connectionString string) (*ServiceBusService, error) {
	client, err := azservicebus.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		return nil, err
	}

	return &ServiceBusService{
		client: client,
	}, nil
}

// SendMessage sends a message to the specified queue
func (s *ServiceBusService) SendMessage(ctx context.Context, queueName, messageBody string) error {
	sender, err := s.client.NewSender(queueName, nil)
	if err != nil {
		return err
	}
	defer func(sender *azservicebus.Sender, ctx context.Context) {
		err := sender.Close(ctx)
		if err != nil {

		}
	}(sender, ctx)

	message := &azservicebus.Message{
		Body: []byte(messageBody),
	}

	err = sender.SendMessage(ctx, message, nil)
	if err != nil {
		return err
	}

	return nil
}
