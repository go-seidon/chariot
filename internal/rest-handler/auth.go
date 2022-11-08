package rest_handler

import (
	"net/http"

	rest_app "github.com/go-seidon/chariot/generated/rest-app"
	"github.com/go-seidon/chariot/internal/auth"
	"github.com/go-seidon/provider/status"
	"github.com/labstack/echo/v4"
)

type authHandler struct {
	authClient auth.AuthClient
}

func (h *authHandler) CreateClient(ctx echo.Context) error {
	req := &rest_app.CreateAuthClientRequest{}
	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, &rest_app.ResponseBodyInfo{
			Code:    status.INVALID_PARAM,
			Message: "invalid request",
		})
	}

	createRes, err := h.authClient.CreateClient(ctx.Request().Context(), auth.CreateClientParam{
		ClientId:     req.ClientId,
		ClientSecret: req.ClientSecret,
		Name:         req.Name,
		Type:         string(req.Type),
		Status:       string(req.Status),
	})
	if err != nil {
		switch err.Code {
		case status.INVALID_PARAM:
			return echo.NewHTTPError(http.StatusBadRequest, &rest_app.ResponseBodyInfo{
				Code:    err.Code,
				Message: err.Message,
			})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, &rest_app.ResponseBodyInfo{
			Code:    err.Code,
			Message: err.Message,
		})
	}

	return ctx.JSON(http.StatusCreated, &rest_app.CreateAuthClientResponse{
		Code:    createRes.Success.Code,
		Message: createRes.Success.Message,
		Data: rest_app.CreateAuthClientData{
			Id:        createRes.Id,
			Name:      createRes.Name,
			Type:      createRes.Type,
			Status:    createRes.Status,
			ClientId:  createRes.ClientId,
			CreatedAt: createRes.CreatedAt.UnixMilli(),
		},
	})
}

func (h *authHandler) GetClientById(ctx echo.Context) error {
	findRes, err := h.authClient.FindClientById(ctx.Request().Context(), auth.FindClientByIdParam{
		Id: ctx.Param("id"),
	})
	if err != nil {
		switch err.Code {
		case status.INVALID_PARAM:
			return echo.NewHTTPError(http.StatusBadRequest, &rest_app.ResponseBodyInfo{
				Code:    err.Code,
				Message: err.Message,
			})
		case status.RESOURCE_NOTFOUND:
			return echo.NewHTTPError(http.StatusNotFound, &rest_app.ResponseBodyInfo{
				Code:    err.Code,
				Message: err.Message,
			})
		}

		return echo.NewHTTPError(http.StatusInternalServerError, &rest_app.ResponseBodyInfo{
			Code:    err.Code,
			Message: err.Message,
		})
	}

	var updatedAt *int64
	if findRes.UpdatedAt != nil {
		updatedDate := findRes.UpdatedAt.UnixMilli()
		updatedAt = &updatedDate
	}

	return ctx.JSON(http.StatusOK, &rest_app.GetAuthClientByIdResponse{
		Code:    findRes.Success.Code,
		Message: findRes.Success.Message,
		Data: rest_app.GetAuthClientByIdData{
			Id:        findRes.Id,
			Name:      findRes.Name,
			Type:      findRes.Type,
			Status:    findRes.Status,
			ClientId:  findRes.ClientId,
			CreatedAt: findRes.CreatedAt.UnixMilli(),
			UpdatedAt: updatedAt,
		},
	})
}

type AuthParam struct {
	AuthClient auth.AuthClient
}

func NewAuth(p AuthParam) *authHandler {
	return &authHandler{
		authClient: p.AuthClient,
	}
}
