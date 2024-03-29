package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-seidon/chariot/internal/repository"
	"github.com/go-seidon/provider/random"
	"github.com/go-seidon/provider/typeconv"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

type file struct {
	gormClient *gorm.DB
	randomizer random.Randomizer
}

func (r *file) CreateFile(ctx context.Context, p repository.CreateFileParam) (*repository.CreateFileResult, error) {
	tx := r.gormClient.
		WithContext(ctx).
		Clauses(dbresolver.Write).
		Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	slug := p.Slug
	checkRes := tx.
		Select(`id, slug`).
		First(&File{}, "slug = ?", slug)
	failedCheckSlug := checkRes.Error != nil &&
		!errors.Is(checkRes.Error, gorm.ErrRecordNotFound)
	if failedCheckSlug {
		txRes := tx.Rollback()
		if txRes.Error != nil {
			return nil, txRes.Error
		}
		return nil, checkRes.Error
	}

	if checkRes.Error == nil {
		token, err := r.randomizer.String(7)
		if err != nil {
			txRes := tx.Rollback()
			if txRes.Error != nil {
				return nil, txRes.Error
			}
			return nil, err
		}

		slug = strings.TrimSuffix(slug, "."+p.Extension)
		slug = fmt.Sprintf("%s-%s", slug, token)
		if p.Extension != "" {
			slug = fmt.Sprintf("%s.%s", slug, p.Extension)
		}
	}

	metas := []FileMeta{}
	if len(p.Meta) > 0 {
		for key, value := range p.Meta {
			metas = append(metas, FileMeta{
				FileId: p.Id,
				Key:    key,
				Value:  value,
			})
		}
	}

	locations := []FileLocation{}
	if len(p.Locations) > 0 {
		for _, location := range p.Locations {
			var externalId sql.NullString
			if location.ExternalId != nil {
				externalId = sql.NullString{
					String: *location.ExternalId,
					Valid:  true,
				}
			}

			var uploadedAt sql.NullInt64
			if location.UploadedAt != nil {
				uploadedAt = sql.NullInt64{
					Int64: location.UploadedAt.UnixMilli(),
					Valid: true,
				}
			}

			locations = append(locations, FileLocation{
				Id:         location.Id,
				FileId:     p.Id,
				BarrelId:   location.BarrelId,
				Status:     location.Status,
				ExternalId: externalId,
				Priority:   location.Priority,
				CreatedAt:  location.CreatedAt.UnixMilli(),
				UpdatedAt:  location.CreatedAt.UnixMilli(),
				UploadedAt: uploadedAt,
			})
		}
	}

	createFile := tx.Create(File{
		Id:         p.Id,
		Slug:       slug,
		Name:       p.Name,
		Mimetype:   p.Mimetype,
		Extension:  p.Extension,
		Size:       p.Size,
		Visibility: p.Visibility,
		Status:     p.Status,
		CreatedAt:  p.CreatedAt.UnixMilli(),
		UpdatedAt:  p.CreatedAt.UnixMilli(),
		UploadedAt: p.UploadedAt.UnixMilli(),
		Metas:      metas,
		Locations:  locations,
	})
	if createFile.Error != nil {
		txRes := tx.Rollback()
		if txRes.Error != nil {
			return nil, txRes.Error
		}
		return nil, createFile.Error
	}

	file := &File{}
	uploadFile := tx.
		Select(`id, slug, name, mimetype, extension, size, visibility, status, created_at, uploaded_at`).
		Preload("Metas", func(tx *gorm.DB) *gorm.DB {
			return tx.Select("file_id, `key`, value")
		}).
		Preload("Locations", func(tx *gorm.DB) *gorm.DB {
			return tx.Select("file_id, barrel_id, external_id, priority, status, created_at, uploaded_at")
		}).
		First(file, "id = ?", p.Id)
	if uploadFile.Error != nil {
		txRes := tx.Rollback()
		if txRes.Error != nil {
			return nil, txRes.Error
		}
		return nil, uploadFile.Error
	}

	txRes := tx.Commit()
	if txRes.Error != nil {
		return nil, txRes.Error
	}

	meta := map[string]string{}
	for _, item := range file.Metas {
		meta[item.Key] = item.Value
	}

	location := []repository.CreateFileLocation{}
	for _, item := range file.Locations {
		var externalId *string
		if item.ExternalId.Valid {
			externalId = typeconv.String(item.ExternalId.String)
		}

		var uploadedAt *time.Time
		if item.UploadedAt.Valid {
			uploadedAt = typeconv.Time(time.UnixMilli(item.UploadedAt.Int64).UTC())
		}

		location = append(location, repository.CreateFileLocation{
			BarrelId:   item.BarrelId,
			ExternalId: externalId,
			Priority:   item.Priority,
			Status:     item.Status,
			CreatedAt:  time.UnixMilli(item.CreatedAt).UTC(),
			UploadedAt: uploadedAt,
		})
	}

	res := &repository.CreateFileResult{
		Id:         file.Id,
		Slug:       file.Slug,
		Name:       file.Name,
		Mimetype:   file.Mimetype,
		Extension:  file.Extension,
		Size:       file.Size,
		Visibility: file.Visibility,
		Status:     file.Status,
		CreatedAt:  time.UnixMilli(file.CreatedAt).UTC(),
		UploadedAt: time.UnixMilli(file.UploadedAt).UTC(),
		Meta:       meta,
		Locations:  location,
	}
	return res, nil
}

