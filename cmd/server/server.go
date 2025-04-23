package main

import (
	"fmt"
	"sync"

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

	// external
	socket     *goczmq.Sock
	serializer services.ModelSerializer
}

func NewController(serializer services.ModelSerializer, options ...ServerOptions) *Server {
	server := &Server{
		port:       5555,
		workers:    10,
		requests:   make(chan [][]byte),
		stopch:     make(chan struct{}),
		routes:     make(map[string]RouteHandler),
		serializer: serializer,
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
	s.socket, _ = goczmq.NewRouter(fmt.Sprintf("tcp://*:%d", s.port))
	if s.socket == nil {
		return fmt.Errorf("failed to create socket")
	}

	// Start the workers
	for i := 0; i < s.workers; i++ {
		s.waitgroup.Add(1)

		go func() {
			defer s.waitgroup.Done()
			s.worker()
		}()
	}

	// Start the main loop
	go func() {
		defer close(s.requests)

		log.Info().Msg("Starting main loop for server ...")
		for {
			select {
			case <-s.stopch:
				log.Info().Msg("Stop signal received, exiting main loop")
				return
			default:
				request, error := s.socket.RecvMessage()
				if error != nil {
					log.Error().Err(error).Msg("Failed to receive message")
					continue
				}

				// Check if the channel is closed
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

func (s *Server) worker() {
	for request := range s.requests {
		// Parse the request
		req, identity, err := s.parseRequest(request)
		if err != nil {
			log.Error().Err(err).Msg("Failed to parse request")

			// Send error response
			response := s.generateErrorResponse(identity, req.ID, req.Type, fmt.Errorf("invalid request format: %w", err))
			s.socket.SendMessage(response)
			continue
		}

		// Process the request
		response, err := s.processRequest(req)
		if err != nil {
			log.Error().Err(err).Msg("Failed to process request")

			// Send error response
			response := s.generateErrorResponse(identity, req.ID, req.Type, fmt.Errorf("failed to process request: %w", err))
			s.socket.SendMessage(response)
			continue
		}

		// Send the response
		encoded := s.generateSuccessResponse(identity, req.ID, req.Type, response)
		if err := s.socket.SendMessage(encoded); err != nil {
			log.Error().Err(err).Msg("Failed to send response")
		}
	}
}
