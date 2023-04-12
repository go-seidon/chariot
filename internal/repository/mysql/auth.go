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

	currentClient := &AuthClient{}
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

	createParam := &AuthClient{
		Id:           p.Id,
		ClientId:     p.ClientId,
		ClientSecret: p.ClientSecret,
		Name:         p.Name,
		Type:         p.Type,
		Status:       p.Status,
		CreatedAt:    p.CreatedAt.UnixMilli(),
		UpdatedAt:    p.CreatedAt.UnixMilli(),
	}
	createRes := tx.Create(createParam)
	if createRes.Error != nil {
		txRes := tx.Rollback()
		if txRes.Error != nil {
			return nil, txRes.Error
		}
		return nil, createRes.Error
	}

	authClient := &AuthClient{}
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
	authClient := &AuthClient{}

	query := r.gormClient.
		WithContext(ctx).
		Clauses(dbresolver.Read)

	findRes := query.Select(`id, client_id, client_secret, name, type, status, created_at, updated_at`)
	if p.ClientId != "" {
		findRes = findRes.First(authClient, "client_id = ?", p.ClientId)
	} else {
		findRes = findRes.First(authClient, "id = ?", p.Id)
	}

	if findRes.Error != nil {
		if errors.Is(findRes.Error, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, findRes.Error
	}

	res := &repository.FindClientResult{
		Id:           authClient.Id,
		ClientId:     authClient.ClientId,
		ClientSecret: authClient.ClientSecret,
		Name:         authClient.Name,
		Type:         authClient.Type,
		Status:       authClient.Status,
		CreatedAt:    time.UnixMilli(authClient.CreatedAt).UTC(),
		UpdatedAt:    typeconv.Time(time.UnixMilli(authClient.UpdatedAt).UTC()),
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
		First(&AuthClient{}, "id = ?", p.Id)
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
		Model(&AuthClient{}).
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

	authClient := &AuthClient{}
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
		UpdatedAt:    time.UnixMilli(authClient.UpdatedAt).UTC(),
	}
	return res, nil
}

func (r *auth) SearchClient(ctx context.Context, p repository.SearchClientParam) (*repository.SearchClientResult, error) {
	query := r.gormClient.
		WithContext(ctx).
		Clauses(dbresolver.Read)

	if len(p.Statuses) > 0 {
		query.Where("status IN ?", p.Statuses)
	}

	if p.Keyword != "" {
		keyword := "%" + p.Keyword + "%"
		query.Where(
			r.gormClient.
				Where("name LIKE ?", keyword).
				Or("client_id LIKE ?", keyword),
		)
	}

	res := &repository.SearchClientResult{
		Summary: repository.SearchClientSummary{},
		Items:   []repository.SearchClientItem{},
	}
	countRes := query.
		Table("auth_client").
		Count(&res.Summary.TotalItems)
	if countRes.Error != nil {
		return nil, countRes.Error
	}

	if p.Limit > 0 {
		query.Limit(int(p.Limit))
	}

	if p.Offset > 0 {
		query.Offset(int(p.Offset))
	}

	authClients := []AuthClient{}
	searchRes := query.
		Select(`id, client_id, client_secret, name, type, status, created_at, updated_at`).
		Find(&authClients)

	if searchRes.Error != nil {
		if errors.Is(searchRes.Error, gorm.ErrRecordNotFound) {
			return res, nil
		}
		return nil, searchRes.Error
	}

	for _, authClient := range authClients {
		res.Items = append(res.Items, repository.SearchClientItem{
			Id:           authClient.Id,
			ClientId:     authClient.ClientId,
			ClientSecret: authClient.ClientSecret,
			Name:         authClient.Name,
			Type:         authClient.Type,
			Status:       authClient.Status,
			CreatedAt:    time.UnixMilli(authClient.CreatedAt).UTC(),
			UpdatedAt:    typeconv.Time(time.UnixMilli(authClient.UpdatedAt).UTC()),
		})
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

type AuthClient struct {
	Id           string `gorm:"column:id;primaryKey"`
	ClientId     string `gorm:"column:client_id"`
	ClientSecret string `gorm:"column:client_secret"`
	Name         string `gorm:"column:name"`
	Type         string `gorm:"column:type"`
	Status       string `gorm:"column:status"`
	CreatedAt    int64  `gorm:"column:created_at"`
	UpdatedAt    int64  `gorm:"column:updated_at;autoUpdateTime:milli"`
}

func (AuthClient) TableName() string {
	return "auth_client"
}
