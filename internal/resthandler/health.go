package resthandler

import (
	"net/http"

	"github.com/go-seidon/chariot/api/restapp"
	"github.com/go-seidon/chariot/internal/healthcheck"
	"github.com/labstack/echo/v4"
)

type healthHandler struct {
	healthClient healthcheck.HealthCheck
}

func (h *healthHandler) CheckHealth(ctx echo.Context) error {
	health, err := h.healthClient.Check(ctx.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, &restapp.ResponseBodyInfo{
			Code:    err.Code,
			Message: err.Message,
		})
	}

	details := restapp.CheckHealthData_Details{
		AdditionalProperties: map[string]restapp.CheckHealthDetail{},
	}
	for _, item := range health.Items {
		details.AdditionalProperties[item.Name] = restapp.CheckHealthDetail{
			Name:      item.Name,
			Status:    item.Status,
			Error:     item.Error,
			CheckedAt: item.CheckedAt.UnixMilli(),
		}
	}

	return ctx.JSON(http.StatusOK, &restapp.CheckHealthResponse{
		Code:    health.Success.Code,
		Message: health.Success.Message,
		Data: restapp.CheckHealthData{
			Details: details,
			Status:  health.Status,
		},
	})
}

type HealthParam struct {
	HealthClient healthcheck.HealthCheck
}

func NewHealth(p HealthParam) *healthHandler {
	return &healthHandler{
		healthClient: p.HealthClient,
	}
}
