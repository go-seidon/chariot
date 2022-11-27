package file

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/go-seidon/chariot/internal/barrel"
	"github.com/go-seidon/chariot/internal/repository"
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
	STATUS_UPLOADING = "uploading"
	STATUS_AVAILABLE = "available"
	STATUS_DELETING  = "deleting"
	STATUS_DELETED   = "deleted"
)

type File interface {
	UploadFile(ctx context.Context, p UploadFileParam) (*UploadFileResult, *system.SystemError)
	RetrieveFileBySlug(ctx context.Context, p RetrieveFileBySlugParam) (*RetrieveFileBySlugResult, *system.SystemError)
}

type UploadFileInfo struct {
	Name      string            `validate:"max=256" label:"name"`
	Mimetype  string            `validate:"max=128" label:"mimetype"`
	Extension string            `validate:"max=32" label:"extension"`
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
	Meta       map[string]string
	UploadedAt time.Time
}

type RetrieveFileBySlugParam struct {
	Slug string `validate:"required,min=1,max=288" label:"slug"`
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

type file struct {
	validator  validation.Validator
	identifier identifier.Identifier
	slugger    slug.Slugger
	clock      datetime.Clock
	router     router.Router
	barrelRepo repository.Barrel
	fileRepo   repository.File
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
		status := STATUS_UPLOADING
		var externalId *string
		var uploadedAt *time.Time
		if i == 0 {
			status = STATUS_AVAILABLE
			externalId = typeconv.String(uploadFile.ObjectId)
			uploadedAt = typeconv.Time(uploadFile.UploadedAt)
		}

		locations = append(locations, repository.CreateFileLocation{
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
		Meta:       createFile.Meta,
		UploadedAt: createFile.UploadedAt,
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

type FileParam struct {
	Validator  validation.Validator
	Identifier identifier.Identifier
	Slugger    slug.Slugger
	Clock      datetime.Clock
	Router     router.Router
	BarrelRepo repository.Barrel
	FileRepo   repository.File
}

func NewFile(p FileParam) *file {
	return &file{
		validator:  p.Validator,
		identifier: p.Identifier,
		slugger:    p.Slugger,
		clock:      p.Clock,
		router:     p.Router,
		barrelRepo: p.BarrelRepo,
		fileRepo:   p.FileRepo,
	}
}
