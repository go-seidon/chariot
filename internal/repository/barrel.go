package repository

import (
	"context"
	"time"
)

type Barrel interface {
	CreateBarrel(ctx context.Context, p CreateBarrelParam) (*CreateBarrelResult, error)
	FindBarrel(ctx context.Context, p FindBarrelParam) (*FindBarrelResult, error)
	UpdateBarrel(ctx context.Context, p UpdateBarrelParam) (*UpdateBarrelResult, error)
	SearchBarrel(ctx context.Context, p SearchBarrelParam) (*SearchBarrelResult, error)
}

type CreateBarrelParam struct {
	Id        string
	Code      string
	Name      string
	Provider  string
	Status    string
	CreatedAt time.Time
}

type CreateBarrelResult struct {
	Id        string
	Code      string
	Name      string
	Provider  string
	Status    string
	CreatedAt time.Time
}

type FindBarrelParam struct {
	Id string
}

type FindBarrelResult struct {
	Id        string
	Code      string
	Name      string
	Provider  string
	Status    string
	CreatedAt time.Time
	UpdatedAt *time.Time
}

type UpdateBarrelParam struct {
	Id        string
	Code      string
	Name      string
	Provider  string
	Status    string
	UpdatedAt time.Time
}

type UpdateBarrelResult struct {
	Id        string
	Code      string
	Name      string
	Provider  string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SearchBarrelParam struct {
	Limit     int32
	Offset    int64
	Keyword   string
	Statuses  []string
	Providers []string
	Codes     []string
}

type SearchBarrelResult struct {
	Summary SearchBarrelSummary
	Items   []SearchBarrelItem
}

// @note: return non-nil search items sorted by a given codes
func (r *SearchBarrelResult) SortCodes(codes []string) []SearchBarrelItem {
	if len(r.Items) == 0 {
		return r.Items
	}

	dc := map[string]SearchBarrelItem{}
	for _, item := range r.Items {
		dc[item.Code] = SearchBarrelItem{
			Id:        item.Id,
			Code:      item.Code,
			Name:      item.Name,
			Provider:  item.Provider,
			Status:    item.Status,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		}
	}

	barrels := []SearchBarrelItem{}
	for _, code := range codes {
		barrels = append(barrels, dc[code])
	}
	return barrels
}

type SearchBarrelSummary struct {
	TotalItems int64
}

type SearchBarrelItem struct {
	Id        string
	Code      string
	Name      string
	Provider  string
	Status    string
	CreatedAt time.Time
	UpdatedAt *time.Time
}
