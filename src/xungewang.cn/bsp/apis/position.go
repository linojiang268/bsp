package apis

import (
	"encoding/json"
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

// append given signal to current request
func (request *PositionRequest) append(signal Signal) {
	*request = append(*request, signal)
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

// request extractor to extract request data from context
type extractRequest func(ctx *routing.Context, request *PositionRequest) error

func (r *positionResource) computePosition(ctx *routing.Context) error {
	// read request data from context
	var request *PositionRequest

	// manual content negotiation
	mime := ctx.Request.Header.Get("Content-Type")
	var extract extractRequest
	if strings.HasPrefix(mime, "application/json") { // body as JSON
		extract = r.extractRequestFromJSON
	} else { // treat as form submission
		request = &PositionRequest{} // no reflect, so we create it by hand
		extract = r.extractRequestFromForm
	}

	// extract request data
	if err := extract(ctx, request); err != nil {
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

func (r *positionResource) extractRequestFromJSON(ctx *routing.Context, request *PositionRequest) error {
	return ctx.Read(&request)
}

func (r *positionResource) extractRequestFromForm(ctx *routing.Context, request *PositionRequest) error {
	if err := ctx.Request.ParseForm(); err != nil {
		return err
	}

	if signals := ctx.Request.Form["signal"]; len(signals) > 0 {
		for _, signal := range signals {
			// parse signal from each text representation
			var parsed Signal
			if err := json.NewDecoder(strings.NewReader(signal)).Decode(&parsed); err != nil {
				return err
			}

			request.append(parsed)
		}

		return nil
	}

	return errors.SimpleInvalidData("no param given")
}

func (position *PositionResult) String() string {
	return fmt.Sprintf("(lat: %f, lng: %f)", position.Lat, position.Lng)
}

func NewPositionResult(lat float64, lng float64) *PositionResult {
	return &PositionResult{lat, lng}
}
