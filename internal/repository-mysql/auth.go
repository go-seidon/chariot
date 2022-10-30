package repository_mysql

import (
	"context"
	"time"

	"github.com/go-seidon/chariot/internal/repository"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

type auth struct {
	dbClient *gorm.DB
}

func (r *auth) CreateClient(ctx context.Context, p repository.CreateClientParam) (*repository.CreateClientResult, error) {
	tx := r.dbClient.
		WithContext(ctx).
		Clauses(dbresolver.Write).
		Begin()
	if tx.Error != nil {
		return nil, tx.Error
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
		CreatedAt:    time.UnixMilli(authClient.CreatedAt),
	}
	return res, nil
}

type AuthParam struct {
	DbClient *gorm.DB
}

func NewAuth(p AuthParam) *auth {
	return &auth{
		dbClient: p.DbClient,
	}
}

type AauthClient struct {
	Id           string `gorm:"column:id;primaryKey"`
	ClientId     string `gorm:"column:client_id"`
	ClientSecret string `gorm:"column:client_secret"`
	Name         string `gorm:"column:name"`
	Type         string `gorm:"column:type"`
	Status       string `gorm:"column:status"`
	CreatedAt    int64  `gorm:"column:created_at"`
	UpdatedAt    int64  `gorm:"column:updated_at;autoUpdateTime:milli;<-:update"`
}

func (AauthClient) TableName() string {
	return "auth_client"
}
