package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"runtime"

	"github.com/foxinuni/distribuidos-central/internal/handler"
	"github.com/foxinuni/distribuidos-central/internal/handler/controllers"
	"github.com/foxinuni/distribuidos-central/internal/services"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var config Config

func init() {
	// Load config from flags
	flag.IntVar(&config.Port, "port", 5555, "Port to listen on")
	flag.IntVar(&config.Workers, "workers", runtime.NumCPU(), "Number of worker goroutines")
	flag.StringVar(&config.DatabaseURL, "database", "postgresql://postgres:postgres@127.0.0.1:5432/distribuidos_central?sslmode=disable", "Database URL")
	flag.BoolVar(&config.Debug, "debug", false, "Enable debug logging")
	flag.Parse()

	// Set up zerolog logger for debug and pretty print
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	if config.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

func main() {
	// 1. Connect to database
	pool, err := pgxpool.New(context.Background(), config.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}

	// 1.1 Ping database
	if err := pool.Ping(context.Background()); err != nil {
		log.Fatal().Err(err).Msg("Failed to ping database")
	}

	// 2. Construct services for server
	serializerService := services.NewJsonModelSerializer()
	allocationsService := services.NewSqlcAllocationService(pool)

	// 3. Construct controllers for server
	healthCheckController := controllers.NewHealthCheckController()
	allocationsController := controllers.NewAllocationsController(allocationsService)

	// 4. Boostrap the server
	server := handler.NewServer(
		healthCheckController,
		allocationsController,
		serializerService,

		// Optional server options
		handler.WithPort(config.Port),
		handler.WithWorkerCount(config.Workers),
	)

	// 5. Start the server
	if err := server.Start(); err != nil {
		log.Error().Err(err).Msg("Failed to start server")
		os.Exit(1)
	}
	defer server.Stop()

	// 6. Wait for shutdown signal (CTRL+C)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
}
