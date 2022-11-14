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

func (h *barrelHandler) UpdateBarrelById(ctx echo.Context) error {
	req := &rest_app.UpdateBarrelByIdRequest{}
	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, &rest_app.ResponseBodyInfo{
			Code:    status.INVALID_PARAM,
			Message: "invalid request",
		})
	}

	updateRes, err := h.barrelClient.UpdateBarrelById(ctx.Request().Context(), barrel.UpdateBarrelByIdParam{
		Id:       ctx.Param("id"),
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

	return ctx.JSON(http.StatusOK, &rest_app.UpdateBarrelByIdResponse{
		Code:    updateRes.Success.Code,
		Message: updateRes.Success.Message,
		Data: rest_app.UpdateBarrelByIdData{
			Id:        updateRes.Id,
			Code:      updateRes.Code,
			Name:      updateRes.Name,
			Provider:  updateRes.Provider,
			Status:    updateRes.Status,
			CreatedAt: updateRes.CreatedAt.UnixMilli(),
			UpdatedAt: updateRes.UpdatedAt.UnixMilli(),
		},
	})
}

func (h *barrelHandler) SearchBarrel(ctx echo.Context) error {
	req := &rest_app.SearchBarrelRequest{}
	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, &rest_app.ResponseBodyInfo{
			Code:    status.INVALID_PARAM,
			Message: "invalid request",
		})
	}

	keyword := ""
	if req.Keyword != nil {
		keyword = *req.Keyword
	}

	statuses := []string{}
	if req.Filter != nil {
		if req.Filter.StatusIn != nil {
			for _, status := range *req.Filter.StatusIn {
				statuses = append(statuses, string(status))
			}
		}
	}

	providers := []string{}
	if req.Filter != nil {
		if req.Filter.ProviderIn != nil {
			for _, provider := range *req.Filter.ProviderIn {
				providers = append(providers, string(provider))
			}
		}
	}

	totalItems := int32(0)
	page := int64(0)
	if req.Pagination != nil {
		totalItems = req.Pagination.TotalItems
		page = req.Pagination.Page
	}

	searchRes, err := h.barrelClient.SearchBarrel(ctx.Request().Context(), barrel.SearchBarrelParam{
		Keyword:    keyword,
		Statuses:   statuses,
		Providers:  providers,
		TotalItems: totalItems,
		Page:       page,
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

	items := []rest_app.SearchBarrelItem{}
	for _, searchItem := range searchRes.Items {
		var updatedAt *int64
		if searchItem.UpdatedAt != nil {
			updated := searchItem.UpdatedAt.UnixMilli()
			updatedAt = &updated
		}

		items = append(items, rest_app.SearchBarrelItem{
			Id:        searchItem.Id,
			Code:      searchItem.Code,
			Name:      searchItem.Name,
			Provider:  searchItem.Provider,
			Status:    searchItem.Status,
			CreatedAt: searchItem.CreatedAt.UnixMilli(),
			UpdatedAt: updatedAt,
		})
	}

	return ctx.JSON(http.StatusOK, &rest_app.SearchBarrelResponse{
		Code:    searchRes.Success.Code,
		Message: searchRes.Success.Message,
		Data: rest_app.SearchBarrelData{
			Items: items,
			Summary: rest_app.SearchBarrelSummary{
				Page:       searchRes.Summary.Page,
				TotalItems: searchRes.Summary.TotalItems,
			},
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
