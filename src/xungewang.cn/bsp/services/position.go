package services

import (
	log "github.com/Sirupsen/logrus"
	"strings"
	"xungewang.cn/bsp/apis"
	"xungewang.cn/bsp/app"
	"xungewang.cn/bsp/errors"
	"xungewang.cn/bsp/models"
	"xungewang.cn/bsp/repos"
)

// positionService should implements apis.positionService interface
type (
	positionService struct {
		repo repos.PositionRepo
	}

	signalAwareStation struct {
		*models.Station

		// strength of related signal
		SignalStrength float64
	}
)

func newSignalAwareStation(station *models.Station, signalStrength float64) *signalAwareStation {
	return &signalAwareStation{station, signalStrength}
}

func (service *positionService) ComputePosition(ctx app.RequestScope, request *apis.PositionRequest) (*apis.PositionResult, error) {
	// In order to attach additional information (e.g., signal) to stations found,
	// we use map as the underlying data structure with station id as its key.
	signals := make(map[string]apis.Signal)
	for _, signal := range *request {
		signals[buildStationId(signal)] = signal
	}

	stations, err := service.repo.FindStations(ctx, signals)
	if err != nil {
		return nil, err
	}

	// not all stations requested found
	if len(stations) != len(signals) {
		log.Debug("not all requested signal has a corresponding station")
		go service.recordUnknownSignals(ctx, stations, signals)
		if len(stations) == 0 { // to bad the request is
			return nil, errors.NotFound("no stations matched")
		}
	}

	// wrap stations with corresponding signal
	signalAwareStations := make([]*signalAwareStation, len(stations))
	for i, station := range stations {
		// signals[station.Id] should always be ok
		signalAwareStations[i] = newSignalAwareStation(&stations[i], signals[station.Id].Strength)
	}

	return doComputePosition(signalAwareStations)
}

func (service *positionService) recordUnknownSignals(ctx app.RequestScope, foundStations []models.Station, requested map[string]apis.Signal) {
	unknowns := make([]apis.Signal, 0, len(requested))
	for id, signal := range requested {
		// check whether it's found
		found := false
		for _, station := range foundStations {
			if station.Id == id {
				found = true
				break
			}
		}

		if !found {
			unknowns = append(unknowns, signal)
		}
	}

	// persist
	log.Debugf("persist unknown signals: %s", unknowns)
	service.repo.RecordUnknownSignals(ctx, unknowns)
}

func buildStationId(signal apis.Signal) string {
	return strings.Join([]string{signal.Mnc, signal.Lac, signal.Cid}, "-")
}

// NewPositionService create an instance of positionService
func NewPositionService(repo repos.PositionRepo) *positionService {
	return &positionService{repo}
}
