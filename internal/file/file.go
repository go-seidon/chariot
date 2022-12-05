package file

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/go-seidon/chariot/internal/barrel"
	"github.com/go-seidon/chariot/internal/repository"
	"github.com/go-seidon/chariot/internal/session"
	"github.com/go-seidon/chariot/internal/storage"
	"github.com/go-seidon/chariot/internal/storage/router"
	"github.com/go-seidon/provider/datetime"
	"github.com/go-seidon/provider/identifier"
	"github.com/go-seidon/provider/slug"
	"github.com/go-seidon/provider/status"
	"github.com/go-seidon/provider/system"
	"github.com/go-seidon/provider/typeconv"
	"github.com/go-seidon/provider/validation"
)

const (
	STATUS_PENDING   = "pending"
	STATUS_UPLOADING = "uploading"
	STATUS_AVAILABLE = "available"
	STATUS_DELETING  = "deleting"
	STATUS_DELETED   = "deleted"

	VISIBILITY_PUBLIC    = "public"
	VISIBILITY_PROTECTED = "protected"
)

type File interface {
	UploadFile(ctx context.Context, p UploadFileParam) (*UploadFileResult, *system.SystemError)
	RetrieveFileBySlug(ctx context.Context, p RetrieveFileBySlugParam) (*RetrieveFileBySlugResult, *system.SystemError)
	GetFileById(ctx context.Context, p GetFileByIdParam) (*GetFileByIdResult, *system.SystemError)
	SearchFile(ctx context.Context, p SearchFileParam) (*SearchFileResult, *system.SystemError)
}

type FileConfig struct {
	AppHost string
}

type UploadFileInfo struct {
	Name      string            `validate:"max=256" label:"name"`
	Mimetype  string            `validate:"max=128" label:"mimetype"`
	Extension string            `validate:"max=32,printascii" label:"extension"`
	Size      int64             `validate:"min=1" label:"size"`
	Meta      map[string]string `validate:"min=0,max=24,dive,keys,printascii,min=1,max=64,endkeys,required,printascii,min=1,max=128" label:"meta"`
}

type UploadFileSetting struct {
	Visibility string   `validate:"required,oneof='public' 'protected'" label:"visibility"`
	Barrels    []string `validate:"required,unique,min=1,max=3,dive,required,lowercase,alphanum,min=6,max=128" label:"barrels"`
}

type UploadFileParam struct {
	Data    io.Reader
	Info    UploadFileInfo
	Setting UploadFileSetting
}

type UploadFileResult struct {
	Success    system.SystemSuccess
	Id         string
	Slug       string
	Name       string
	Mimetype   string
	Extension  string
	Size       int64
	Visibility string
	Status     string
	FileUrl    string
	UploadedAt time.Time
	Meta       map[string]string
}

type RetrieveFileBySlugParam struct {
	Slug  string `validate:"required,min=1,max=288" label:"slug"`
	Token string
}

type RetrieveFileBySlugResult struct {
	Data       io.ReadCloser
	Success    system.SystemSuccess
	Id         string
	Slug       string
	Name       string
	Mimetype   string
	Extension  string
	Size       int64
	Visibility string
	Status     string
	Meta       map[string]string
	UploadedAt time.Time
}

type GetFileByIdParam struct {
	Id string `validate:"required,min=5,max=64" label:"id"`
}

type GetFileByIdResult struct {
	Success    system.SystemSuccess
	Id         string
	Slug       string
	Name       string
	Mimetype   string
	Extension  string
	Size       int64
	Visibility string
	Status     string
	UploadedAt time.Time
	CreatedAt  time.Time
	UpdatedAt  *time.Time
	DeletedAt  *time.Time
	Meta       map[string]string
	Locations  []GetFileByIdLocation
}

type GetFileByIdBarrel struct {
	Id       string
	Code     string
	Provider string
	Status   string
}

type GetFileByIdLocation struct {
	Barrel     GetFileByIdBarrel
	ExternalId *string
	Priority   int32
	Status     string
	CreatedAt  time.Time
	UpdatedAt  *time.Time
	UploadedAt *time.Time
}

type SearchFileParam struct {
	Keyword       string   `validate:"omitempty,min=2,max=64" label:"keyword"`
	TotalItems    int32    `validate:"numeric,min=1,max=100" label:"total_items"`
	Page          int64    `validate:"numeric,min=1" label:"page"`
	Sort          string   `validate:"omitempty,oneof='latest_upload' 'newest_upload' 'highest_size' 'lowest_size'" label:"sort"`
	StatusIn      []string `validate:"unique,min=0,max=4,dive,oneof='uploading' 'available' 'deleting' 'deleted'" label:"status_in"`
	VisibilityIn  []string `validate:"unique,min=0,max=2,dive,oneof='public' 'protected'" label:"visibility_in"`
	ExtensionIn   []string `validate:"unique,min=0,max=16,dive" label:"extension_in"`
	SizeGte       int64    `validate:"numeric,min=0" label:"size_gte"`
	SizeLte       int64    `validate:"numeric,min=0" label:"size_lte"`
	UploadDateGte int64    `validate:"numeric,min=0" label:"upload_date_gte"`
	UploadDateLte int64    `validate:"numeric,min=0" label:"upload_date_lte"`
}

