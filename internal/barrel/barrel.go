package barrel

const (
	STATUS_ACTIVE   = "active"
	STATUS_INACTIVE = "inactive"
)

const (
	PROVIDER_HIPPO    = "goseidon_hippo"
	PROVIDER_AWSS3    = "aws_s3"
	PROVIDER_GCLOUD   = "gcloud_storage"
	PROVIDER_ALICLOUD = "alicloud_oss"
)

var (
	SUPPORTED_PROVIDERS = map[string]bool{
		PROVIDER_HIPPO: true,
	}
)
