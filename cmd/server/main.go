package main

import (
	"os"
	"os/signal"

	"github.com/foxinuni/distribuidos-central/internal/services"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	// Set up zerolog logger for debug and pretty print
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
}

func main() {
	// 1. Construct services for server
	serializer := services.NewJsonModelSerializer()

	// 2. Boostrap the server
	server := NewController(serializer, WithWorkerCount(20))

	// 3. Start the server
	if err := server.Start(); err != nil {
		log.Error().Err(err).Msg("Failed to start server")
		os.Exit(1)
	}
	defer server.Stop()

	// 4. Wait for shutdown signal (CTRL+C)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
}
