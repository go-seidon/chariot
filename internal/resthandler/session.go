package resthandler

import (
	"net/http"
	"time"

	"github.com/go-seidon/chariot/api/restapp"
	"github.com/go-seidon/chariot/internal/session"
	"github.com/go-seidon/provider/status"
	"github.com/labstack/echo/v4"
)

type sessionHandler struct {
	sessionClient session.Session
}

func (h *sessionHandler) CreateSession(ctx echo.Context) error {
	req := &restapp.CreateSessionRequest{}
	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, &restapp.ResponseBodyInfo{
			Code:    status.INVALID_PARAM,
			Message: "invalid request",
		})
	}

	features := []string{}
	for _, feature := range req.Features {
		features = append(features, string(feature))
	}

	createRes, err := h.sessionClient.CreateSession(ctx.Request().Context(), session.CreateSessionParam{
		Duration: time.Duration(req.Duration),
		Features: features,
	})
	if err != nil {
		switch err.Code {
		case status.INVALID_PARAM:
			return echo.NewHTTPError(http.StatusBadRequest, &restapp.ResponseBodyInfo{
				Code:    err.Code,
				Message: err.Message,
			})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, &restapp.ResponseBodyInfo{
			Code:    err.Code,
			Message: err.Message,
		})
	}

	return ctx.JSON(http.StatusCreated, &restapp.CreateSessionResponse{
		Code:    createRes.Success.Code,
		Message: createRes.Success.Message,
		Data: restapp.CreateSessionData{
			Token:     createRes.Token,
			ExpiresAt: createRes.ExpiresAt.UnixMilli(),
			CreatedAt: createRes.CreatedAt.UnixMilli(),
		},
	})
}

type SessionParam struct {
	Session session.Session
}

func NewSession(p SessionParam) *sessionHandler {
	return &sessionHandler{
		sessionClient: p.Session,
	}
}
