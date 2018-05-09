package services

import (
	"math"
	"testing"
	"time"
	"xungewang.cn/bsp/apis"
	"xungewang.cn/bsp/app"
	"xungewang.cn/bsp/models"
)

type mockPositionRepo struct {
	foundStations  []models.Station
	unknownSignals []apis.Signal
}

func (repo *mockPositionRepo) FindStations(ctx app.RequestScope, signals map[string]apis.Signal) ([]models.Station, error) {
	return repo.foundStations, nil
}
func (repo *mockPositionRepo) RecordUnknownSignals(ctx app.RequestScope, signals []apis.Signal) {
	repo.unknownSignals = signals
}

func TestPositionService_ComputePosition(t *testing.T) {
	repo := mockPositionRepo{
		foundStations: []models.Station{
			{Id: "0-32838-60122", Lat: 30.732796, Lng: 103.962357},
		},
	}
	positionService := NewPositionService(&repo)

	var request apis.PositionRequest = []apis.Signal{
		{Mnc: "0", Lac: "32838", Cid: "60122", Strength: -78},
		{Mnc: "0", Lac: "32838", Cid: "60123", Strength: -79}, // won't find this
	}
	if r, err := positionService.ComputePosition(nil, &request); err == nil {
		// wait for the coroutine to finish
		time.Sleep(200 * time.Millisecond)

		// check the unknown signals
		unknown := repo.unknownSignals
		if len(unknown) != 1 && unknown[0].Cid != "60123" {
			t.Errorf("unexpected unknown signal: %v", unknown)
		}

		// check the result
		if !(math.Abs(math.Dim(r.Lat, 30.732796)) < EPSILON &&
			math.Abs(math.Dim(r.Lng, 103.962357)) < EPSILON) {
			t.Error("lat/lng not expected")
		}
	} else {
		t.Error(err)
	}
}
