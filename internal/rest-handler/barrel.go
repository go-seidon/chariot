package rest_handler

import (
	"net/http"

	rest_app "github.com/go-seidon/chariot/generated/rest-app"
	"github.com/go-seidon/chariot/internal/barrel"
	"github.com/go-seidon/provider/status"
	"github.com/labstack/echo/v4"
)

type barrelHandler struct {
	barrelClient barrel.Barrel
}

func (h *barrelHandler) CreateBarrel(ctx echo.Context) error {
	req := &rest_app.CreateBarrelRequest{}
	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, &rest_app.ResponseBodyInfo{
			Code:    status.INVALID_PARAM,
			Message: "invalid request",
		})
	}

	createRes, err := h.barrelClient.CreateBarrel(ctx.Request().Context(), barrel.CreateBarrelParam{
		Code:     req.Code,
		Name:     req.Name,
		Provider: string(req.Provider),
		Status:   string(req.Status),
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

	return ctx.JSON(http.StatusCreated, &rest_app.CreateBarrelResponse{
		Code:    createRes.Success.Code,
		Message: createRes.Success.Message,
		Data: rest_app.CreateBarrelData{
			Id:        createRes.Id,
			Code:      createRes.Code,
			Name:      createRes.Name,
			Provider:  createRes.Provider,
			Status:    createRes.Status,
			CreatedAt: createRes.CreatedAt.UnixMilli(),
		},
	})
}

func (h *barrelHandler) GetBarrelById(ctx echo.Context) error {
	findRes, err := h.barrelClient.FindBarrelById(ctx.Request().Context(), barrel.FindBarrelByIdParam{
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

	return ctx.JSON(http.StatusOK, &rest_app.GetBarrelByIdResponse{
		Code:    findRes.Success.Code,
		Message: findRes.Success.Message,
		Data: rest_app.GetBarrelByIdData{
			Id:        findRes.Id,
			Code:      findRes.Code,
			Name:      findRes.Name,
			Provider:  findRes.Provider,
			Status:    findRes.Status,
			CreatedAt: findRes.CreatedAt.UnixMilli(),
			UpdatedAt: updatedAt,
		},
	})
}

type BarrelParam struct {
	Barrel barrel.Barrel
}

func NewBarrel(p BarrelParam) *barrelHandler {
	return &barrelHandler{
		barrelClient: p.Barrel,
	}
}
