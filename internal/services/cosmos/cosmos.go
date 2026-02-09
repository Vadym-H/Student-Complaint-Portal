package cosmos

import (
	"errors"
	"log/slog"

	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"
)

// Error constants
var (
	ErrInvalidRole           = errors.New("invalid user role")
	ErrEmailAlreadyExists    = errors.New("user with this email already exists")
	ErrUsernameAlreadyExists = errors.New("user with this username already exists")
)

type Service struct {
	client              *azcosmos.Client
	database            string
	usersContainer      string
	complaintsContainer string
	log                 *slog.Logger
}

// NewCosmosService creates a new CosmosService with the given endpoint, key, and database
func NewCosmosService(endpoint, key, database string, log *slog.Logger) (*Service, error) {
	const module = "cosmos"
	log = log.With(
		slog.String("module", module),
	)
	cred, err := azcosmos.NewKeyCredential(key)
	if err != nil {
		return nil, err
	}

	client, err := azcosmos.NewClientWithKey(endpoint, cred, nil)
	if err != nil {
		return nil, err
	}

	log.Info("cosmos DB service initialized", slog.String("database", database))

	return &Service{
		client:              client,
		database:            database,
		usersContainer:      "users",
		complaintsContainer: "complaints",
		log:                 log,
	}, nil
}

// PublicServiceTest For testing only
type PublicServiceTest struct {
	Client              *azcosmos.Client
	Database            string
	UsersContainer      string
	ComplaintsContainer string
}

func NewCosmosServiceTest(endpoint, key, database string) (*PublicServiceTest, error) {
	cred, err := azcosmos.NewKeyCredential(key)
	if err != nil {
		return nil, err
	}

	client, err := azcosmos.NewClientWithKey(endpoint, cred, nil)
	if err != nil {
		return nil, err
	}

	return &PublicServiceTest{
		Client:              client,
		Database:            database,
		UsersContainer:      "users",
		ComplaintsContainer: "complaints",
	}, nil
}
