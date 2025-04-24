package controllers

import (
	"context"
	"fmt"

	"github.com/foxinuni/distribuidos-central/internal/models"
	"github.com/foxinuni/distribuidos-central/internal/services"
	"github.com/mitchellh/mapstructure"
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
	req := &models.AllocateRequest{}
	if err := mapstructure.Decode(body, req); err != nil {
		return nil, fmt.Errorf("failed to decode request: %w", err)
	}

	log.Info().Msgf("Received AllocateRequest: %+v", req)
	return c.service.Allocate(context.Background(), req)
}

func (c *AllocationsController) Confirm(body interface{}) (interface{}, error) {
	req := &models.ConfirmRequest{}
	if err := mapstructure.Decode(body, req); err != nil {
		return nil, fmt.Errorf("failed to decode request: %w", err)
	}

	log.Info().Msgf("Received ConfirmRequest: %+v", req)
	return c.service.Confirm(context.Background(), req)
}
