package main

import (
	"github.com/ashurov-imomali/pr-service/internal/api"
	"github.com/ashurov-imomali/pr-service/internal/repository"
	"github.com/ashurov-imomali/pr-service/internal/server"
	"github.com/ashurov-imomali/pr-service/internal/usecase"
	"github.com/ashurov-imomali/pr-service/migration"
	"github.com/ashurov-imomali/pr-service/pkg/db"
	"github.com/ashurov-imomali/pr-service/pkg/logger"
	"golang.org/x/net/context"
	"os"
	"os/signal"
	"time"
)

func main() {
	log := logger.New()
	dbDSN := os.Getenv("DB_DSN")
	if dbDSN == "" {
		dbDSN = "host=localhost user=postgres password=postgres dbname=postgres sslmode=disable"
	}

	// Get APP port
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	if err := migration.Run(dbDSN); err != nil {
		log.Fatalf("Error in migration. Err %v", err)
	}

	pgConnection, err := db.New(dbDSN)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	r := repository.New(pgConnection)

	prs := usecase.NewPRService(r, log)
	us := usecase.NewUserService(r, log)
	ts := usecase.NewTeamService(r, log)

	h := api.New(prs, us, ts)

	srv := server.NewServer(":"+port, h)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		log.Infof("Server starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-stop
	log.Infof("%s", "Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Errorf("Graceful shutdown failed: %v", err)
	} else {
		log.Infof("%s", "Server stopped gracefully")
	}
}
