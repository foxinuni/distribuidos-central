package main

import (
	"flag"
	"os"

	"github.com/foxinuni/distribuidos-central/internal/models"
	"github.com/foxinuni/distribuidos-central/internal/services"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/zeromq/goczmq.v4"
)

var config Config

func init() {
	flag.IntVar(&config.Port, "port", 5555, "Port to listen on")
	flag.IntVar(&config.Programs, "programs", 10, "Number of programs")
	flag.IntVar(&config.Classrooms, "classrooms", 10, "Number of classrooms")
	flag.IntVar(&config.Laboratories, "laboratories", 10, "Number of laboratories")
	flag.Parse()

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

	// 3. Create request
	content := &models.AllocateRequest{
		Semester: "2025-1",
		Faculty:  "Ingeniería",
		Programs: []models.ProgramInfo{},
	}

	for i := 0; i < config.Programs; i++ {
		info := models.ProgramInfo{
			Name:         ProgramNames[i],
			Classrooms:   config.Classrooms,
			Laboratories: config.Laboratories,
		}

		content.Programs = append(content.Programs, info)
	}

	request := &models.Request{
		ID:      1,
		Type:    "allocate",
		Content: content,
	}

	// 4. Encode the request
	encoded, err := serializer.Encode(request)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to serialize request")
	}

	// 5. Send the request
	if err := dealer.SendMessage([][]byte{encoded}); err != nil {
		log.Fatal().Err(err).Msg("Failed to send request")
	}

	// 6. Receive the response
	response, err := dealer.RecvMessage()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to receive response")
	}

	// 7. Decode the response
	var resp models.Response
	if err := serializer.Decode(response[0], &resp); err != nil {
		log.Fatal().Err(err).Msg("Failed to deserialize response")
	}

	// 8. Print the response
	log.Info().Interface("response", resp).Msg("Received response")
}
