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

type auth struct {
	gormClient *gorm.DB
}

// @note: return `ErrExists` if client_id is already created
func (r *auth) CreateClient(ctx context.Context, p repository.CreateClientParam) (*repository.CreateClientResult, error) {
	tx := r.gormClient.
		WithContext(ctx).
		Clauses(dbresolver.Write).
		Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	currentClient := &AauthClient{}
	checkRes := tx.
		Select("id, client_id").
		First(currentClient, "client_id = ?", p.ClientId)
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

	createParam := &AauthClient{
		Id:           p.Id,
		ClientId:     p.ClientId,
		ClientSecret: p.ClientSecret,
		Name:         p.Name,
		Type:         p.Type,
		Status:       p.Status,
		CreatedAt:    p.CreatedAt.UnixMilli(),
	}
	createRes := tx.Create(createParam)
	if createRes.Error != nil {
		txRes := tx.Rollback()
		if txRes.Error != nil {
			return nil, txRes.Error
		}
		return nil, createRes.Error
	}

	authClient := &AauthClient{}
	findRes := tx.
		Select("id, client_id, client_secret, name, type, status, created_at").
		First(authClient, "id = ?", p.Id)
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

	res := &repository.CreateClientResult{
		Id:           authClient.Id,
		ClientId:     authClient.ClientId,
		ClientSecret: authClient.ClientSecret,
		Name:         authClient.Name,
		Type:         authClient.Type,
		Status:       authClient.Status,
		CreatedAt:    time.UnixMilli(authClient.CreatedAt).UTC(),
	}
	return res, nil
}

func (r *auth) FindClient(ctx context.Context, p repository.FindClientParam) (*repository.FindClientResult, error) {
	authClient := &AauthClient{}

	query := r.gormClient.
		WithContext(ctx).
		Clauses(dbresolver.Read)

	findRes := query.
		Select(`id, client_id, client_secret, name, type, status, created_at, updated_at`).
		First(authClient, "id = ?", p.Id)
	if findRes.Error != nil {
		if errors.Is(findRes.Error, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, findRes.Error
	}

	var updatedAt *time.Time
	if authClient.UpdatedAt.Valid {
		updatedDate := time.UnixMilli(authClient.UpdatedAt.Int64).UTC()
		updatedAt = &updatedDate
	}

	res := &repository.FindClientResult{
		Id:           authClient.Id,
		ClientId:     authClient.ClientId,
		ClientSecret: authClient.ClientSecret,
		Name:         authClient.Name,
		Type:         authClient.Type,
		Status:       authClient.Status,
		CreatedAt:    time.UnixMilli(authClient.CreatedAt).UTC(),
		UpdatedAt:    updatedAt,
	}
	return res, nil
}

func (r *auth) UpdateClient(ctx context.Context, p repository.UpdateClientParam) (*repository.UpdateClientResult, error) {
	tx := r.gormClient.
		WithContext(ctx).
		Clauses(dbresolver.Write).
		Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	findRes := tx.
		Select(`id, client_id, name, type, status`).
		First(&AauthClient{}, "id = ?", p.Id)
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
		Model(&AauthClient{}).
		Where("id = ?", p.Id).
		Updates(map[string]interface{}{
			"client_id":  p.ClientId,
			"name":       p.Name,
			"type":       p.Type,
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

	authClient := &AauthClient{}
	checkRes := tx.
		Select(`id, client_id, client_secret, name, type, status, created_at, updated_at`).
		First(authClient, "id = ?", p.Id)
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

	res := &repository.UpdateClientResult{
		Id:           authClient.Id,
		ClientId:     authClient.ClientId,
		ClientSecret: authClient.ClientSecret,
		Name:         authClient.Name,
		Type:         authClient.Type,
		Status:       authClient.Status,
		CreatedAt:    time.UnixMilli(authClient.CreatedAt).UTC(),
		UpdatedAt:    time.UnixMilli(authClient.UpdatedAt.Int64).UTC(),
	}
	return res, nil
}

type AuthParam struct {
	GormClient *gorm.DB
}

func NewAuth(p AuthParam) *auth {
	return &auth{
		gormClient: p.GormClient,
	}
}

type AauthClient struct {
	Id           string        `gorm:"column:id;primaryKey"`
	ClientId     string        `gorm:"column:client_id"`
	ClientSecret string        `gorm:"column:client_secret"`
	Name         string        `gorm:"column:name"`
	Type         string        `gorm:"column:type"`
	Status       string        `gorm:"column:status"`
	CreatedAt    int64         `gorm:"column:created_at"`
	UpdatedAt    sql.NullInt64 `gorm:"column:updated_at;autoUpdateTime:milli;<-:update"`
}

func (AauthClient) TableName() string {
	return "auth_client"
}
