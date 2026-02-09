package cosmos

import (
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
)

type Service struct {
	client              *azcosmos.Client
	database            string
	usersContainer      string
	complaintsContainer string
}

// NewCosmosService creates a new CosmosService with the given endpoint, key, and database
func NewCosmosService(endpoint, key, database string) (*Service, error) {
	cred, err := azcosmos.NewKeyCredential(key)
	if err != nil {
		return nil, err
	}

	client, err := azcosmos.NewClientWithKey(endpoint, cred, nil)
	if err != nil {
		return nil, err
	}

	return &Service{
		client:              client,
		database:            database,
		usersContainer:      "users",
		complaintsContainer: "complaints",
	}, nil
}
