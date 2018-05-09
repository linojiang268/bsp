package app

import (
	"github.com/go-ozzo/ozzo-dbx"
	"net/http"
)

type RequestScope interface {
	// Tx returns the currently active database transaction that can be used for DB query purpose
	Db() *dbx.DB
	// SetTx sets the database transaction
	SetDb(db *dbx.DB)
}

type requestScope struct {
	db *dbx.DB // the currently active transaction
}

func (scope *requestScope) Db() *dbx.DB {
	return scope.db
}

func (scope *requestScope) SetDb(db *dbx.DB) {
	scope.db = db
}

// newRequestScope creates a new RequestScope with the current request information.
func newRequestScope(request *http.Request) RequestScope {
	return &requestScope{}
}
