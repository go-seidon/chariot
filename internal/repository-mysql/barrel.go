package repository_mysql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-seidon/chariot/internal/repository"
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

type BarrelParam struct {
	GormClient *gorm.DB
}

func NewBarrel(p BarrelParam) *barrel {
	return &barrel{
		gormClient: p.GormClient,
	}
}

type Barrel struct {
	Id        string        `gorm:"column:id;primaryKey"`
	Code      string        `gorm:"column:code"`
	Name      string        `gorm:"column:name"`
	Provider  string        `gorm:"column:provider"`
	Status    string        `gorm:"column:status"`
	CreatedAt int64         `gorm:"column:created_at"`
	UpdatedAt sql.NullInt64 `gorm:"column:updated_at;autoUpdateTime:milli;<-:update"`
}

func (Barrel) TableName() string {
	return "barrel"
}
