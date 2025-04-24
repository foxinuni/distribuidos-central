package controllers

import (
	"context"
	"fmt"

	"github.com/foxinuni/distribuidos-central/internal/models"
	"github.com/foxinuni/distribuidos-central/internal/services"
	"github.com/rs/zerolog/log"
)

type AllocationsController struct {
	service services.AllocationService
}

func NewAllocationsController(service services.AllocationService) *AllocationsController {
	return &AllocationsController{
		service: service,
	}
}

func (c *AllocationsController) Allocate(body interface{}) (interface{}, error) {
	req, ok := body.(*models.AllocateRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request format: expected *models.AllocateRequest, got %T", body)
	}

	log.Info().Msgf("Received AllocateRequest: %+v", req)
	return c.service.Allocate(context.Background(), req)
}
