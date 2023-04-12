package resthandler

import (
	"net/http"

	"github.com/go-seidon/chariot/api/restapp"
	"github.com/go-seidon/chariot/internal/service"
	"github.com/go-seidon/provider/status"
	"github.com/go-seidon/provider/typeconv"
	"github.com/labstack/echo/v4"
)

type barrelHandler struct {
	barrelClient service.Barrel
}

func (h *barrelHandler) CreateBarrel(ctx echo.Context) error {
	req := &restapp.CreateBarrelRequest{}
	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, &restapp.ResponseBodyInfo{
			Code:    status.INVALID_PARAM,
			Message: "invalid request",
		})
	}

	createRes, err := h.barrelClient.CreateBarrel(ctx.Request().Context(), service.CreateBarrelParam{
		Code:     req.Code,
		Name:     req.Name,
		Provider: string(req.Provider),
		Status:   string(req.Status),
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

	return ctx.JSON(http.StatusCreated, &restapp.CreateBarrelResponse{
		Code:    createRes.Success.Code,
		Message: createRes.Success.Message,
		Data: restapp.CreateBarrelData{
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
	findRes, err := h.barrelClient.FindBarrelById(ctx.Request().Context(), service.FindBarrelByIdParam{
		Id: ctx.Param("id"),
	})
	if err != nil {
		switch err.Code {
		case status.INVALID_PARAM:
			return echo.NewHTTPError(http.StatusBadRequest, &restapp.ResponseBodyInfo{
				Code:    err.Code,
				Message: err.Message,
			})
		case status.RESOURCE_NOTFOUND:
			return echo.NewHTTPError(http.StatusNotFound, &restapp.ResponseBodyInfo{
				Code:    err.Code,
				Message: err.Message,
			})
		}

		return echo.NewHTTPError(http.StatusInternalServerError, &restapp.ResponseBodyInfo{
			Code:    err.Code,
			Message: err.Message,
		})
	}

	var updatedAt *int64
	if findRes.UpdatedAt != nil {
		updatedAt = typeconv.Int64(findRes.UpdatedAt.UnixMilli())
	}

	return ctx.JSON(http.StatusOK, &restapp.GetBarrelByIdResponse{
		Code:    findRes.Success.Code,
		Message: findRes.Success.Message,
		Data: restapp.GetBarrelByIdData{
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
	req := &restapp.UpdateBarrelByIdRequest{}
	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, &restapp.ResponseBodyInfo{
			Code:    status.INVALID_PARAM,
			Message: "invalid request",
		})
	}

	updateRes, err := h.barrelClient.UpdateBarrelById(ctx.Request().Context(), service.UpdateBarrelByIdParam{
		Id:       ctx.Param("id"),
		Code:     req.Code,
		Name:     req.Name,
		Provider: string(req.Provider),
		Status:   string(req.Status),
	})
	if err != nil {
		switch err.Code {
		case status.INVALID_PARAM:
			return echo.NewHTTPError(http.StatusBadRequest, &restapp.ResponseBodyInfo{
				Code:    err.Code,
				Message: err.Message,
			})
		case status.RESOURCE_NOTFOUND:
			return echo.NewHTTPError(http.StatusNotFound, &restapp.ResponseBodyInfo{
				Code:    err.Code,
				Message: err.Message,
			})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, &restapp.ResponseBodyInfo{
			Code:    err.Code,
			Message: err.Message,
		})
	}

	return ctx.JSON(http.StatusOK, &restapp.UpdateBarrelByIdResponse{
		Code:    updateRes.Success.Code,
		Message: updateRes.Success.Message,
		Data: restapp.UpdateBarrelByIdData{
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
	req := &restapp.SearchBarrelRequest{}
	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, &restapp.ResponseBodyInfo{
			Code:    status.INVALID_PARAM,
			Message: "invalid request",
		})
	}

	statuses := []string{}
	providers := []string{}
	if req.Filter != nil {
		if req.Filter.StatusIn != nil {
			for _, status := range *req.Filter.StatusIn {
				statuses = append(statuses, string(status))
			}
		}

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

	searchRes, err := h.barrelClient.SearchBarrel(ctx.Request().Context(), service.SearchBarrelParam{
		Keyword:    typeconv.StringVal(req.Keyword),
		Statuses:   statuses,
		Providers:  providers,
		TotalItems: totalItems,
		Page:       page,
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

	items := []restapp.SearchBarrelItem{}
	for _, searchItem := range searchRes.Items {
		var updatedAt *int64
		if searchItem.UpdatedAt != nil {
			updatedAt = typeconv.Int64(searchItem.UpdatedAt.UnixMilli())
		}

		items = append(items, restapp.SearchBarrelItem{
			Id:        searchItem.Id,
			Code:      searchItem.Code,
			Name:      searchItem.Name,
			Provider:  searchItem.Provider,
			Status:    searchItem.Status,
			CreatedAt: searchItem.CreatedAt.UnixMilli(),
			UpdatedAt: updatedAt,
		})
	}

	return ctx.JSON(http.StatusOK, &restapp.SearchBarrelResponse{
		Code:    searchRes.Success.Code,
		Message: searchRes.Success.Message,
		Data: restapp.SearchBarrelData{
			Items: items,
			Summary: restapp.SearchBarrelSummary{
				Page:       searchRes.Summary.Page,
				TotalItems: searchRes.Summary.TotalItems,
			},
		},
	})
}

type BarrelParam struct {
	Barrel service.Barrel
}

func NewBarrel(p BarrelParam) *barrelHandler {
	return &barrelHandler{
		barrelClient: p.Barrel,
	}
}
