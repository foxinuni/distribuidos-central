package handler

import (
	"fmt"
	"sync"

	"github.com/foxinuni/distribuidos-central/internal/handler/controllers"
	"github.com/foxinuni/distribuidos-central/internal/services"
	"github.com/rs/zerolog/log"
	"gopkg.in/zeromq/goczmq.v4"
)

type Server struct {
	// internal
	port      int
	workers   int
	waitgroup sync.WaitGroup

	stopch   chan struct{}
	requests chan [][]byte
	routes   map[string]RouteHandler

	// controllers
	healthCheckController *controllers.HealthCheckController
	allocationsController *controllers.AllocationsController

	// external
	socket     *goczmq.Channeler
	serializer services.ModelSerializer
}

func NewServer(
	healthCheckController *controllers.HealthCheckController,
	allocationsController *controllers.AllocationsController,
	serializer services.ModelSerializer,
	options ...ServerOptions,
) *Server {
	server := &Server{
		port:                  5555,
		workers:               10,
		requests:              make(chan [][]byte),
		stopch:                make(chan struct{}),
		routes:                make(map[string]RouteHandler),
		serializer:            serializer,
		healthCheckController: healthCheckController,
		allocationsController: allocationsController,
	}

	for _, applyOption := range options {
		applyOption(server)
	}

	server.registerRoutes()

	return server
}

func (s *Server) Start() error {

	log.Info().Msgf("Starting server on port %d with %d workers", s.port, s.workers)

	// Start the socket
	s.socket = goczmq.NewRouterChanneler(fmt.Sprintf("tcp://*:%d", s.port))
	if s.socket == nil {
		return fmt.Errorf("failed to create socket")
	}

	// Start the workers
	for i := 0; i < s.workers; i++ {
		s.waitgroup.Add(1)

		go func() {
			defer s.waitgroup.Done()
			s.worker(i + 1)
		}()
	}

	// Start the main loop
	go func() {
		defer close(s.requests)

		log.Info().Msg("Starting main loop for server ...")
		for {
			select {
			case <-s.stopch:
				log.Warn().Msg("Stop signal received, exiting main loop")
				return
			case request := <-s.socket.RecvChan:
				// Check if the channel is closed
				// log.Debug().Msgf("Received request from client: %v", request)
				s.requests <- request
			}
		}
	}()

	return nil
}

func (s *Server) Stop() {
	log.Info().Msg("Initiating shutdown sequence for server ...")

	// Send stop signal to the main loop
	s.stopch <- struct{}{}
	close(s.stopch)

	// Wait for all workers to finish
	s.waitgroup.Wait()

	// Shutdown the socket
	if s.socket != nil {
		s.socket.Destroy()
	}

	log.Info().Msg("Server shutdown complete.")
}

func (s *Server) worker(number int) {
	defer func() {
		if r := recover(); r != nil {
			log.Error().Msgf("Panic recovered in worker %d: %v", number, r)
		}
	}()

	for request := range s.requests {
		log.Debug().Msgf("Received request from client (worker: %d, size: %d, identity: %v)", number, len(request[1]), request[0])

		// Parse the request
		req, identity, err := s.parseRequest(request)
		if err != nil {
			log.Error().Err(err).Msg("Failed to parse request")

			// Send error encoded
			encoded := s.generateErrorResponse(identity, req.ID, req.Type, fmt.Errorf("invalid request format: %w", err))
			s.socket.SendChan <- encoded
			continue
		}

		// Process the request
		response, err := s.processRequest(req)
		if err != nil {
			log.Error().Err(err).Msg("Failed to process request")

			// Send error encoded
			encoded := s.generateErrorResponse(identity, req.ID, req.Type, fmt.Errorf("failed to process request: %w", err))
			s.socket.SendChan <- encoded
			continue
		}

		// Send the response
		encoded := s.generateSuccessResponse(identity, req.ID, req.Type, response)
		s.socket.SendChan <- encoded
	}
}
