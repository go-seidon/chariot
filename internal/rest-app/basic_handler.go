package rest_app

import (
	"net/http"

	rest_v1 "github.com/go-seidon/chariot/generated/rest-v1"
	"github.com/go-seidon/provider/status"
	"github.com/labstack/echo/v4"
)

type basicHandler struct {
	config *RestAppConfig
}

func (h *basicHandler) GetAppInfo(ctx echo.Context) error {
	res := &rest_v1.GetAppInfoResponse{
		Code:    status.ACTION_SUCCESS,
		Message: "success get app info",
		Data: rest_v1.GetAppInfoData{
			AppName:    h.config.AppName,
			AppVersion: h.config.AppVersion,
		},
	}
	return ctx.JSON(http.StatusOK, res)
}

type BasicHandlerParam struct {
	Config *RestAppConfig
}

func NewBasicHandler(p BasicHandlerParam) *basicHandler {
	return &basicHandler{
		config: p.Config,
	}
}