type SearchFileResult struct {
	Success system.SystemSuccess
	Items   []SearchFileItem
	Summary SearchFileSummary
}

type SearchFileItem struct {
	Id         string
	Slug       string
	Name       string
	Mimetype   string
	Extension  string
	Size       int64
	Visibility string
	Status     string
	UploadedAt time.Time
	CreatedAt  time.Time
	UpdatedAt  *time.Time
	DeletedAt  *time.Time
	Meta       map[string]string
}

type SearchFileSummary struct {
	TotalItems int64
	Page       int64
}

type file struct {
	config        *FileConfig
	validator     validation.Validator
	identifier    identifier.Identifier
	sessionClient session.Session
	slugger       slug.Slugger
	clock         datetime.Clock
	router        router.Router
	barrelRepo    repository.Barrel
	fileRepo      repository.File
}

func (f *file) UploadFile(ctx context.Context, p UploadFileParam) (*UploadFileResult, *system.SystemError) {
	if p.Data == nil {
		return nil, &system.SystemError{
			Code:    status.INVALID_PARAM,
			Message: "file is not specified",
		}
	}

	err := f.validator.Validate(p)
	if err != nil {
		return nil, &system.SystemError{
			Code:    status.INVALID_PARAM,
			Message: err.Error(),
		}
	}

	var token string
	if p.Setting.Visibility == VISIBILITY_PROTECTED {
		session, err := f.sessionClient.CreateSession(ctx, session.CreateSessionParam{
			Duration: 30 * time.Minute,
			Features: []string{"retrieve_file"},
		})
		if err != nil {
			return nil, &system.SystemError{
				Code:    status.ACTION_FAILED,
				Message: err.Error(),
			}
		}
		token = session.Token
	}

	searchBarrels, err := f.barrelRepo.SearchBarrel(ctx, repository.SearchBarrelParam{
		Codes:    p.Setting.Barrels,
		Statuses: []string{barrel.STATUS_ACTIVE},
	})
	if err != nil {
		return nil, &system.SystemError{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	if len(searchBarrels.Items) != len(p.Setting.Barrels) {
		return nil, &system.SystemError{
			Code:    status.INVALID_PARAM,
			Message: "there is invalid barrel",
		}
	}

	searchItems := searchBarrels.SortCodes(p.Setting.Barrels)
	barrels := []struct {
		ObjectId   string
		BarrelId   string
		BarrelCode string
	}{}
	for _, searchItem := range searchItems {
		id, err := f.identifier.GenerateId()
		if err != nil {
			return nil, &system.SystemError{
				Code:    status.ACTION_FAILED,
				Message: err.Error(),
			}
		}

		barrels = append(barrels, struct {
			ObjectId   string
			BarrelId   string
			BarrelCode string
		}{
			ObjectId:   id,
			BarrelId:   searchItem.Id,
			BarrelCode: searchItem.Code,
		})
	}

	primaryBarrel := barrels[0]
	uploader, err := f.router.CreateStorage(ctx, router.CreateStorageParam{
		BarrelCode: primaryBarrel.BarrelCode,
	})
	if err != nil {
		return nil, &system.SystemError{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	uploadFile, err := uploader.UploadObject(ctx, storage.UploadObjectParam{
		Data:      p.Data,
		Id:        typeconv.String(primaryBarrel.ObjectId),
		Name:      typeconv.String(p.Info.Name),
		Mimetype:  typeconv.String(p.Info.Mimetype),
		Extension: typeconv.String(p.Info.Extension),
	})
	if err != nil {
		return nil, &system.SystemError{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	currentTs := f.clock.Now()
	locations := []repository.CreateFileLocation{}
	for i, barrel := range barrels {
		id, err := f.identifier.GenerateId()
		if err != nil {
			return nil, &system.SystemError{
				Code:    status.ACTION_FAILED,
				Message: err.Error(),
			}
		}

		status := STATUS_PENDING
		var externalId *string
		var uploadedAt *time.Time
		if i == 0 {
			status = STATUS_AVAILABLE
			externalId = typeconv.String(uploadFile.ObjectId)
			uploadedAt = typeconv.Time(uploadFile.UploadedAt)
		}

		locations = append(locations, repository.CreateFileLocation{
			Id:         id,
			BarrelId:   barrel.BarrelId,
			Priority:   int32(i) + 1,
			CreatedAt:  currentTs,
			ExternalId: externalId,
			Status:     status,
			UploadedAt: uploadedAt,
		})
	}

	slug := f.slugger.GenerateSlug(p.Info.Name)
	if p.Info.Extension != "" {
		slug = fmt.Sprintf("%s.%s", slug, p.Info.Extension)
	}

	createFile, err := f.fileRepo.CreateFile(ctx, repository.CreateFileParam{
		Id:         primaryBarrel.ObjectId,
		Slug:       slug,
		Name:       p.Info.Name,
		Mimetype:   p.Info.Mimetype,
		Extension:  p.Info.Extension,
		Size:       p.Info.Size,
		Meta:       p.Info.Meta,
		Status:     STATUS_AVAILABLE,
		Visibility: p.Setting.Visibility,
		CreatedAt:  currentTs,
		UploadedAt: uploadFile.UploadedAt,
		Locations:  locations,
	})
	if err != nil {
		return nil, &system.SystemError{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	fileUrl := fmt.Sprintf("%s/file/%s", f.config.AppHost, createFile.Slug)
	if createFile.Visibility == VISIBILITY_PROTECTED {
		fileUrl = fmt.Sprintf("%s?token=%s", fileUrl, token)
	}

	res := &UploadFileResult{
		Success: system.SystemSuccess{
			Code:    status.ACTION_SUCCESS,
			Message: "success upload file",
		},
		Id:         createFile.Id,
		Slug:       createFile.Slug,
		Name:       createFile.Name,
		Mimetype:   createFile.Mimetype,
		Extension:  createFile.Extension,
		Size:       createFile.Size,
		Visibility: createFile.Visibility,
		Status:     createFile.Status,
		UploadedAt: createFile.UploadedAt,
		FileUrl:    fileUrl,
		Meta:       createFile.Meta,
	}
	return res, nil
}

func (f *file) RetrieveFileBySlug(ctx context.Context, p RetrieveFileBySlugParam) (*RetrieveFileBySlugResult, *system.SystemError) {
	err := f.validator.Validate(p)
	if err != nil {
		return nil, &system.SystemError{
			Code:    status.INVALID_PARAM,
			Message: err.Error(),
		}
	}

	findFile, err := f.fileRepo.FindFile(ctx, repository.FindFileParam{
		Slug: p.Slug,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, &system.SystemError{
				Code:    status.RESOURCE_NOTFOUND,
				Message: "file is not available",
			}
		}
		return nil, &system.SystemError{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	if findFile.Visibility == VISIBILITY_PROTECTED {
		_, err := f.sessionClient.VerifySession(ctx, session.VerifySessionParam{
			Token:   p.Token,
			Feature: "retrieve_file",
		})
		if err != nil {
			return nil, &system.SystemError{
				Code:    status.ACTION_FORBIDDEN,
				Message: err.Error(),
			}
		}
	}

	var ferr error
	var data io.ReadCloser
	for i, location := range findFile.Locations {
		if location.Barrel.Status != barrel.STATUS_ACTIVE {
			if i+1 == len(findFile.Locations) {
				ferr = fmt.Errorf("barrels are not active")
			}
			continue
		}

		st, err := f.router.CreateStorage(ctx, router.CreateStorageParam{
			BarrelCode: location.Barrel.Code,
		})
		if err != nil {
			ferr = err
			break
		}

		retrieveObj, err := st.RetrieveObject(ctx, storage.RetrieveObjectParam{
			ObjectId: typeconv.StringVal(location.ExternalId),
		})
		if err == nil {
			data = retrieveObj.Data
			break
		}

		if i+1 == len(findFile.Locations) {
			ferr = fmt.Errorf("failed retrieve file from barrel")
		}
	}

	if ferr != nil {
		return nil, &system.SystemError{
			Code:    status.ACTION_FAILED,
			Message: ferr.Error(),
		}
	}

	if data == nil {
		return nil, &system.SystemError{
			Code:    status.ACTION_FAILED,
			Message: "file data is invalid",
		}
	}

	res := &RetrieveFileBySlugResult{
		Success: system.SystemSuccess{
			Code:    status.ACTION_SUCCESS,
			Message: "success retrieve file",
		},
		Data:       data,
		Id:         findFile.Id,
		Slug:       findFile.Slug,
		Name:       findFile.Name,
		Mimetype:   findFile.Mimetype,
		Extension:  findFile.Extension,
		Size:       findFile.Size,
		Visibility: findFile.Visibility,
		Status:     findFile.Status,
		Meta:       findFile.Meta,
		UploadedAt: findFile.UploadedAt,
	}
	return res, nil
}

func (f *file) GetFileById(ctx context.Context, p GetFileByIdParam) (*GetFileByIdResult, *system.SystemError) {
	err := f.validator.Validate(p)
	if err != nil {
		return nil, &system.SystemError{
			Code:    status.INVALID_PARAM,
			Message: err.Error(),
		}
	}

	findFile, err := f.fileRepo.FindFile(ctx, repository.FindFileParam{
		Id: p.Id,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, &system.SystemError{
				Code:    status.RESOURCE_NOTFOUND,
				Message: "file is not available",
			}
		}
		return nil, &system.SystemError{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	locations := []GetFileByIdLocation{}
	for _, location := range findFile.Locations {
		locations = append(locations, GetFileByIdLocation{
			Barrel: GetFileByIdBarrel{
				Id:       location.Barrel.Id,
				Code:     location.Barrel.Code,
				Provider: location.Barrel.Provider,
				Status:   location.Barrel.Status,
			},
			ExternalId: location.ExternalId,
			Priority:   location.Priority,
			Status:     location.Status,
			CreatedAt:  location.CreatedAt,
			UpdatedAt:  location.UpdatedAt,
			UploadedAt: location.UploadedAt,
		})
	}

	res := &GetFileByIdResult{
		Success: system.SystemSuccess{
			Code:    status.ACTION_SUCCESS,
			Message: "success get file",
		},
		Id:         findFile.Id,
		Slug:       findFile.Slug,
		Name:       findFile.Name,
		Mimetype:   findFile.Mimetype,
		Extension:  findFile.Extension,
		Size:       findFile.Size,
		Visibility: findFile.Visibility,
		Status:     findFile.Status,
		UploadedAt: findFile.UploadedAt,
		CreatedAt:  findFile.CreatedAt,
		UpdatedAt:  findFile.UpdatedAt,
		DeletedAt:  findFile.DeletedAt,
		Meta:       findFile.Meta,
		Locations:  locations,
	}
	return res, nil
}

func (f *file) SearchFile(ctx context.Context, p SearchFileParam) (*SearchFileResult, *system.SystemError) {
	err := f.validator.Validate(p)
	if err != nil {
		return nil, &system.SystemError{
			Code:    status.INVALID_PARAM,
			Message: err.Error(),
		}
	}

	offset := int64(0)
	if p.Page > 1 {
		offset = (p.Page - 1) * int64(p.TotalItems)
	}

	searchRes, err := f.fileRepo.SearchFile(ctx, repository.SearchFileParam{
		Sort:          p.Sort,
		Limit:         p.TotalItems,
		Offset:        offset,
		Keyword:       p.Keyword,
		StatusIn:      p.StatusIn,
		VisibilityIn:  p.VisibilityIn,
		ExtensionIn:   p.ExtensionIn,
		SizeGte:       p.SizeGte,
		SizeLte:       p.SizeLte,
		UploadDateGte: p.UploadDateGte,
		UploadDateLte: p.UploadDateLte,
	})
	if err != nil {
		return nil, &system.SystemError{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	items := []SearchFileItem{}
	for _, file := range searchRes.Items {
		items = append(items, SearchFileItem{
			Id:         file.Id,
			Name:       file.Name,
			Slug:       file.Slug,
			Mimetype:   file.Mimetype,
			Extension:  file.Extension,
			Size:       file.Size,
			Visibility: file.Visibility,
			Status:     file.Status,
			UploadedAt: file.UploadedAt,
			CreatedAt:  file.CreatedAt,
			UpdatedAt:  file.UpdatedAt,
			DeletedAt:  file.DeletedAt,
			Meta:       file.Meta,
		})
	}

	res := &SearchFileResult{
		Success: system.SystemSuccess{
			Code:    status.ACTION_SUCCESS,
			Message: "success search file",
		},
		Items: items,
		Summary: SearchFileSummary{
			TotalItems: searchRes.Summary.TotalItems,
			Page:       p.Page,
		},
	}
	return res, nil
}

type FileParam struct {
	Config        *FileConfig
	Validator     validation.Validator
	Identifier    identifier.Identifier
	SessionClient session.Session
	Slugger       slug.Slugger
	Clock         datetime.Clock
	Router        router.Router
	BarrelRepo    repository.Barrel
	FileRepo      repository.File
}

func NewFile(p FileParam) *file {
	return &file{
		config:        p.Config,
		validator:     p.Validator,
		identifier:    p.Identifier,
		sessionClient: p.SessionClient,
		slugger:       p.Slugger,
		clock:         p.Clock,
		router:        p.Router,
		barrelRepo:    p.BarrelRepo,
		fileRepo:      p.FileRepo,
	}
}