func (r *file) FindFile(ctx context.Context, p repository.FindFileParam) (*repository.FindFileResult, error) {
	file := &File{}

	query := r.gormClient.
		WithContext(ctx).
		Clauses(dbresolver.Read)

	findRes := query.
		Select(`
			file.id, file.slug, file.name, file.mimetype, 
			file.extension, file.size, file.visibility, file.status, 
			file.created_at, file.updated_at, file.deleted_at, file.uploaded_at
		`).
		Preload("Metas", func(tx *gorm.DB) *gorm.DB {
			return tx.Select("file_id, `key`, value")
		}).
		Preload("Locations", func(tx *gorm.DB) *gorm.DB {
			return tx.Select("id, file_id, barrel_id, external_id, priority, status, created_at, updated_at, uploaded_at")
		}).
		Preload("Locations.Barrel", func(tx *gorm.DB) *gorm.DB {
			return tx.Select("id, code, provider, status")
		})

	if p.Slug != "" {
		findRes = findRes.First(file, "file.slug = ?", p.Slug)
	} else if p.Id != "" {
		findRes = findRes.First(file, "file.id = ?", p.Id)
	} else {
		findRes = findRes.
			Joins("LEFT JOIN file_location AS fl ON fl.file_id = file.id").
			Where("fl.id = ?", p.LocationId).
			First(file)
	}

	if findRes.Error != nil {
		if errors.Is(findRes.Error, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, findRes.Error
	}

	meta := map[string]string{}
	for _, item := range file.Metas {
		meta[item.Key] = item.Value
	}

	location := []repository.FindFileLocation{}
	for _, item := range file.Locations {
		var externalId *string
		if item.ExternalId.Valid {
			externalId = typeconv.String(item.ExternalId.String)
		}

		var uploadedAt *time.Time
		if item.UploadedAt.Valid {
			uploadedAt = typeconv.Time(time.UnixMilli(item.UploadedAt.Int64).UTC())
		}

		location = append(location, repository.FindFileLocation{
			Barrel: repository.FindFileBarrel{
				Id:       item.Barrel.Id,
				Code:     item.Barrel.Code,
				Provider: item.Barrel.Provider,
				Status:   item.Barrel.Status,
			},
			Id:         item.Id,
			ExternalId: externalId,
			Priority:   item.Priority,
			Status:     item.Status,
			CreatedAt:  time.UnixMilli(item.CreatedAt).UTC(),
			UploadedAt: uploadedAt,
			UpdatedAt:  typeconv.Time(time.UnixMilli(item.UpdatedAt).UTC()),
		})
	}

	var deletedAt *time.Time
	if file.DeletedAt.Valid {
		deletedAt = typeconv.Time(time.UnixMilli(file.DeletedAt.Int64).UTC())
	}

	res := &repository.FindFileResult{
		Id:         file.Id,
		Slug:       file.Slug,
		Name:       file.Name,
		Mimetype:   file.Mimetype,
		Extension:  file.Extension,
		Size:       file.Size,
		Visibility: file.Visibility,
		Status:     file.Status,
		CreatedAt:  time.UnixMilli(file.CreatedAt).UTC(),
		UpdatedAt:  typeconv.Time(time.UnixMilli(file.UpdatedAt).UTC()),
		UploadedAt: time.UnixMilli(file.UploadedAt).UTC(),
		DeletedAt:  deletedAt,
		Meta:       meta,
		Locations:  location,
	}
	return res, nil
}

func (r *file) SearchFile(ctx context.Context, p repository.SearchFileParam) (*repository.SearchFileResult, error) {
	query := r.gormClient.
		WithContext(ctx).
		Clauses(dbresolver.Read)

	if p.Keyword != "" {
		query.Where("name LIKE ?", "%"+p.Keyword+"%")
	}

	if len(p.StatusIn) > 0 {
		query.Where("status IN ?", p.StatusIn)
	}

	if len(p.VisibilityIn) > 0 {
		query.Where("visibility IN ?", p.VisibilityIn)
	}

	if len(p.ExtensionIn) > 0 {
		query.Where("extension IN ?", p.ExtensionIn)
	}

	if p.SizeGte > 0 {
		query.Where("size >= ?", p.SizeGte)
	}

	if p.SizeLte > 0 {
		query.Where("size <= ?", p.SizeLte)
	}

	if p.UploadDateGte > 0 {
		query.Where("uploaded_at >= ?", p.UploadDateGte)
	}

	if p.UploadDateLte > 0 {
		query.Where("uploaded_at <= ?", p.UploadDateLte)
	}

	res := &repository.SearchFileResult{
		Summary: repository.SearchFileSummary{},
		Items:   []repository.SearchFileItem{},
	}
	countRes := query.
		Table("file").
		Count(&res.Summary.TotalItems)
	if countRes.Error != nil {
		return nil, countRes.Error
	}

	switch p.Sort {
	case "latest_upload":
		query.Order("uploaded_at DESC")
	case "newest_upload":
		query.Order("uploaded_at ASC")
	case "highest_size":
		query.Order("size DESC")
	case "lowest_size":
		query.Order("size ASC")
	}

	if p.Limit > 0 {
		query.Limit(int(p.Limit))
	}
	if p.Offset > 0 {
		query.Offset(int(p.Offset))
	}

	files := []File{}
	searchRes := query.
		Select(`id, slug, name, mimetype, extension, size, visibility, status, uploaded_at, created_at, updated_at, deleted_at`).
		Preload("Metas", func(tx *gorm.DB) *gorm.DB {
			return tx.Select("file_id, `key`, value")
		}).
		Find(&files)
	if searchRes.Error != nil {
		if errors.Is(searchRes.Error, gorm.ErrRecordNotFound) {
			return res, nil
		}
		return nil, searchRes.Error
	}

	for _, file := range files {
		meta := map[string]string{}
		for _, item := range file.Metas {
			meta[item.Key] = item.Value
		}

		res.Items = append(res.Items, repository.SearchFileItem{
			Id:         file.Id,
			Slug:       file.Slug,
			Name:       file.Name,
			Mimetype:   file.Mimetype,
			Extension:  file.Extension,
			Size:       file.Size,
			Visibility: file.Visibility,
			Status:     file.Status,
			UploadedAt: time.UnixMilli(file.UploadedAt).UTC(),
			CreatedAt:  time.UnixMilli(file.CreatedAt).UTC(),
			UpdatedAt:  typeconv.Time(time.UnixMilli(file.UpdatedAt).UTC()),
			DeletedAt:  typeconv.SqlUnixMilli(file.DeletedAt),
			Meta:       meta,
		})
	}

	return res, nil
}

func (r *file) UpdateFile(ctx context.Context, p repository.UpdateFileParam) (*repository.UpdateFileResult, error) {
	tx := r.gormClient.
		WithContext(ctx).
		Clauses(dbresolver.Write).
		Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	data := map[string]interface{}{
		"updated_at": p.UpdatedAt.UnixMilli(),
	}
	if p.Status != nil {
		data["status"] = p.Status
	}
	if p.DeletedAt != nil {
		data["deleted_at"] = p.DeletedAt.UnixMilli()
	}

	updateRes := tx.
		Model(&File{}).
		Where("id = ?", p.Id).
		Updates(data)
	if updateRes.Error != nil {
		txRes := tx.Rollback()
		if txRes.Error != nil {
			return nil, txRes.Error
		}
		return nil, updateRes.Error
	}

	file := &File{}
	checkRes := tx.
		Select(`id, slug, name, mimetype, extension, size, visibility, status, created_at, updated_at, uploaded_at, deleted_at`).
		First(file, "id = ?", p.Id)
	if checkRes.Error != nil {
		txRes := tx.Rollback()
		if txRes.Error != nil {
			return nil, txRes.Error
		}
		return nil, checkRes.Error
	}

	txRes := tx.Commit()
	if txRes.Error != nil {
		return nil, txRes.Error
	}

	var deletedAt *time.Time
	if file.DeletedAt.Valid {
		deletedAt = typeconv.Time(time.UnixMilli(file.DeletedAt.Int64).UTC())
	}

	res := &repository.UpdateFileResult{
		Id:         file.Id,
		Slug:       file.Slug,
		Name:       file.Name,
		Mimetype:   file.Mimetype,
		Extension:  file.Extension,
		Size:       file.Size,
		Visibility: file.Visibility,
		Status:     file.Status,
		CreatedAt:  time.UnixMilli(file.CreatedAt).UTC(),
		UpdatedAt:  time.UnixMilli(file.UpdatedAt).UTC(),
		UploadedAt: time.UnixMilli(file.UploadedAt).UTC(),
		DeletedAt:  deletedAt,
	}
	return res, nil
}

func (r *file) SearchLocation(ctx context.Context, p repository.SearchLocationParam) (*repository.SearchLocationResult, error) {
	query := r.gormClient.
		WithContext(ctx).
		Clauses(dbresolver.Read)

	searchQuery := query.Select(`id, file_id, barrel_id, priority, status`)

	if len(p.Statuses) > 0 {
		searchQuery = searchQuery.
			Where("status IN ?", p.Statuses)
	}

	countQuery := searchQuery.Table("file_location")

	if p.Limit > 0 {
		searchQuery = searchQuery.Limit(int(p.Limit))
	}

	locations := []FileLocation{}
	searchRes := searchQuery.
		Order("created_at ASC").
		Find(&locations)

	res := &repository.SearchLocationResult{
		Summary: repository.SearchLocationSummary{
			TotalItems: 0,
		},
		Items: []repository.SearchLocationItem{},
	}
	if searchRes.Error != nil {
		if errors.Is(searchRes.Error, gorm.ErrRecordNotFound) {
			return res, nil
		}
		return nil, searchRes.Error
	}

	for _, location := range locations {
		res.Items = append(res.Items, repository.SearchLocationItem{
			Id:       location.Id,
			FileId:   location.FileId,
			BarrelId: location.BarrelId,
			Priority: location.Priority,
			Status:   location.Status,
		})
	}

	countRes := countQuery.Count(&res.Summary.TotalItems)
	if countRes.Error != nil {
		return nil, countRes.Error
	}
	return res, nil
}

func (r *file) UpdateLocationByIds(ctx context.Context, p repository.UpdateLocationByIdsParam) (*repository.UpdateLocationByIdsResult, error) {
	tx := r.gormClient.
		WithContext(ctx).
		Clauses(dbresolver.Write).
		Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	data := map[string]interface{}{
		"updated_at": p.UpdatedAt.UnixMilli(),
	}
	if p.Status != nil {
		data["status"] = p.Status
	}
	if p.ExternalId != nil {
		data["external_id"] = p.ExternalId
	}
	if p.UploadedAt != nil {
		data["uploaded_at"] = p.UploadedAt.UnixMilli()
	}
	if p.DeletedAt != nil {
		data["deleted_at"] = p.DeletedAt.UnixMilli()
	}

	updateRes := tx.
		Model(&FileLocation{}).
		Where("id IN ?", p.Ids).
		Updates(data)
	if updateRes.Error != nil {
		txRes := tx.Rollback()
		if txRes.Error != nil {
			return nil, txRes.Error
		}
		return nil, updateRes.Error
	}

	txRes := tx.Commit()
	if txRes.Error != nil {
		return nil, txRes.Error
	}

	res := &repository.UpdateLocationByIdsResult{
		TotalUpdated: updateRes.RowsAffected,
	}
	return res, nil
}

type FileParam struct {
	GormClient *gorm.DB
	Randomizer random.Randomizer
}

func NewFile(p FileParam) *file {
	return &file{
		gormClient: p.GormClient,
		randomizer: p.Randomizer,
	}
}

type File struct {
	Id         string         `gorm:"column:id;primaryKey"`
	Slug       string         `gorm:"column:slug"`
	Name       string         `gorm:"column:name"`
	Mimetype   string         `gorm:"column:mimetype"`
	Extension  string         `gorm:"column:extension"`
	Size       int64          `gorm:"column:size"`
	Visibility string         `gorm:"column:visibility"`
	Status     string         `gorm:"column:status"`
	UploadedAt int64          `gorm:"column:uploaded_at"`
	CreatedAt  int64          `gorm:"column:created_at"`
	UpdatedAt  int64          `gorm:"column:updated_at;autoUpdateTime:milli"`
	DeletedAt  sql.NullInt64  `gorm:"column:deleted_at;<-:update"`
	Metas      []FileMeta     `gorm:"foreignKey:FileId;references:Id"`
	Locations  []FileLocation `gorm:"foreignKey:FileId;references:Id"`
}

func (File) TableName() string {
	return "file"
}

type FileLocation struct {
	Id         string         `gorm:"column:id;primaryKey"`
	FileId     string         `gorm:"column:file_id"`
	BarrelId   string         `gorm:"column:barrel_id"`
	ExternalId sql.NullString `gorm:"column:external_id"`
	Priority   int32          `gorm:"column:priority"`
	Status     string         `gorm:"column:status"`
	UploadedAt sql.NullInt64  `gorm:"column:uploaded_at"`
	CreatedAt  int64          `gorm:"column:created_at"`
	UpdatedAt  int64          `gorm:"column:updated_at;autoUpdateTime:milli"`
	Barrel     Barrel         `gorm:"foreignKey:BarrelId;references:Id"`
}

func (FileLocation) TableName() string {
	return "file_location"
}

type FileMeta struct {
	FileId string `gorm:"column:file_id"`
	Key    string `gorm:"column:key"`
	Value  string `gorm:"column:value"`
}

func (FileMeta) TableName() string {
	return "file_meta"
}
