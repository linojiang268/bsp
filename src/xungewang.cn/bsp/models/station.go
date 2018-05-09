package models

import "github.com/go-ozzo/ozzo-validation"

// Station represents an base station.
type Station struct {
	Id  string  `db:"id"`
	Lat float64 `db:"lat"`
	Lng float64 `db:"lng"`
}

// Validate validates the Station fields
func (station Station) Validate() error {
	return validation.ValidateStruct(&station,
		validation.Field(&station.Id, validation.Required),
	)
}
