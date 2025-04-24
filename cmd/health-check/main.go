package main

import (
	"os"
	"time"

	"github.com/foxinuni/distribuidos-central/internal/models"
	"github.com/foxinuni/distribuidos-central/internal/services"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/zeromq/goczmq.v4"
)

func init() {
	// Set up zerolog logger for debug and pretty print
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
}

func main() {
	// 1. Create the serializer
	serializer := services.NewJsonModelSerializer()

	// 2. Create the dealer
	dealer, err := goczmq.NewDealer("tcp://localhost:5555")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create dealer socket")
	}
	defer dealer.Destroy()

	// 3. Send healthcheck request
	request := &models.Request{
		ID:   1,
		Type: "health-check",
	}
	
	encoded, err := serializer.Encode(request)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to serialize request")
	}

	for {
		log.Info().Msg("Sending heartbeat request to server")
		if err := dealer.SendMessage([][]byte{encoded}); err != nil {
			log.Warn().Err(err).Msg("Failed to send request")
		}

		response, err := dealer.RecvMessage()
		if err != nil {
			log.Warn().Err(err).Msg("Failed to receive response")
			continue
		}

		// Deserialize the response
		var resp models.Response
		if err := serializer.Decode(response[0], &resp); err != nil {
			log.Warn().Err(err).Msg("Failed to deserialize response")
			continue
		}

		// Print the response
		log.Info().Interface("response", resp).Msg("Received response")

		// Sleep for 1 second before sending the next request
		time.Sleep(1 * time.Second)

	}
}
