package file

import (
	"context"
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
}

type UploadFileInfo struct {
	Name      string            `validate:"max=1024" label:"name"`
	Mimetype  string            `validate:"max=128" label:"mimetype"`
	Extension string            `validate:"max=32" label:"extension"`
	Size      int64             `validate:"min=1" label:"size"`
	Meta      map[string]string `validate:"min=0,max=24,dive,keys,printascii,min=1,max=64,endkeys,required,printascii,min=1,max=128" label:"meta"`
}

type UploadFileSetting struct {
	Visibility string   `validate:"required,oneof='public' 'protected'" label:"visibility"`
	Barrels    []string `validate:"required,unique,min=1,max=3,dive,min=1,max=128" label:"barrels"`
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
		FileId     string
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
			FileId     string
			BarrelId   string
			BarrelCode string
		}{
			FileId:     id,
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
		Id:        typeconv.String(primaryBarrel.FileId),
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
		Id:         primaryBarrel.FileId,
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
