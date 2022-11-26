package mysql

import (
	"context"
	"errors"
	"time"

	"github.com/go-seidon/chariot/internal/repository"
	"github.com/go-seidon/provider/typeconv"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

type barrel struct {
	gormClient *gorm.DB
}

// @note: return `ErrExists` if code is already created
func (r *barrel) CreateBarrel(ctx context.Context, p repository.CreateBarrelParam) (*repository.CreateBarrelResult, error) {
	tx := r.gormClient.
		WithContext(ctx).
		Clauses(dbresolver.Write).
		Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	currentBarrel := &Barrel{}
	checkRes := tx.
		Select("id, code").
		First(currentBarrel, "code = ?", p.Code)
	if !errors.Is(checkRes.Error, gorm.ErrRecordNotFound) {
		txRes := tx.Rollback()
		if txRes.Error != nil {
			return nil, txRes.Error
		}
		if checkRes.Error == nil {
			return nil, repository.ErrExists
		}
		return nil, checkRes.Error
	}

	createParam := &Barrel{
		Id:        p.Id,
		Code:      p.Code,
		Name:      p.Name,
		Provider:  p.Provider,
		Status:    p.Status,
		CreatedAt: p.CreatedAt.UnixMilli(),
		UpdatedAt: p.CreatedAt.UnixMilli(),
	}
	createRes := tx.Create(createParam)
	if createRes.Error != nil {
		txRes := tx.Rollback()
		if txRes.Error != nil {
			return nil, txRes.Error
		}
		return nil, createRes.Error
	}

	barrel := &Barrel{}
	findRes := tx.
		Select("id, code, name, provider, status, created_at").
		First(barrel, "id = ?", p.Id)
	if findRes.Error != nil {
		txRes := tx.Rollback()
		if txRes.Error != nil {
			return nil, txRes.Error
		}
		return nil, findRes.Error
	}

	txRes := tx.Commit()
	if txRes.Error != nil {
		return nil, txRes.Error
	}

	res := &repository.CreateBarrelResult{
		Id:        barrel.Id,
		Code:      barrel.Code,
		Name:      barrel.Name,
		Provider:  barrel.Provider,
		Status:    barrel.Status,
		CreatedAt: time.UnixMilli(barrel.CreatedAt).UTC(),
	}
	return res, nil
}

func (r *barrel) FindBarrel(ctx context.Context, p repository.FindBarrelParam) (*repository.FindBarrelResult, error) {
	barrel := &Barrel{}

	query := r.gormClient.
		WithContext(ctx).
		Clauses(dbresolver.Read)

	findRes := query.
		Select(`id, code, name, provider, status, created_at, updated_at`).
		First(barrel, "id = ?", p.Id)
	if findRes.Error != nil {
		if errors.Is(findRes.Error, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, findRes.Error
	}

	res := &repository.FindBarrelResult{
		Id:        barrel.Id,
		Code:      barrel.Code,
		Name:      barrel.Name,
		Provider:  barrel.Provider,
		Status:    barrel.Status,
		CreatedAt: time.UnixMilli(barrel.CreatedAt).UTC(),
		UpdatedAt: typeconv.Time(time.UnixMilli(barrel.UpdatedAt).UTC()),
	}
	return res, nil
}

func (r *barrel) UpdateBarrel(ctx context.Context, p repository.UpdateBarrelParam) (*repository.UpdateBarrelResult, error) {
	tx := r.gormClient.
		WithContext(ctx).
		Clauses(dbresolver.Write).
		Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	findRes := tx.
		Select(`id, code, name, provider, status`).
		First(&Barrel{}, "id = ?", p.Id)
	if findRes.Error != nil {
		txRes := tx.Rollback()
		if txRes.Error != nil {
			return nil, txRes.Error
		}
		if errors.Is(findRes.Error, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, findRes.Error
	}

	updateRes := tx.
		Model(&Barrel{}).
		Where("id = ?", p.Id).
		Updates(map[string]interface{}{
			"code":       p.Code,
			"name":       p.Name,
			"provider":   p.Provider,
			"status":     p.Status,
			"updated_at": p.UpdatedAt.UnixMilli(),
		})
	if updateRes.Error != nil {
		txRes := tx.Rollback()
		if txRes.Error != nil {
			return nil, txRes.Error
		}
		return nil, updateRes.Error
	}

	barrel := &Barrel{}
	checkRes := tx.
		Select(`id, code, name, provider, status, created_at, updated_at`).
		First(barrel, "id = ?", p.Id)
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

	res := &repository.UpdateBarrelResult{
		Id:        barrel.Id,
		Code:      barrel.Code,
		Name:      barrel.Name,
		Provider:  barrel.Provider,
		Status:    barrel.Status,
		CreatedAt: time.UnixMilli(barrel.CreatedAt).UTC(),
		UpdatedAt: time.UnixMilli(barrel.UpdatedAt).UTC(),
	}
	return res, nil
}

func (r *barrel) SearchBarrel(ctx context.Context, p repository.SearchBarrelParam) (*repository.SearchBarrelResult, error) {
	query := r.gormClient.
		WithContext(ctx).
		Clauses(dbresolver.Read)

	searchQuery := query
	if p.Keyword != "" {
		searchQuery = searchQuery.
			Where("name LIKE ?", "%"+p.Keyword+"%").
			Or("code LIKE ?", "%"+p.Keyword+"%")
	}

	if len(p.Statuses) > 0 {
		searchQuery = searchQuery.
			Where("status IN ?", p.Statuses)
	}

	if len(p.Providers) > 0 {
		searchQuery = searchQuery.
			Where("provider IN ?", p.Providers)
	}

	if len(p.Codes) > 0 {
		searchQuery = searchQuery.
			Where("code IN ?", p.Codes)
	}

	countQuery := searchQuery.Table("barrel")

	if p.Limit > 0 {
		searchQuery = searchQuery.Limit(int(p.Limit))
	}

	if p.Offset > 0 {
		searchQuery = searchQuery.Offset(int(p.Offset))
	}

	res := &repository.SearchBarrelResult{
		Summary: repository.SearchBarrelSummary{},
		Items:   []repository.SearchBarrelItem{},
	}
	barrels := []Barrel{}

	searchRes := searchQuery.
		Select(`id, code, name, provider, status, created_at, updated_at`).
		Find(&barrels)

	if searchRes.Error != nil {
		if errors.Is(searchRes.Error, gorm.ErrRecordNotFound) {
			return res, nil
		}
		return nil, searchRes.Error
	}

	for _, barrel := range barrels {
		res.Items = append(res.Items, repository.SearchBarrelItem{
			Id:        barrel.Id,
			Code:      barrel.Code,
			Name:      barrel.Name,
			Provider:  barrel.Provider,
			Status:    barrel.Status,
			CreatedAt: time.UnixMilli(barrel.CreatedAt).UTC(),
			UpdatedAt: typeconv.Time(time.UnixMilli(barrel.UpdatedAt).UTC()),
		})
	}

	countRes := countQuery.Count(&res.Summary.TotalItems)
	if countRes.Error != nil {
		return nil, countRes.Error
	}
	return res, nil
}

type BarrelParam struct {
	GormClient *gorm.DB
}

func NewBarrel(p BarrelParam) *barrel {
	return &barrel{
		gormClient: p.GormClient,
	}
}

type Barrel struct {
	Id        string `gorm:"column:id;primaryKey"`
	Code      string `gorm:"column:code"`
	Name      string `gorm:"column:name"`
	Provider  string `gorm:"column:provider"`
	Status    string `gorm:"column:status"`
	CreatedAt int64  `gorm:"column:created_at"`
	UpdatedAt int64  `gorm:"column:updated_at;autoUpdateTime:milli"`
}

func (Barrel) TableName() string {
	return "barrel"
}
