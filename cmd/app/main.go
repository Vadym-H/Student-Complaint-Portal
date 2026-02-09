package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Vadym-H/Student-Complaint-Portal/internal/config"
	"github.com/Vadym-H/Student-Complaint-Portal/internal/lib/logger"
	"github.com/Vadym-H/Student-Complaint-Portal/internal/services"
)

func main() {
	cfg := config.MustLoad()

	log := logger.SetupLogger(cfg.ENV)
	log.Info("Application started")

	cosmos, err := services.NewCosmosService(cfg.CosmosDB.Endpoint, cfg.CosmosDB.Key, cfg.CosmosDB.Database)
	if err != nil {
		panic(err)
	}
	fmt.Println("Cosmos DB connected!")

	//user := &models.User{
	//	Email:        "user@example.com",
	//	PasswordHash: "hashedPassword",
	//	Role:         "student",
	//	CreatedAt:    time.Now(),
	//}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	//if err := cosmos.CreateUser(ctx, user); err != nil {
	//	panic(err)
	//}

	user, err := cosmos.GetUserByEmail(ctx, "user@example.com")
	if err != nil {
		log.Error("Error occurred while retrieving user")
		panic(err)
	}
	log.Info("retrieved user", "id", user.ID, "email", user.Email, "role", user.Role)
}
