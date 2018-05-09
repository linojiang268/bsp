package app

import (
	log "github.com/Sirupsen/logrus"
	"github.com/go-ozzo/ozzo-routing"
	"github.com/go-ozzo/ozzo-routing/fault"
	"github.com/go-ozzo/ozzo-validation"
	"net/http"
	"xungewang.cn/bsp/errors"
)

func Init() routing.Handler {
	return func(ctx *routing.Context) error {
		scope := newRequestScope(ctx.Request)
		ctx.Set("Context", scope)

		// fault.Recovery() logs too much, we'll have it reduced
		recovery(log.Errorf, convertError)(ctx)
		return nil
	}
}

func recovery(logf fault.LogFunc, errorf ...fault.ConvertErrorFunc) routing.Handler {
	handlePanic := fault.PanicHandler(logf) // log panic errors
	return func(c *routing.Context) error {
		if err := handlePanic(c); err != nil {
			if len(errorf) > 0 {
				err = errorf[0](c, err)
			}

			// errors returned by application or middleware won't be logged

			writeError(c, err)
			c.Abort()
		}
		return nil
	}
}

func convertError(ctx *routing.Context, err error) error {
	switch err.(type) {
	case *errors.APIError:
		return err
	case validation.Errors:
		return errors.InvalidData(err.(validation.Errors))
	case routing.HTTPError:
		switch err.(routing.HTTPError).StatusCode() {
		case http.StatusNotFound:
			return errors.NotFound(err.Error())
		}
	}

	log.Errorf("error while handling request(%s): %s", ctx.Request.RequestURI, err)
	return errors.InternalServerError(err)
}

// writeError writes the error to the response.
// If the error implements HTTPError, it will set the HTTP status as the result of the StatusCode() call of the error.
// Otherwise, the HTTP status will be set as http.StatusInternalServerError.
//
// Copied from fault package.
func writeError(c *routing.Context, err error) {
	if httpError, ok := err.(routing.HTTPError); ok {
		c.Response.WriteHeader(httpError.StatusCode())
	} else {
		c.Response.WriteHeader(http.StatusInternalServerError)
	}
	c.Write(err)
}

func GetRequestScope(ctx *routing.Context) RequestScope {
	return ctx.Get("Context").(RequestScope)
}
