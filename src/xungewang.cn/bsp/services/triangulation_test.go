package services

import (
	"math"
	"sort"
	"testing"
	"xungewang.cn/bsp/models"
)

const EPSILON = 1e-6

func TestDoComputePosition_Normal(t *testing.T) {
	stations := []*signalAwareStation{
		{&models.Station{Id: "0-32838-60122", Lat: 30.732796, Lng: 103.962357}, -77},
		{&models.Station{Id: "0-32838-60123", Lat: 30.734688, Lng: 103.961433}, -83},
		{&models.Station{Id: "0-32838-36861", Lat: 30.730850, Lng: 103.965279}, -88},
		{&models.Station{Id: "0-32838-60125", Lat: 30.732283, Lng: 103.961327}, -95},
		{&models.Station{Id: "0-32838-60124", Lat: 30.732937, Lng: 103.965981}, -96},
		{&models.Station{Id: "0-32838-36863", Lat: 30.732002, Lng: 103.958771}, -97},
	}

	if r, err := doComputePosition(stations); err == nil {
		if !(math.Abs(math.Dim(r.Lat, 30.732924)) < EPSILON &&
			math.Abs(math.Dim(r.Lng, 103.962488)) < EPSILON) {
			t.Error("lat/lng not expected")
		}
	} else {
		t.Error(err)
	}
}

func TestFindClosestStations(t *testing.T) {
	stations := []*signalAwareStation{
		{&models.Station{Id: "0-32838-60122", Lat: 30.732796, Lng: 103.962357}, -77},
		{&models.Station{Id: "0-32838-60123", Lat: 30.734688, Lng: 103.961433}, -83},
		{&models.Station{Id: "0-32838-36861", Lat: 30.730850, Lng: 104.965279}, -88},
		{&models.Station{Id: "0-32838-60125", Lat: 30.732283, Lng: 103.961327}, -95},
	}

	closetStations := findClosestStations(stations)

	// station with id '0-32838-36861' should be excluded as it's too far away from other stations
	closetStationIds := make([]string, len(closetStations), len(closetStations))
	for i, station := range closetStations {
		closetStationIds[i] = station.Id
	}
	sort.Strings(closetStationIds)

	if len(closetStationIds) != 3 || closetStationIds[0] != "0-32838-60122" ||
		closetStationIds[1] != "0-32838-60123" || closetStationIds[2] != "0-32838-60125" {
		t.Error("station(0-32838-36861) should be excluded")
	}
}
