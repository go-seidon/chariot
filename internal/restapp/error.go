package restapp

import (
	"net/http"

	"github.com/go-seidon/provider/logging"
	"github.com/go-seidon/provider/status"
	"github.com/labstack/echo/v4"
)

type ErrorHandler = func(err error, c echo.Context)

type ErrorHandlerParam struct {
	Debug  bool
	Logger logging.Logger
}

func NewErrorHandler(p ErrorHandlerParam) ErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		he, ok := err.(*echo.HTTPError)
		if ok {
			if he.Internal != nil {
				if herr, ok := he.Internal.(*echo.HTTPError); ok {
					he = herr
				}
			}
		} else {
			he = &echo.HTTPError{
				Code:    http.StatusInternalServerError,
				Message: http.StatusText(http.StatusInternalServerError),
			}
		}

		// Issue #1426
		code := he.Code
		message := he.Message
		m, ok := he.Message.(string)
		if ok {
			message = echo.Map{
				"code":    status.ACTION_FAILED,
				"message": m,
			}
			if p.Debug {
				message = echo.Map{
					"code":    status.ACTION_FAILED,
					"message": m,
					"error":   err.Error(),
				}
			}
		}

		if c.Request().Method == http.MethodHead { // Issue #608
			err = c.NoContent(he.Code)
		} else {
			err = c.JSON(code, message)
		}
		if err != nil {
			p.Logger.Error(err)
		}
	}
}
