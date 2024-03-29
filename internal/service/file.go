package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/go-seidon/chariot/api/queue"
	"github.com/go-seidon/chariot/internal/barrel"
	"github.com/go-seidon/chariot/internal/file"
	"github.com/go-seidon/chariot/internal/repository"
	"github.com/go-seidon/chariot/internal/storage"
	"github.com/go-seidon/chariot/internal/storage/router"
	"github.com/go-seidon/provider/datetime"
	"github.com/go-seidon/provider/identity"
	"github.com/go-seidon/provider/queueing"
	"github.com/go-seidon/provider/serialization"
	"github.com/go-seidon/provider/slug"
	"github.com/go-seidon/provider/status"
	"github.com/go-seidon/provider/system"
	"github.com/go-seidon/provider/typeconv"
	"github.com/go-seidon/provider/validation"
)

type File interface {
	UploadFile(ctx context.Context, p UploadFileParam) (*UploadFileResult, *system.Error)
	RetrieveFileBySlug(ctx context.Context, p RetrieveFileBySlugParam) (*RetrieveFileBySlugResult, *system.Error)
	GetFileById(ctx context.Context, p GetFileByIdParam) (*GetFileByIdResult, *system.Error)
	SearchFile(ctx context.Context, p SearchFileParam) (*SearchFileResult, *system.Error)
	DeleteFileById(ctx context.Context, p DeleteFileByIdParam) (*DeleteFileByIdResult, *system.Error)
	ProceedDeletion(ctx context.Context, p ProceedDeletionParam) (*ProceedDeletionResult, *system.Error)
	ScheduleReplication(ctx context.Context, p ScheduleReplicationParam) (*ScheduleReplicationResult, *system.Error)
	ProceedReplication(ctx context.Context, p ProceedReplicationParam) (*ProceedReplicationResult, *system.Error)
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
	Success    system.Success
	Id         string
	Slug       string
	Name       string
	Mimetype   string
	Extension  string
	Size       int64
	Visibility string
	Status     string
	FileUrl    string
	AccessUrl  string
	UploadedAt time.Time
	Meta       map[string]string
}

type RetrieveFileBySlugParam struct {
	Slug  string `validate:"required,min=1,max=288" label:"slug"`
	Token string
}

type RetrieveFileBySlugResult struct {
	Data       io.ReadCloser
	Success    system.Success
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
	Success    system.Success
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
	Success system.Success
	Items   []SearchFileItem
	Summary SearchFileSummary
}

type DeleteFileByIdParam struct {
	Id string `validate:"required,min=5,max=64" label:"id"`
}

type DeleteFileByIdResult struct {
	Success     system.Success
	RequestedAt time.Time
}

type ProceedDeletionParam struct {
	LocationId string `validate:"required,min=5,max=64" label:"location_id"`
}

