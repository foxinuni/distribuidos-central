package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand/v2"
	"os"
	"sync"

	"github.com/foxinuni/distribuidos-central/internal/models"
	"github.com/foxinuni/distribuidos-central/internal/services"
	"github.com/go-zeromq/zmq4"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var config Config

func init() {
	flag.IntVar(&config.Faculties, "faculties", 10, "Number of facultires")
	flag.StringVar(&config.Address, "address", "tcp://127.0.0.1:5555", "The server address")
	flag.Parse()

	// Set up zerolog logger for debug and pretty print
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
}

func main() {
	log.Info().Msgf("Starting faculty client with %d faculties (for: %q)", config.Faculties, config.Address)

	// 1. Create the serializer
	serializer := services.NewJsonModelSerializer()

	// 2. Create waitgroup and start threads
	var waitgroup sync.WaitGroup
	for i := 0; i < config.Faculties; i++ {
		waitgroup.Add(1)

		// 2.1 Create faculty thread
		go func() {
			defer waitgroup.Done()
			facultyWorker(i, serializer)
		}()
	}

	// 3. Wait for threads to finish
	waitgroup.Wait()
	log.Info().Msg("All threads finished")
}

func facultyWorker(id int, serializer *services.JsonModelSerializer) {
	logger := log.With().Str("faculty", Faculties[id]).Logger()

	// 1. Create the dealer
	dealer := zmq4.NewReq(context.Background())
	defer dealer.Close()

	if err := dealer.Dial(config.Address); err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to server")
	}

	log.Info().Msgf("Starting faculty worker for %s", Faculties[id])

	// 2. Open log file
	file, err := os.OpenFile(fmt.Sprintf("logs/%s.json", Faculties[id]), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to open file")
	}
	defer file.Close()

	{
		// 3. Create request
		content := &models.AllocateRequest{
			Semester: "2025-1",
			Faculty:  Faculties[id],
			Programs: []models.ProgramInfo{},
		}

		for i := 0; i < 5; i++ {
			info := models.ProgramInfo{
				Name:         Programs[id][i],
				Classrooms:   7 + rand.IntN(4),
				Laboratories: 2 + rand.IntN(3),
			}

			content.Programs = append(content.Programs, info)
		}

		request := &models.Request{
			ID:      id,
			Type:    "allocate",
			Content: content,
		}

		// 4. Encode the request
		encoded, err := serializer.Encode(request)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to serialize request")
		}

		// 5. Send the request
		if err := dealer.Send(zmq4.NewMsgFrom([][]byte{encoded}...)); err != nil {
			logger.Fatal().Err(err).Msg("Failed to send request")
		}

		// 6. Receive the response
		response, err := dealer.Recv()
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to receive response")
		}

		// 7. Decode the response
		var resp models.Response
		if err := serializer.Decode(response.Frames[0], &resp); err != nil {
			logger.Fatal().Err(err).Msg("Failed to deserialize response")
		}

		// 8. Print the response
		logger.Info().Interface("response", resp).Msg("Received response")

		// 9. Write to a file
		if _, err := file.Write(response.Frames[0]); err != nil {
			logger.Warn().Err(err).Msg("Failed to write to file")
		}
	}

	{
		// 10. Confirm request
		content := &models.ConfirmRequest{
			Semester: "2025-1",
			Faculty:  Faculties[id],
			Accept:   true,
		}

		request := &models.Request{
			ID:      id,
			Type:    "confirm",
			Content: content,
		}

		// 11. Encode the request
		encoded, err := serializer.Encode(request)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to serialize request")
		}

		// 12. Send the request
		if err := dealer.Send(zmq4.NewMsgFrom([][]byte{encoded}...)); err != nil {
			logger.Fatal().Err(err).Msg("Failed to send request")
		}

		// 13. Receive the response
		response, err := dealer.Recv()
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to receive response")
		}

		// 14. Decode the response
		var resp models.Response
		if err := serializer.Decode(response.Frames[0], &resp); err != nil {
			logger.Fatal().Err(err).Msg("Failed to deserialize response")
		}

		// 15. Print the response
		logger.Info().Interface("response", resp).Msg("Received response")

		// 16. Write to a file
		if _, err := file.Write(response.Frames[0]); err != nil {
			logger.Warn().Err(err).Msg("Failed to write to file")
		}
	}
}
