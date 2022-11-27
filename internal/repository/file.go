package repository

import (
	"context"
	"time"
)

type File interface {
	CreateFile(ctx context.Context, p CreateFileParam) (*CreateFileResult, error)
	FindFile(ctx context.Context, p FindFileParam) (*FindFileResult, error)
}

type CreateFileParam struct {
	Id         string
	Slug       string
	Name       string
	Mimetype   string
	Extension  string
	Size       int64
	Visibility string
	Status     string
	Meta       map[string]string
	CreatedAt  time.Time
	UploadedAt time.Time
	Locations  []CreateFileLocation
}

type CreateFileLocation struct {
	BarrelId   string
	ExternalId *string
	Priority   int32
	Status     string
	CreatedAt  time.Time
	UploadedAt *time.Time
}

type CreateFileResult struct {
	Id         string
	Slug       string
	Name       string
	Mimetype   string
	Extension  string
	Size       int64
	Visibility string
	Status     string
	Meta       map[string]string
	CreatedAt  time.Time
	UploadedAt time.Time
	Locations  []CreateFileLocation
}

type FindFileParam struct {
	Id   string
	Slug string
}

type FindFileResult struct {
	Id         string
	Slug       string
	Name       string
	Mimetype   string
	Extension  string
	Size       int64
	Visibility string
	Status     string
	CreatedAt  time.Time
	UpdatedAt  *time.Time
	UploadedAt time.Time
	DeletedAt  *time.Time
	Meta       map[string]string
	Locations  []FindFileLocation
}

type FindFileBarrel struct {
	Id       string
	Code     string
	Provider string
	Status   string
}

type FindFileLocation struct {
	Barrel     FindFileBarrel
	ExternalId *string
	Priority   int32
	Status     string
	CreatedAt  time.Time
	UpdatedAt  *time.Time
	UploadedAt *time.Time
}
