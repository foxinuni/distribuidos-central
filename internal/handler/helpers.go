package handler

import (
	"fmt"

	"github.com/foxinuni/distribuidos-central/internal/models"
	"github.com/rs/zerolog/log"
)

func (s *Server) parseRequest(request [][]byte) (*models.Request, string, error) {
	if len(request) < 2 {
		return nil, "", fmt.Errorf("invalid request format")
	}

	// Deserialize the request
	var req models.Request
	if err := s.serializer.Decode(request[1], &req); err != nil {
		return nil, "", fmt.Errorf("failed to decode request: %w", err)
	}

	// Get the sender ID
	identity := string(request[0])

	return &req, identity, nil
}

func (s *Server) processRequest(request *models.Request) (interface{}, error) {
	// Find the handler for the request type
	handler, ok := s.routes[request.Type]
	if !ok {
		return nil, fmt.Errorf("no handler found for request type: %s", request.Type)
	}

	// Call the handler
	return handler(request.Content)
}

func (s *Server) generateErrorResponse(identity string, id int, handler string, err error) [][]byte {
	response := &models.Response{
		ID:      id,
		Type:    handler,
		Success: false,
		Error:   err.Error(),
	}

	// Serialize the response
	encoded, err := s.serializer.Encode(response)
	if err != nil {
		log.Error().Err(err).Msg("Failed to serialize error response")
		return nil
	}

	// Send the response
	return [][]byte{[]byte(identity), encoded}
}

func (s *Server) generateSuccessResponse(identity string, id int, handler string, content interface{}) [][]byte {
	response := &models.Response{
		ID:      id,
		Type:    handler,
		Success: true,
		Content: content,
	}

	// Serialize the response
	encoded, err := s.serializer.Encode(response)
	if err != nil {
		log.Error().Err(err).Msg("Failed to serialize success response")
		return nil
	}

	// Send the response
	return [][]byte{[]byte(identity), encoded}
}
