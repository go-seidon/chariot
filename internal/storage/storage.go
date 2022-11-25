package storage

import (
	"context"
	"io"
	"time"
)

const (
	PROVIDER_GOSEIDON_HIPPO = "goseidon_hippo"
	PROVIDER_AWS_S3         = "aws_s3"
	PROVIDER_GCLOUD_STORAGE = "gcloud_storage"
	PROVIDER_ALICLOUD_OSS   = "alicloud_oss"
)

type Storage interface {
	UploadObject(ctx context.Context, p UploadObjectParam) (*UploadObjectResult, error)
}

type UploadObjectParam struct {
	Data      io.Reader
	Id        *string //optional, some provider might require this field
	Name      *string
	Mimetype  *string
	Extension *string
}

type UploadObjectResult struct {
	ObjectId   string    //required, this id might be different than the specified id
	UploadedAt time.Time //required, represented in UTC format
}