type ProceedDeletionResult struct {
	Success   system.Success
	DeletedAt time.Time
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

type ScheduleReplicationParam struct {
	MaxItems int32 `validate:"numeric,min=1,max=50" label:"max_items"`
}

type ScheduleReplicationResult struct {
	Success     system.Success
	TotalItems  int32
	ScheduledAt *time.Time
}

type ProceedReplicationParam struct {
	LocationId string `validate:"required,min=5,max=64" label:"location_id"`
}

type ProceedReplicationResult struct {
	Success    system.Success
	LocationId *string
	BarrelId   *string
	ExternalId *string
	UploadedAt *time.Time
}

var _ File = (*fileService)(nil)

type fileService struct {
	config        *FileConfig
	validator     validation.Validator
	identifier    identity.Identifier
	sessionClient Session
	slugger       slug.Slugger
	clock         datetime.Clock
	serializer    serialization.Serializer
	pubsub        queueing.Pubsub
	router        router.Router
	barrelRepo    repository.Barrel
	fileRepo      repository.File
}

func (f *fileService) UploadFile(ctx context.Context, p UploadFileParam) (*UploadFileResult, *system.Error) {
	if p.Data == nil {
		return nil, &system.Error{
			Code:    status.INVALID_PARAM,
			Message: "file is not specified",
		}
	}

	err := f.validator.Validate(p)
	if err != nil {
		return nil, &system.Error{
			Code:    status.INVALID_PARAM,
			Message: err.Error(),
		}
	}

	var token string
	if p.Setting.Visibility == file.VISIBILITY_PROTECTED {
		session, err := f.sessionClient.CreateSession(ctx, CreateSessionParam{
			Duration: 1800,
			Features: []string{"retrieve_file"},
		})
		if err != nil {
			return nil, &system.Error{
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
		return nil, &system.Error{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	if len(searchBarrels.Items) != len(p.Setting.Barrels) {
		return nil, &system.Error{
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
			return nil, &system.Error{
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
		return nil, &system.Error{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	uploadFile, err := uploader.UploadObject(ctx, storage.UploadObjectParam{
		Data:      p.Data,
		Id:        typeconv.String(primaryBarrel.ObjectId),
		Name:      typeconv.String(strings.ToLower(p.Info.Name)),
		Mimetype:  typeconv.String(strings.ToLower(p.Info.Mimetype)),
		Extension: typeconv.String(strings.ToLower(p.Info.Extension)),
	})
	if err != nil {
		return nil, &system.Error{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	currentTs := f.clock.Now()
	locations := []repository.CreateFileLocation{}
	for i, barrel := range barrels {
		status := file.STATUS_PENDING
		var externalId *string
		var uploadedAt *time.Time
		if i == 0 {
			status = file.STATUS_AVAILABLE
			externalId = typeconv.String(uploadFile.ObjectId)
			uploadedAt = typeconv.Time(uploadFile.UploadedAt)
		}

		locations = append(locations, repository.CreateFileLocation{
			Id:         barrel.ObjectId,
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
		Name:       strings.ToLower(p.Info.Name),
		Mimetype:   strings.ToLower(p.Info.Mimetype),
		Extension:  strings.ToLower(p.Info.Extension),
		Size:       p.Info.Size,
		Meta:       p.Info.Meta,
		Status:     file.STATUS_AVAILABLE,
		Visibility: p.Setting.Visibility,
		CreatedAt:  currentTs,
		UploadedAt: uploadFile.UploadedAt,
		Locations:  locations,
	})
	if err != nil {
		return nil, &system.Error{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	fileUrl := fmt.Sprintf("%s/file/%s", f.config.AppHost, createFile.Slug)
	accessUrl := fileUrl
	if createFile.Visibility == file.VISIBILITY_PROTECTED {
		accessUrl = fmt.Sprintf("%s?token=%s", fileUrl, token)
	}

	res := &UploadFileResult{
		Success: system.Success{
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
		AccessUrl:  accessUrl,
		Meta:       createFile.Meta,
	}
	return res, nil
}

func (f *fileService) RetrieveFileBySlug(ctx context.Context, p RetrieveFileBySlugParam) (*RetrieveFileBySlugResult, *system.Error) {
	err := f.validator.Validate(p)
	if err != nil {
		return nil, &system.Error{
			Code:    status.INVALID_PARAM,
			Message: err.Error(),
		}
	}

	findFile, err := f.fileRepo.FindFile(ctx, repository.FindFileParam{
		Slug: p.Slug,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, &system.Error{
				Code:    status.RESOURCE_NOTFOUND,
				Message: "file is not available",
			}
		}
		return nil, &system.Error{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	if findFile.Status != file.STATUS_AVAILABLE {
		return nil, &system.Error{
			Code:    status.RESOURCE_NOTFOUND,
			Message: "file is not available",
		}
	}

	if findFile.Visibility == file.VISIBILITY_PROTECTED {
		_, err := f.sessionClient.VerifySession(ctx, VerifySessionParam{
			Token:   p.Token,
			Feature: "retrieve_file",
		})
		if err != nil {
			return nil, &system.Error{
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

		if location.Status != file.STATUS_AVAILABLE {
			if i+1 == len(findFile.Locations) {
				ferr = fmt.Errorf("file replicas are not available")
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
		return nil, &system.Error{
			Code:    status.ACTION_FAILED,
			Message: ferr.Error(),
		}
	}

	if data == nil {
		return nil, &system.Error{
			Code:    status.ACTION_FAILED,
			Message: "file data is invalid",
		}
	}

	res := &RetrieveFileBySlugResult{
		Success: system.Success{
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

func (f *fileService) GetFileById(ctx context.Context, p GetFileByIdParam) (*GetFileByIdResult, *system.Error) {
	err := f.validator.Validate(p)
	if err != nil {
		return nil, &system.Error{
			Code:    status.INVALID_PARAM,
			Message: err.Error(),
		}
	}

	findFile, err := f.fileRepo.FindFile(ctx, repository.FindFileParam{
		Id: p.Id,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, &system.Error{
				Code:    status.RESOURCE_NOTFOUND,
				Message: "file is not available",
			}
		}
		return nil, &system.Error{
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
		Success: system.Success{
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

func (f *fileService) SearchFile(ctx context.Context, p SearchFileParam) (*SearchFileResult, *system.Error) {
	err := f.validator.Validate(p)
	if err != nil {
		return nil, &system.Error{
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
		return nil, &system.Error{
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
		Success: system.Success{
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

// @note for refactoring: use go-routine
func (f *fileService) DeleteFileById(ctx context.Context, p DeleteFileByIdParam) (*DeleteFileByIdResult, *system.Error) {
	err := f.validator.Validate(p)
	if err != nil {
		return nil, &system.Error{
			Code:    status.INVALID_PARAM,
			Message: err.Error(),
		}
	}

	findFile, err := f.fileRepo.FindFile(ctx, repository.FindFileParam{
		Id: p.Id,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, &system.Error{
				Code:    status.RESOURCE_NOTFOUND,
				Message: "file is not available",
			}
		}
		return nil, &system.Error{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	if findFile.Status != file.STATUS_AVAILABLE {
		return nil, &system.Error{
			Code:    status.RESOURCE_NOTFOUND,
			Message: "file is not available",
		}
	}

	currentTs := f.clock.Now().UTC()
	updated, err := f.fileRepo.UpdateFile(ctx, repository.UpdateFileParam{
		Id:        findFile.Id,
		UpdatedAt: currentTs,
		DeletedAt: typeconv.Time(currentTs),
		Status:    typeconv.String(file.STATUS_DELETED),
	})
	if err != nil {
		return nil, &system.Error{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	msgs := [][]byte{}
	for _, location := range findFile.Locations {
		msg, err := f.serializer.Marshal(&queue.DeleteFileMessage{
			LocationId:  location.Id,
			BarrelId:    location.Barrel.Id,
			FileId:      updated.Id,
			Status:      updated.Status,
			RequestedAt: updated.UpdatedAt.UnixMilli(),
		})
		if err != nil {
			return nil, &system.Error{
				Code:    status.ACTION_FAILED,
				Message: err.Error(),
			}
		}
		msgs = append(msgs, msg)
	}

	for _, msg := range msgs {
		err = f.pubsub.Publish(ctx, queueing.PublishParam{
			ExchangeName: "file_deletion",
			MessageBody:  msg,
		})
		if err != nil {
			return nil, &system.Error{
				Code:    status.ACTION_FAILED,
				Message: err.Error(),
			}
		}
	}

	res := &DeleteFileByIdResult{
		Success: system.Success{
			Code:    status.ACTION_SUCCESS,
			Message: "success delete file",
		},
		RequestedAt: updated.UpdatedAt,
	}
	return res, nil
}

func (f *fileService) ProceedDeletion(ctx context.Context, p ProceedDeletionParam) (*ProceedDeletionResult, *system.Error) {
	err := f.validator.Validate(p)
	if err != nil {
		return nil, &system.Error{
			Code:    status.INVALID_PARAM,
			Message: err.Error(),
		}
	}

	findFile, err := f.fileRepo.FindFile(ctx, repository.FindFileParam{
		LocationId: p.LocationId,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, &system.Error{
				Code:    status.RESOURCE_NOTFOUND,
				Message: "file is not available",
			}
		}
		return nil, &system.Error{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	var deleteLocation repository.FindFileLocation
	for _, lc := range findFile.Locations {
		if lc.Id == p.LocationId {
			deleteLocation = lc
		}
	}

	if deleteLocation.Status == file.STATUS_DELETING {
		return nil, &system.Error{
			Code:    status.ACTION_FORBIDDEN,
			Message: "deletion is already proceeded",
		}
	}

	deleteStorage, err := f.router.CreateStorage(ctx, router.CreateStorageParam{
		BarrelCode: deleteLocation.Barrel.Code,
	})
	if err != nil {
		return nil, &system.Error{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	currentTs := f.clock.Now().UTC()
	_, err = f.fileRepo.UpdateLocationByIds(ctx, repository.UpdateLocationByIdsParam{
		Ids:       []string{deleteLocation.Id},
		Status:    typeconv.String(file.STATUS_DELETING),
		UpdatedAt: currentTs,
	})
	if err != nil {
		return nil, &system.Error{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	if deleteLocation.Status == file.STATUS_AVAILABLE {
		_, err := deleteStorage.DeleteObject(ctx, storage.DeleteObjectParam{
			ObjectId: typeconv.StringVal(deleteLocation.ExternalId),
		})
		if err != nil {
			return nil, &system.Error{
				Code:    status.ACTION_FAILED,
				Message: err.Error(),
			}
		}
	}

	currentTs = f.clock.Now().UTC()
	_, err = f.fileRepo.UpdateLocationByIds(ctx, repository.UpdateLocationByIdsParam{
		Ids:       []string{deleteLocation.Id},
		UpdatedAt: currentTs,
		Status:    typeconv.String(file.STATUS_DELETED),
		DeletedAt: typeconv.Time(currentTs),
	})
	if err != nil {
		return nil, &system.Error{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	res := &ProceedDeletionResult{
		Success: system.Success{
			Code:    status.ACTION_SUCCESS,
			Message: "success delete file",
		},
		DeletedAt: currentTs,
	}
	return res, nil
}

// @note for refactoring: use go-routine
func (f *fileService) ScheduleReplication(ctx context.Context, p ScheduleReplicationParam) (*ScheduleReplicationResult, *system.Error) {
	err := f.validator.Validate(p)
	if err != nil {
		return nil, &system.Error{
			Code:    status.INVALID_PARAM,
			Message: err.Error(),
		}
	}

	searchres, err := f.fileRepo.SearchLocation(ctx, repository.SearchLocationParam{
		Limit:    p.MaxItems,
		Statuses: []string{file.STATUS_PENDING},
	})
	if err != nil {
		return nil, &system.Error{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	if len(searchres.Items) == 0 {
		res := &ScheduleReplicationResult{
			Success: system.Success{
				Code:    status.ACTION_SUCCESS,
				Message: "there is no pending replication",
			},
			TotalItems: 0,
		}
		return res, nil
	}

	ids := []string{}
	msgs := [][]byte{}
	for _, location := range searchres.Items {
		ids = append(ids, location.Id)
		msg, err := f.serializer.Marshal(&queue.ScheduleReplicationMessage{
			LocationId: location.Id,
			FileId:     location.FileId,
			BarrelId:   location.BarrelId,
			Priority:   location.Priority,
			Status:     location.Status,
		})
		if err != nil {
			return nil, &system.Error{
				Code:    status.ACTION_FAILED,
				Message: err.Error(),
			}
		}
		msgs = append(msgs, msg)
	}

	currentTs := f.clock.Now().UTC()
	_, err = f.fileRepo.UpdateLocationByIds(ctx, repository.UpdateLocationByIdsParam{
		Ids:       ids,
		Status:    typeconv.String(file.STATUS_REPLICATING),
		UpdatedAt: currentTs,
	})
	if err != nil {
		return nil, &system.Error{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	for _, msg := range msgs {
		err = f.pubsub.Publish(ctx, queueing.PublishParam{
			ExchangeName: "file_replication",
			MessageBody:  msg,
		})
		if err != nil {
			return nil, &system.Error{
				Code:    status.ACTION_FAILED,
				Message: err.Error(),
			}
		}
	}

	res := &ScheduleReplicationResult{
		Success: system.Success{
			Code:    status.ACTION_SUCCESS,
			Message: "success schedule replication",
		},
		TotalItems:  int32(len(searchres.Items)),
		ScheduledAt: typeconv.Time(currentTs),
	}
	return res, nil
}

func (f *fileService) ProceedReplication(ctx context.Context, p ProceedReplicationParam) (*ProceedReplicationResult, *system.Error) {
	err := f.validator.Validate(p)
	if err != nil {
		return nil, &system.Error{
			Code:    status.INVALID_PARAM,
			Message: err.Error(),
		}
	}

	findFile, err := f.fileRepo.FindFile(ctx, repository.FindFileParam{
		LocationId: p.LocationId,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, &system.Error{
				Code:    status.RESOURCE_NOTFOUND,
				Message: "file is not available",
			}
		}
		return nil, &system.Error{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	var primaryLocation repository.FindFileLocation
	var replicaLocation repository.FindFileLocation
	for _, lc := range findFile.Locations {
		if lc.Priority == 1 {
			primaryLocation = lc
		}

		if lc.Id == p.LocationId {
			replicaLocation = lc
		}
	}

	if replicaLocation.Status != file.STATUS_REPLICATING {
		return nil, &system.Error{
			Code:    status.ACTION_FORBIDDEN,
			Message: "replication is already proceeded",
		}
	}

	currentTs := f.clock.Now().UTC()
	_, err = f.fileRepo.UpdateLocationByIds(ctx, repository.UpdateLocationByIdsParam{
		Ids:       []string{replicaLocation.Id},
		Status:    typeconv.String(file.STATUS_UPLOADING),
		UpdatedAt: currentTs,
	})
	if err != nil {
		return nil, &system.Error{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	primaryStorage, err := f.router.CreateStorage(ctx, router.CreateStorageParam{
		BarrelCode: primaryLocation.Barrel.Code,
	})
	if err != nil {
		return nil, &system.Error{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	replicaStorage, err := f.router.CreateStorage(ctx, router.CreateStorageParam{
		BarrelCode: replicaLocation.Barrel.Code,
	})
	if err != nil {
		return nil, &system.Error{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	retrieval, err := primaryStorage.RetrieveObject(ctx, storage.RetrieveObjectParam{
		ObjectId: typeconv.StringVal(primaryLocation.ExternalId),
	})
	if err != nil {
		return nil, &system.Error{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	uploaded, err := replicaStorage.UploadObject(ctx, storage.UploadObjectParam{
		Data:      retrieval.Data,
		Name:      typeconv.String(findFile.Name),
		Mimetype:  typeconv.String(findFile.Mimetype),
		Extension: typeconv.String(findFile.Extension),
	})
	if err != nil {
		return nil, &system.Error{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	currentTs = f.clock.Now().UTC()
	_, err = f.fileRepo.UpdateLocationByIds(ctx, repository.UpdateLocationByIdsParam{
		Ids:        []string{replicaLocation.Id},
		Status:     typeconv.String(file.STATUS_AVAILABLE),
		ExternalId: typeconv.String(uploaded.ObjectId),
		UploadedAt: typeconv.Time(uploaded.UploadedAt),
		UpdatedAt:  currentTs,
	})
	if err != nil {
		return nil, &system.Error{
			Code:    status.ACTION_FAILED,
			Message: err.Error(),
		}
	}

	res := &ProceedReplicationResult{
		Success: system.Success{
			Code:    status.ACTION_SUCCESS,
			Message: "success replicate file",
		},
		ExternalId: typeconv.String(uploaded.ObjectId),
		LocationId: typeconv.String(replicaLocation.Id),
		BarrelId:   typeconv.String(replicaLocation.Barrel.Id),
		UploadedAt: typeconv.Time(uploaded.UploadedAt),
	}
	return res, nil
}

type FileParam struct {
	Config        *FileConfig
	Validator     validation.Validator
	Identifier    identity.Identifier
	SessionClient Session
	Slugger       slug.Slugger
	Clock         datetime.Clock
	Serializer    serialization.Serializer
	Pubsub        queueing.Pubsub
	Router        router.Router
	BarrelRepo    repository.Barrel
	FileRepo      repository.File
}

func NewFile(p FileParam) *fileService {
	return &fileService{
		config:        p.Config,
		validator:     p.Validator,
		identifier:    p.Identifier,
		sessionClient: p.SessionClient,
		slugger:       p.Slugger,
		clock:         p.Clock,
		serializer:    p.Serializer,
		pubsub:        p.Pubsub,
		router:        p.Router,
		barrelRepo:    p.BarrelRepo,
		fileRepo:      p.FileRepo,
	}
}
