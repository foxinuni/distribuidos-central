package main

import (
	"context"
	"flag"
	"os"

	"github.com/foxinuni/distribuidos-central/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var config Config

func init() {
	flag.IntVar(&config.Classrooms, "classrooms", 350, "Number of classrooms")
	flag.IntVar(&config.Laboratories, "laboratories", 100, "Number of laboratories")
	flag.StringVar(&config.DatabaseURL, "database", "postgresql://postgres:postgres@127.0.0.1:5432/distribuidos_central?sslmode=disable", "Database URL")
	flag.Parse()

	// Set up zerolog logger for debug and pretty print
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
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

	// 2. Generate querier
	querier := repository.New(pool)

	// 3. Generate migrations
	log.Info().Msg("Generating migrations ...")
	if querier.GenerateRooms(context.Background(), repository.GenerateRoomsParams{
		NormalRooms:  int32(config.Classrooms),
		Laboratories: int32(config.Laboratories),
	}); err != nil {
		log.Fatal().Err(err).Msg("Failed to generate rooms")
	}
}
