package main

import (
	"github.com/Vadym-H/Student-Complaint-Portal/internal/config"
	"github.com/Vadym-H/Student-Complaint-Portal/internal/lib/logger"
)

func main() {
	cfg := config.MustLoad()

	log := logger.SetupLogger(cfg.ENV)
	log.Info("Application started")

}
