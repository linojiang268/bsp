package app

import (
	"github.com/go-ozzo/ozzo-dbx"
	"github.com/go-ozzo/ozzo-routing"
)

// DbAware returns a handler that associate request context with a DB instance.
func DbAware(db *dbx.DB) routing.Handler {
	return func(ctx *routing.Context) error {
		scope := GetRequestScope(ctx)
		scope.SetDb(db)

		return nil
	}
}
