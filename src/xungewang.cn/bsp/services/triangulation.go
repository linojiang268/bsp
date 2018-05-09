package services

import (
	"math"
	"xungewang.cn/bsp/apis"
	"xungewang.cn/bsp/errors"
)

// actual method to calculate the position
func doComputePosition(stations []*signalAwareStation) (*apis.PositionResult, error) {
	// stations can diverge, find closest ones (in other words, eliminate those far away)
	stations = findClosestStations(stations)
	if len(stations) == 0 {
		return nil, errors.NotFound("no suitable stations")
	}

	return triangulate(stations), nil
}

const (
	// standard derivation threshold to detect non-close stations
	stdDevThreshold = 0.03

	// degree to radians
	d2R = math.Pi / 180.0
)

// find stations that live close
func findClosestStations(stations []*signalAwareStation) []*signalAwareStation {
	for len(stations) > 1 { // compute if and only if there's more than one station
		// by close, it means the standard derivation of stations, precisely, their latitudes and longitudes,
		// are less than the defined threshold. To calculate standard derivation, first compute average value.
		avgLat, avgLng := 0.0, 0.0
		for _, station := range stations {
			avgLat += station.Lat
			avgLng += station.Lng
		}
		avgLat /= float64(len(stations))
		avgLng /= float64(len(stations))

		// calculate standard derivation, and track the one far from the flock.
		// maxDistance: the max distance from the center denoted by (avgLat, avgLng)
		// maxDistanceIndex: the index (zero based) of the most far away station, which will be eliminated
		// sum: sum of square of distances, used for calculating standard deviation later
		maxDistance, maxDistanceIndex, sum := 0.0, 0, 0.0
		for index, station := range stations {
			distance := math.Pow(station.Lat-avgLat, 2) + math.Pow(station.Lng-avgLng, 2)
			if distance > maxDistance {
				maxDistance = distance
				maxDistanceIndex = index
			}

			sum += distance
		}

		if math.Sqrt(sum/float64(len(stations))) < stdDevThreshold { // stations are close enough
			return stations
		}

		// eliminate the station far away, and check again
		stations = append(stations[0:maxDistanceIndex], stations[maxDistanceIndex+1:]...)
	}

	return stations
}

func triangulate(stations []*signalAwareStation) *apis.PositionResult {
	lats := make([]float64, len(stations))
	lngs := make([]float64, len(stations))
	distanceWeights := make([]float64, len(stations))
	distanceProduct, distanceSum := 1.0, 0.0

	for i, station := range stations {
		lats[i] = station.Lat * d2R
		lngs[i] = station.Lng * d2R

		// the method to calculate distance weight written in C language is:
		//
		//   freq := 1000
		//   distanceWeights[i] = math.Pow(10, ((130 + station.SignalStrength - 20 * math.Log10(freq)) / 20)
		//
		// The value of freq variable never changes. In that case, math.Log10(freq) will always be 3,
		// and the whole express '((130 + station.SignalStrength - 20 * math.Log10(freq)) / 20' can be simplified as
		// '3.5+station.SignalStrength/20'
		distanceWeights[i] = math.Pow(10, 3.5+station.SignalStrength/20)

		distanceProduct *= distanceWeights[i]
		distanceSum += distanceWeights[i]
	}

	x, y, z := 0.0, 0.0, 0.0
	for i := range stations {
		x += math.Cos(lats[i]) * math.Cos(lngs[i]) * distanceWeights[i]
		y += math.Cos(lats[i]) * math.Sin(lngs[i]) * distanceWeights[i]
		z += math.Sin(lats[i]) * distanceWeights[i]
	}

	x /= distanceSum
	y /= distanceSum
	z /= distanceSum

	lat := math.Atan(z/math.Sqrt(x*x+y*y)) / d2R
	lng := math.Atan(y/x) / d2R
	if lng < 0 {
		lng += 180
	}

	return apis.NewPositionResult(lat, lng)
}
