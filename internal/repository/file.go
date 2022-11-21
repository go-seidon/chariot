package repository

import (
	"context"
	"time"
)

type File interface {
	CreateFile(ctx context.Context, p CreateFileParam) (*CreateFileResult, error)
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
