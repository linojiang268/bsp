package apis

import (
	"fmt"
	"github.com/go-ozzo/ozzo-dbx"
	"github.com/go-ozzo/ozzo-routing"
	"github.com/go-ozzo/ozzo-routing/content"
	"github.com/go-ozzo/ozzo-validation"
	"strings"
	"xungewang.cn/bsp/app"
	"xungewang.cn/bsp/errors"
)

type (
	// contract position related behaviors
	positionService interface {
		ComputePosition(ctx app.RequestScope, request *PositionRequest) (*PositionResult, error)
	}

	PositionResult struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	}

	// wrapper of positionService
	positionResource struct {
		service positionService
	}
)

func SetupPositionRouter(group *routing.RouteGroup, db *dbx.DB, service positionService) {
	r := &positionResource{service}

	group.Use(
		content.TypeNegotiator(content.JSON),
		app.DbAware(db),
	)

	group.Post("/position", r.computePosition)
}

type PositionRequest []Signal

func (request *PositionRequest) Validate() error {
	if err := validation.Validate(*request, validation.NilOrNotEmpty); err != nil {
		return err
	}

	// validate each signal in the request
	for _, signal := range *request {
		if err := signal.Validate(); err != nil {
			return err
		}
	}

	return nil
}

type Signal struct {
	// Mobile Network Code
	Mnc string `json:"mnc"`

	// Location Area Code is a unique number of current location area.
	Lac string `json:"lac"`

	// a generally unique number used to identify each Base transceiver station (BTS)
	// or sector of a BTS within a Location area code
	Cid string `json:"cid"`

	// signal strength refers to the transmitter power output as received by a reference antenna
	// at a distance from the transmitting antenna.
	Strength float64 `json:"str"`
}

func (signal *Signal) Validate() error {
	return validation.ValidateStruct(signal,
		validation.Field(&signal.Mnc, validation.Required),
		validation.Field(&signal.Lac, validation.Required),
		validation.Field(&signal.Cid, validation.Required),
		// single strength (or in short, RSSI - Received Signal Strength Indication), in theory, should
		// range from -113 to -51. But in practise, power like -120 can be received. So we enlarge the
		// range from -150 to 0 to tolerate. The RSSI is measured in dBm, which should be negative.
		validation.Field(&signal.Strength, validation.Min(-150.0), validation.Max(0.0)))
}

func (signal *Signal) String() string {
	return fmt.Sprintf("(lac:%s, cid:%s, str:%f)", signal.Lac, signal.Cid, signal.Strength)
}

func (r *positionResource) computePosition(ctx *routing.Context) error {
	var request *PositionRequest
	if err := ctx.Read(&request); err != nil {
		if v := ctx.Request.Header.Get("Content-Type"); !strings.HasPrefix(v, "application/json") {
			return errors.SimpleInvalidData(fmt.Sprintf("request not acceptable. given Content-Type is %s", v))
		}

		return errors.SimpleInvalidData("request not acceptable. pay special attention on fields type and value. For " +
			"example, mnc, lac and cid should be of string type, and str double.")
	}

	if err := request.Validate(); err != nil {
		return err
	}

	if pos, err := r.service.ComputePosition(app.GetRequestScope(ctx), request); err != nil {
		return err
	} else {
		return ctx.Write(pos)
	}
}

func (position *PositionResult) String() string {
	return fmt.Sprintf("(lat: %f, lng: %f)", position.Lat, position.Lng)
}

func NewPositionResult(lat float64, lng float64) *PositionResult {
	return &PositionResult{lat, lng}
}

