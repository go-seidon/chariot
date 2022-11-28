package resthandler

import (
	"net/http"

	"github.com/go-seidon/chariot/generated/restapp"
	"github.com/go-seidon/chariot/internal/auth"
	"github.com/go-seidon/provider/status"
	"github.com/go-seidon/provider/typeconv"
	"github.com/labstack/echo/v4"
)

type authHandler struct {
	authClient auth.AuthClient
}

func (h *authHandler) CreateClient(ctx echo.Context) error {
	req := &restapp.CreateAuthClientRequest{}
	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, &restapp.ResponseBodyInfo{
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

	return ctx.JSON(http.StatusCreated, &restapp.CreateAuthClientResponse{
		Code:    createRes.Success.Code,
		Message: createRes.Success.Message,
		Data: restapp.CreateAuthClientData{
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

	return ctx.JSON(http.StatusOK, &restapp.GetAuthClientByIdResponse{
		Code:    findRes.Success.Code,
		Message: findRes.Success.Message,
		Data: restapp.GetAuthClientByIdData{
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

func (h *authHandler) UpdateClientById(ctx echo.Context) error {
	req := &restapp.UpdateAuthClientByIdRequest{}
	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, &restapp.ResponseBodyInfo{
			Code:    status.INVALID_PARAM,
			Message: "invalid request",
		})
	}

	updateRes, err := h.authClient.UpdateClientById(ctx.Request().Context(), auth.UpdateClientByIdParam{
		Id:       ctx.Param("id"),
		ClientId: req.ClientId,
		Name:     req.Name,
		Type:     string(req.Type),
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

	return ctx.JSON(http.StatusOK, &restapp.UpdateAuthClientByIdResponse{
		Code:    updateRes.Success.Code,
		Message: updateRes.Success.Message,
		Data: restapp.UpdateAuthClientByIdData{
			Id:        updateRes.Id,
			Name:      updateRes.Name,
			Type:      updateRes.Type,
			Status:    updateRes.Status,
			ClientId:  updateRes.ClientId,
			CreatedAt: updateRes.CreatedAt.UnixMilli(),
			UpdatedAt: updateRes.UpdatedAt.UnixMilli(),
		},
	})
}

func (h *authHandler) SearchClient(ctx echo.Context) error {
	req := &restapp.SearchAuthClientRequest{}
	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, &restapp.ResponseBodyInfo{
			Code:    status.INVALID_PARAM,
			Message: "invalid request",
		})
	}

	statuses := []string{}
	if req.Filter != nil {
		if req.Filter.StatusIn != nil {
			for _, status := range *req.Filter.StatusIn {
				statuses = append(statuses, string(status))
			}
		}
	}

	totalItems := int32(0)
	page := int64(0)
	if req.Pagination != nil {
		totalItems = req.Pagination.TotalItems
		page = req.Pagination.Page
	}

	searchRes, err := h.authClient.SearchClient(ctx.Request().Context(), auth.SearchClientParam{
		Keyword:    typeconv.StringVal(req.Keyword),
		Statuses:   statuses,
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

	items := []restapp.SearchAuthClientItem{}
	for _, searchItem := range searchRes.Items {
		var updatedAt *int64
		if searchItem.UpdatedAt != nil {
			updated := searchItem.UpdatedAt.UnixMilli()
			updatedAt = &updated
		}

		items = append(items, restapp.SearchAuthClientItem{
			Id:        searchItem.Id,
			ClientId:  searchItem.ClientId,
			Name:      searchItem.Name,
			Type:      searchItem.Type,
			Status:    searchItem.Status,
			CreatedAt: searchItem.CreatedAt.UnixMilli(),
			UpdatedAt: updatedAt,
		})
	}

	return ctx.JSON(http.StatusOK, &restapp.SearchAuthClientResponse{
		Code:    searchRes.Success.Code,
		Message: searchRes.Success.Message,
		Data: restapp.SearchAuthClientData{
			Items: items,
			Summary: restapp.SearchAuthClientSummary{
				Page:       searchRes.Summary.Page,
				TotalItems: searchRes.Summary.TotalItems,
			},
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
