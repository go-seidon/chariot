package resthandler

import (
	"net/http"
	"strings"

	"github.com/go-seidon/chariot/generated/restapp"
	"github.com/go-seidon/chariot/internal/file"
	"github.com/go-seidon/chariot/internal/storage/multipart"
	"github.com/go-seidon/provider/serialization"
	"github.com/go-seidon/provider/status"
	"github.com/go-seidon/provider/typeconv"
	"github.com/labstack/echo/v4"
)

type fileHandler struct {
	fileClient file.File
	fileParser multipart.Parser
	serializer serialization.Serializer
}

func (h *fileHandler) UploadFile(ctx echo.Context) error {
	fileHeader, ferr := ctx.FormFile("file")
	if ferr != nil {
		return echo.NewHTTPError(http.StatusBadRequest, &restapp.ResponseBodyInfo{
			Code:    status.INVALID_PARAM,
			Message: ferr.Error(),
		})
	}

	fileInfo, ferr := h.fileParser(fileHeader)
	if ferr != nil {
		return echo.NewHTTPError(http.StatusBadRequest, &restapp.ResponseBodyInfo{
			Code:    status.INVALID_PARAM,
			Message: ferr.Error(),
		})
	}

	meta := map[string]string{}
	metas := strings.TrimSpace(ctx.FormValue("meta"))
	if metas != "" {
		err := h.serializer.Unmarshal([]byte(metas), &meta)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, &restapp.ResponseBodyInfo{
				Code:    status.INVALID_PARAM,
				Message: err.Error(),
			})
		}
	}

	uploadFile, err := h.fileClient.UploadFile(ctx.Request().Context(), file.UploadFileParam{
		Data: fileInfo.Data,
		Info: file.UploadFileInfo{
			Name:      fileInfo.Name,
			Size:      fileInfo.Size,
			Mimetype:  fileInfo.Mimetype,
			Extension: fileInfo.Extension,
			Meta:      meta,
		},
		Setting: file.UploadFileSetting{
			Visibility: ctx.FormValue("visibility"),
			Barrels:    strings.Split(ctx.FormValue("barrels"), ","),
		},
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

	return ctx.JSON(http.StatusCreated, &restapp.UploadFileResponse{
		Code:    uploadFile.Success.Code,
		Message: uploadFile.Success.Message,
		Data: restapp.UploadFileData{
			Id:         uploadFile.Id,
			Slug:       uploadFile.Slug,
			Name:       uploadFile.Name,
			Extension:  uploadFile.Extension,
			Size:       uploadFile.Size,
			Mimetype:   uploadFile.Mimetype,
			Visibility: restapp.UploadFileDataVisibility(uploadFile.Visibility),
			Status:     restapp.UploadFileDataStatus(uploadFile.Status),
			UploadedAt: uploadFile.UploadedAt.UnixMilli(),
			Meta: &restapp.UploadFileData_Meta{
				AdditionalProperties: uploadFile.Meta,
			},
		},
	})
}

func (h *fileHandler) RetrieveFileBySlug(ctx echo.Context) error {
	findFile, err := h.fileClient.RetrieveFileBySlug(ctx.Request().Context(), file.RetrieveFileBySlugParam{
		Slug: ctx.Param("slug"),
	})
	if err != nil {
		httpCode := http.StatusInternalServerError
		switch err.Code {
		case status.INVALID_PARAM:
			httpCode = http.StatusBadRequest
		case status.RESOURCE_NOTFOUND:
			httpCode = http.StatusNotFound
		}
		return echo.NewHTTPError(httpCode, &restapp.ResponseBodyInfo{
			Code:    err.Code,
			Message: err.Message,
		})
	}
	return ctx.Stream(http.StatusOK, findFile.Mimetype, findFile.Data)
}

func (h *fileHandler) GetFileById(ctx echo.Context) error {
	getFile, err := h.fileClient.GetFileById(ctx.Request().Context(), file.GetFileByIdParam{
		Id: ctx.Param("id"),
	})
	if err != nil {
		httpCode := http.StatusInternalServerError
		switch err.Code {
		case status.INVALID_PARAM:
			httpCode = http.StatusBadRequest
		case status.RESOURCE_NOTFOUND:
			httpCode = http.StatusNotFound
		}
		return echo.NewHTTPError(httpCode, &restapp.ResponseBodyInfo{
			Code:    err.Code,
			Message: err.Message,
		})
	}

	var updatedAt *int64
	if getFile.UpdatedAt != nil {
		updatedAt = typeconv.Int64(getFile.UpdatedAt.UnixMilli())
	}

	var deletedAt *int64
	if getFile.DeletedAt != nil {
		deletedAt = typeconv.Int64(getFile.DeletedAt.UnixMilli())
	}

	locations := []restapp.GetFileByIdLocation{}
	for _, location := range getFile.Locations {
		var updatedAt *int64
		if location.UpdatedAt != nil {
			updatedAt = typeconv.Int64(location.UpdatedAt.UnixMilli())
		}

		var uploadedAt *int64
		if location.UploadedAt != nil {
			uploadedAt = typeconv.Int64(location.UploadedAt.UnixMilli())
		}

		locations = append(locations, restapp.GetFileByIdLocation{
			ExternalId:     location.ExternalId,
			Priority:       location.Priority,
			Status:         location.Status,
			CreatedAt:      location.CreatedAt.UnixMilli(),
			UpdatedAt:      updatedAt,
			UploadedAt:     uploadedAt,
			BarrelId:       location.Barrel.Id,
			BarrelCode:     location.Barrel.Code,
			BarrelProvider: location.Barrel.Provider,
			BarrelStatus:   location.Barrel.Status,
		})
	}

	return ctx.JSON(http.StatusOK, &restapp.GetFileByIdResponse{
		Code:    getFile.Success.Code,
		Message: getFile.Success.Message,
		Data: restapp.GetFileByIdData{
			Id:         getFile.Id,
			Slug:       getFile.Slug,
			Name:       getFile.Name,
			Mimetype:   getFile.Mimetype,
			Extension:  getFile.Extension,
			Size:       getFile.Size,
			Visibility: restapp.GetFileByIdDataVisibility(getFile.Visibility),
			Status:     restapp.GetFileByIdDataStatus(getFile.Status),
			UploadedAt: getFile.UploadedAt.Local().UnixMilli(),
			CreatedAt:  getFile.CreatedAt.UnixMilli(),
			UpdatedAt:  updatedAt,
			DeletedAt:  deletedAt,
			Meta:       &restapp.GetFileByIdData_Meta{AdditionalProperties: getFile.Meta},
			Locations:  locations,
		},
	})
}

type FileParam struct {
	File       file.File
	FileParser multipart.Parser
	Serializer serialization.Serializer
}

func NewFile(p FileParam) *fileHandler {
	return &fileHandler{
		fileClient: p.File,
		fileParser: p.FileParser,
		serializer: p.Serializer,
	}
}
