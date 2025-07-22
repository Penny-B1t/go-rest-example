package handlers

import (
	errors2 "errors"

	"go-rest-example/internal/db"
	"go-rest-example/internal/logger"
)

type DevicesHandler struct {
	dvRepo db.DevicesDataService
	logger *logger.AppLogger
}

func NewDevicesHandler(lgr *logger.AppLogger,dvRepo db.DevicesDataService )(*DevicesHandler, error){
	if lgr == nil || dvRepo == nil {
		return nil, errors2.New("missing required parameters to create orders handler")
	}

	return &DevicesHandler{dvRepo: dvRepo, logger: lgr}, nil
}

