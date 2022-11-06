package rest_handler

import (
	"net/http"

	rest_app "github.com/go-seidon/chariot/generated/rest-app"
	"github.com/go-seidon/provider/status"
	"github.com/labstack/echo/v4"
)

type basicHandler struct {
	config *BasicConfig
}

func (h *basicHandler) GetAppInfo(ctx echo.Context) error {
	res := &rest_app.GetAppInfoResponse{
		Code:    status.ACTION_SUCCESS,
		Message: "success get app info",
		Data: rest_app.GetAppInfoData{
			AppName:    h.config.AppName,
			AppVersion: h.config.AppVersion,
		},
	}
	return ctx.JSON(http.StatusOK, res)
}

type BasicParam struct {
	Config *BasicConfig
}

type BasicConfig struct {
	AppName    string
	AppVersion string
}

func NewBasic(p BasicParam) *basicHandler {
	return &basicHandler{
		config: p.Config,
	}
}
