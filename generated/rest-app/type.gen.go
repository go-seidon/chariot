// Package rest_v1 provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.11.0 DO NOT EDIT.
package rest_v1

import (
	"encoding/json"
	"fmt"
)

const (
	BasicAuthScopes = "basicAuth.Scopes"
)

// Defines values for CreateAuthClientRequestStatus.
const (
	CreateAuthClientRequestStatusActive   CreateAuthClientRequestStatus = "active"
	CreateAuthClientRequestStatusInactive CreateAuthClientRequestStatus = "inactive"
)

// Defines values for CreateAuthClientRequestType.
const (
	CreateAuthClientRequestTypeBasicAuth CreateAuthClientRequestType = "basic_auth"
)

// Defines values for GetFileByIdDataStatus.
const (
	GetFileByIdDataStatusAvailable GetFileByIdDataStatus = "available"
	GetFileByIdDataStatusDeleted   GetFileByIdDataStatus = "deleted"
)

// Defines values for GetFileByIdDataVisibility.
const (
	GetFileByIdDataVisibilityProtected GetFileByIdDataVisibility = "protected"
	GetFileByIdDataVisibilityPublic    GetFileByIdDataVisibility = "public"
)

// Defines values for SearchAuthClientFilterStatusIn.
const (
	SearchAuthClientFilterStatusInActive   SearchAuthClientFilterStatusIn = "active"
	SearchAuthClientFilterStatusInInactive SearchAuthClientFilterStatusIn = "inactive"
)

// Defines values for SearchBarrelFilterProviderIn.
const (
	AlicloudOss   SearchBarrelFilterProviderIn = "alicloud_oss"
	AwsS3         SearchBarrelFilterProviderIn = "aws_s3"
	GcloudStroage SearchBarrelFilterProviderIn = "gcloud_stroage"
	GoseidonHippo SearchBarrelFilterProviderIn = "goseidon_hippo"
)

// Defines values for SearchBarrelFilterStatusIn.
const (
	SearchBarrelFilterStatusInActive   SearchBarrelFilterStatusIn = "active"
	SearchBarrelFilterStatusInInactive SearchBarrelFilterStatusIn = "inactive"
)

// Defines values for SearchFileFilterStatusIn.
const (
	SearchFileFilterStatusInActive   SearchFileFilterStatusIn = "active"
	SearchFileFilterStatusInInactive SearchFileFilterStatusIn = "inactive"
)

// Defines values for SearchFileFilterVisibilityIn.
const (
	SearchFileFilterVisibilityInProtected SearchFileFilterVisibilityIn = "protected"
	SearchFileFilterVisibilityInPublic    SearchFileFilterVisibilityIn = "public"
)

// Defines values for SearchFileItemStatus.
const (
	SearchFileItemStatusAvailable SearchFileItemStatus = "available"
	SearchFileItemStatusDeleted   SearchFileItemStatus = "deleted"
)

// Defines values for SearchFileItemVisibility.
const (
	SearchFileItemVisibilityProtected SearchFileItemVisibility = "protected"
	SearchFileItemVisibilityPublic    SearchFileItemVisibility = "public"
)

// Defines values for UpdateAuthClientByIdRequestStatus.
const (
	Active   UpdateAuthClientByIdRequestStatus = "active"
	Inactive UpdateAuthClientByIdRequestStatus = "inactive"
)

// Defines values for UpdateAuthClientByIdRequestType.
const (
	UpdateAuthClientByIdRequestTypeBasicAuth UpdateAuthClientByIdRequestType = "basic_auth"
)

// Defines values for UploadFileDataStatus.
const (
	Available UploadFileDataStatus = "available"
	Deleted   UploadFileDataStatus = "deleted"
	Deleting  UploadFileDataStatus = "deleting"
	Uploading UploadFileDataStatus = "uploading"
)

// Defines values for UploadFileDataVisibility.
const (
	UploadFileDataVisibilityProtected UploadFileDataVisibility = "protected"
	UploadFileDataVisibilityPublic    UploadFileDataVisibility = "public"
)

// Defines values for UploadFileRequestVisibility.
const (
	Protected UploadFileRequestVisibility = "protected"
	Public    UploadFileRequestVisibility = "public"
)

// CreateAuthClientData defines model for CreateAuthClientData.
type CreateAuthClientData struct {
	ClientId  string `json:"client_id"`
	CreatedAt int64  `json:"created_at"`
	Id        string `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Type      string `json:"type"`
}

// CreateAuthClientRequest defines model for CreateAuthClientRequest.
type CreateAuthClientRequest struct {
	ClientId     string                        `json:"client_id"`
	ClientSecret string                        `json:"client_secret"`
	Name         string                        `json:"name"`
	Status       CreateAuthClientRequestStatus `json:"status"`
	Type         CreateAuthClientRequestType   `json:"type"`
}

// CreateAuthClientRequestStatus defines model for CreateAuthClientRequest.Status.
type CreateAuthClientRequestStatus string

// CreateAuthClientRequestType defines model for CreateAuthClientRequest.Type.
type CreateAuthClientRequestType string

// CreateAuthClientResponse defines model for CreateAuthClientResponse.
type CreateAuthClientResponse struct {
	Code    int32                `json:"code"`
	Data    CreateAuthClientData `json:"data"`
	Message string               `json:"message"`
}

// CreateBarrelData defines model for CreateBarrelData.
type CreateBarrelData struct {
	Code      string `json:"code"`
	CreatedAt int64  `json:"created_at"`
	Id        string `json:"id"`
	Name      string `json:"name"`
	Provider  string `json:"provider"`
	Status    string `json:"status"`
}

// CreateBarrelRequest defines model for CreateBarrelRequest.
type CreateBarrelRequest struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Status   string `json:"status"`
}

// CreateBarrelResponse defines model for CreateBarrelResponse.
type CreateBarrelResponse struct {
	Code    int32            `json:"code"`
	Data    CreateBarrelData `json:"data"`
	Message string           `json:"message"`
}

// DeleteFileByIdData defines model for DeleteFileByIdData.
type DeleteFileByIdData struct {
	RequestedAt int64 `json:"requested_at"`
}

// DeleteFileByIdResponse defines model for DeleteFileByIdResponse.
type DeleteFileByIdResponse struct {
	Code    int32              `json:"code"`
	Data    DeleteFileByIdData `json:"data"`
	Message string             `json:"message"`
}

// GetAppInfoData defines model for GetAppInfoData.
type GetAppInfoData struct {
	AppName    string `json:"app_name"`
	AppVersion string `json:"app_version"`
}

// GetAppInfoResponse defines model for GetAppInfoResponse.
type GetAppInfoResponse struct {
	Code    int32          `json:"code"`
	Data    GetAppInfoData `json:"data"`
	Message string         `json:"message"`
}

// GetAuthClientByIdData defines model for GetAuthClientByIdData.
type GetAuthClientByIdData struct {
	ClientId  string `json:"client_id"`
	CreatedAt int64  `json:"created_at"`
	Id        string `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Type      string `json:"type"`
	UpdatedAt *int64 `json:"updated_at,omitempty"`
}

// GetAuthClientByIdResponse defines model for GetAuthClientByIdResponse.
type GetAuthClientByIdResponse struct {
	Code    int32                 `json:"code"`
	Data    GetAuthClientByIdData `json:"data"`
	Message string                `json:"message"`
}

// GetBarrelByIdData defines model for GetBarrelByIdData.
type GetBarrelByIdData struct {
	Code      string `json:"code"`
	CreatedAt int64  `json:"created_at"`
	Id        string `json:"id"`
	Name      string `json:"name"`
	Provider  string `json:"provider"`
	Status    string `json:"status"`
	UpdatedAt *int64 `json:"updated_at,omitempty"`
}

// GetBarrelByIdResponse defines model for GetBarrelByIdResponse.
type GetBarrelByIdResponse struct {
	Code    int32             `json:"code"`
	Data    GetBarrelByIdData `json:"data"`
	Message string            `json:"message"`
}

// GetFileByIdData defines model for GetFileByIdData.
type GetFileByIdData struct {
	CreatedAt  int64                     `json:"created_at"`
	DeletedAt  *int64                    `json:"deleted_at,omitempty"`
	Extension  string                    `json:"extension"`
	Id         string                    `json:"id"`
	Meta       *GetFileByIdData_Meta     `json:"meta,omitempty"`
	Mimetype   string                    `json:"mimetype"`
	Name       string                    `json:"name"`
	Size       int                       `json:"size"`
	Slug       string                    `json:"slug"`
	Status     GetFileByIdDataStatus     `json:"status"`
	UpdatedAt  *int64                    `json:"updated_at,omitempty"`
	UploadedAt int64                     `json:"uploaded_at"`
	Visibility GetFileByIdDataVisibility `json:"visibility"`
}

// GetFileByIdData_Meta defines model for GetFileByIdData.Meta.
type GetFileByIdData_Meta struct {
	AdditionalProperties map[string]string `json:"-"`
}

// GetFileByIdDataStatus defines model for GetFileByIdData.Status.
type GetFileByIdDataStatus string

// GetFileByIdDataVisibility defines model for GetFileByIdData.Visibility.
type GetFileByIdDataVisibility string

// GetFileByIdResponse defines model for GetFileByIdResponse.
type GetFileByIdResponse struct {
	Code    int32           `json:"code"`
	Data    GetFileByIdData `json:"data"`
	Message string          `json:"message"`
}

// RequestPagination defines model for RequestPagination.
type RequestPagination struct {
	// min = 1
	Page int64 `json:"page"`

	// min = 1, max = 200
	TotalItems int32 `json:"total_items"`
}

// ResponseBodyInfo defines model for ResponseBodyInfo.
type ResponseBodyInfo struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

// RetrieveFileBySlugResponse defines model for RetrieveFileBySlugResponse.
type RetrieveFileBySlugResponse = string

// SearchAuthClientData defines model for SearchAuthClientData.
type SearchAuthClientData struct {
	Items   []SearchAuthClientItem  `json:"items"`
	Summary SearchAuthClientSummary `json:"summary"`
}

// SearchAuthClientFilter defines model for SearchAuthClientFilter.
type SearchAuthClientFilter struct {
	StatusIn *[]SearchAuthClientFilterStatusIn `json:"status_in,omitempty"`
}

// auth client status
type SearchAuthClientFilterStatusIn string

// SearchAuthClientItem defines model for SearchAuthClientItem.
type SearchAuthClientItem struct {
	ClientId  string `json:"client_id"`
	CreatedAt int64  `json:"created_at"`
	Id        string `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Type      string `json:"type"`
	UpdatedAt *int64 `json:"updated_at,omitempty"`
}

// SearchAuthClientRequest defines model for SearchAuthClientRequest.
type SearchAuthClientRequest struct {
	Filter *SearchAuthClientFilter `json:"filter,omitempty"`

	// min = 2 character
	Keyword    *string            `json:"keyword,omitempty"`
	Pagination *RequestPagination `json:"pagination,omitempty"`
}

// SearchAuthClientResponse defines model for SearchAuthClientResponse.
type SearchAuthClientResponse struct {
	Code    int32                `json:"code"`
	Data    SearchAuthClientData `json:"data"`
	Message string               `json:"message"`
}

// SearchAuthClientSummary defines model for SearchAuthClientSummary.
type SearchAuthClientSummary struct {
	// current page
	Page int64 `json:"page"`

	// total matched items with a given parameter
	TotalItems int64 `json:"total_items"`
}

// SearchBarrelData defines model for SearchBarrelData.
type SearchBarrelData struct {
	Items   []SearchBarrelItem  `json:"items"`
	Summary SearchBarrelSummary `json:"summary"`
}

// SearchBarrelFilter defines model for SearchBarrelFilter.
type SearchBarrelFilter struct {
	ProviderIn *[]SearchBarrelFilterProviderIn `json:"provider_in,omitempty"`
	StatusIn   *[]SearchBarrelFilterStatusIn   `json:"status_in,omitempty"`
}

// barrel provider
type SearchBarrelFilterProviderIn string

// barrel status
type SearchBarrelFilterStatusIn string

// SearchBarrelItem defines model for SearchBarrelItem.
type SearchBarrelItem struct {
	Code      string `json:"code"`
	CreatedAt int64  `json:"created_at"`
	Id        string `json:"id"`
	Name      string `json:"name"`
	Provider  string `json:"provider"`
	Status    string `json:"status"`
	UpdatedAt *int64 `json:"updated_at,omitempty"`
}

// SearchBarrelRequest defines model for SearchBarrelRequest.
type SearchBarrelRequest struct {
	Filter *SearchBarrelFilter `json:"filter,omitempty"`

	// min = 2 character
	Keyword    *string            `json:"keyword,omitempty"`
	Pagination *RequestPagination `json:"pagination,omitempty"`
}

// SearchBarrelResponse defines model for SearchBarrelResponse.
type SearchBarrelResponse struct {
	Code    int32            `json:"code"`
	Data    SearchBarrelData `json:"data"`
	Message string           `json:"message"`
}

// SearchBarrelSummary defines model for SearchBarrelSummary.
type SearchBarrelSummary struct {
	// current page
	Page int64 `json:"page"`

	// total matched items with a given parameter
	TotalItems int64 `json:"total_items"`
}

// SearchFileFilter defines model for SearchFileFilter.
type SearchFileFilter struct {
	ExtensionIn   *[]string                       `json:"extension_in,omitempty"`
	SizeGte       *int                            `json:"size_gte,omitempty"`
	SizeLte       *int                            `json:"size_lte,omitempty"`
	StatusIn      *[]SearchFileFilterStatusIn     `json:"status_in,omitempty"`
	UploadDateGte *int64                          `json:"upload_date_gte,omitempty"`
	UploadDateLte *int64                          `json:"upload_date_lte,omitempty"`
	VisibilityIn  *[]SearchFileFilterVisibilityIn `json:"visibility_in,omitempty"`
}

// file status
type SearchFileFilterStatusIn string

// file visibility
type SearchFileFilterVisibilityIn string

// SearchFileItem defines model for SearchFileItem.
type SearchFileItem struct {
	CreatedAt  int64                    `json:"created_at"`
	DeletedAt  *int64                   `json:"deleted_at,omitempty"`
	Extension  string                   `json:"extension"`
	Id         string                   `json:"id"`
	Meta       *SearchFileItem_Meta     `json:"meta,omitempty"`
	Mimetype   string                   `json:"mimetype"`
	Name       string                   `json:"name"`
	Size       int                      `json:"size"`
	Slug       string                   `json:"slug"`
	Status     SearchFileItemStatus     `json:"status"`
	UpdatedAt  *int64                   `json:"updated_at,omitempty"`
	UploadedAt int64                    `json:"uploaded_at"`
	Visibility SearchFileItemVisibility `json:"visibility"`
}

// SearchFileItem_Meta defines model for SearchFileItem.Meta.
type SearchFileItem_Meta struct {
	AdditionalProperties map[string]string `json:"-"`
}

// SearchFileItemStatus defines model for SearchFileItem.Status.
type SearchFileItemStatus string

// SearchFileItemVisibility defines model for SearchFileItem.Visibility.
type SearchFileItemVisibility string

// SearchFileRequest defines model for SearchFileRequest.
type SearchFileRequest struct {
	Filter *SearchFileFilter `json:"filter,omitempty"`

	// min = 2 character
	Keyword    *string            `json:"keyword,omitempty"`
	Pagination *RequestPagination `json:"pagination,omitempty"`
}

// SearchFileResponse defines model for SearchFileResponse.
type SearchFileResponse struct {
	Code    int32            `json:"code"`
	Data    []SearchFileItem `json:"data"`
	Message string           `json:"message"`
}

// UpdateAuthClientByIdData defines model for UpdateAuthClientByIdData.
type UpdateAuthClientByIdData struct {
	ClientId  string `json:"client_id"`
	CreatedAt int64  `json:"created_at"`
	Id        string `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Type      string `json:"type"`
	UpdatedAt int64  `json:"updated_at"`
}

// UpdateAuthClientByIdRequest defines model for UpdateAuthClientByIdRequest.
type UpdateAuthClientByIdRequest struct {
	ClientId string                            `json:"client_id"`
	Name     string                            `json:"name"`
	Status   UpdateAuthClientByIdRequestStatus `json:"status"`
	Type     UpdateAuthClientByIdRequestType   `json:"type"`
}

// UpdateAuthClientByIdRequestStatus defines model for UpdateAuthClientByIdRequest.Status.
type UpdateAuthClientByIdRequestStatus string

// UpdateAuthClientByIdRequestType defines model for UpdateAuthClientByIdRequest.Type.
type UpdateAuthClientByIdRequestType string

// UpdateAuthClientByIdResponse defines model for UpdateAuthClientByIdResponse.
type UpdateAuthClientByIdResponse struct {
	Code    int32                    `json:"code"`
	Data    UpdateAuthClientByIdData `json:"data"`
	Message string                   `json:"message"`
}

// UpdateBarrelByIdData defines model for UpdateBarrelByIdData.
type UpdateBarrelByIdData struct {
	Code      string `json:"code"`
	CreatedAt int64  `json:"created_at"`
	Id        string `json:"id"`
	Name      string `json:"name"`
	Provider  string `json:"provider"`
	Status    string `json:"status"`
	UpdatedAt int64  `json:"updated_at"`
}

// UpdateBarrelByIdRequest defines model for UpdateBarrelByIdRequest.
type UpdateBarrelByIdRequest struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Status   string `json:"status"`
}

// UpdateBarrelByIdResponse defines model for UpdateBarrelByIdResponse.
type UpdateBarrelByIdResponse struct {
	Code    int32                `json:"code"`
	Data    UpdateBarrelByIdData `json:"data"`
	Message string               `json:"message"`
}

// UploadFileData defines model for UploadFileData.
type UploadFileData struct {
	Extension  string                   `json:"extension"`
	Id         string                   `json:"id"`
	Meta       *UploadFileData_Meta     `json:"meta,omitempty"`
	Mimetype   string                   `json:"mimetype"`
	Name       string                   `json:"name"`
	Size       int64                    `json:"size"`
	Slug       string                   `json:"slug"`
	Status     UploadFileDataStatus     `json:"status"`
	UploadedAt int64                    `json:"uploaded_at"`
	Visibility UploadFileDataVisibility `json:"visibility"`
}

// UploadFileData_Meta defines model for UploadFileData.Meta.
type UploadFileData_Meta struct {
	AdditionalProperties map[string]string `json:"-"`
}

// UploadFileDataStatus defines model for UploadFileData.Status.
type UploadFileDataStatus string

// UploadFileDataVisibility defines model for UploadFileData.Visibility.
type UploadFileDataVisibility string

// UploadFileRequest defines model for UploadFileRequest.
type UploadFileRequest struct {
	// csv of barrel code
	Barrels string `json:"barrels"`
	File    string `json:"file"`

	// json metadata
	Meta       *string                     `json:"meta,omitempty"`
	Visibility UploadFileRequestVisibility `json:"visibility"`
}

// UploadFileRequestVisibility defines model for UploadFileRequest.Visibility.
type UploadFileRequestVisibility string

// UploadFileResponse defines model for UploadFileResponse.
type UploadFileResponse struct {
	Code    int32          `json:"code"`
	Data    UploadFileData `json:"data"`
	Message string         `json:"message"`
}

// CorrelationId defines model for CorrelationId.
type CorrelationId = string

// ObjectId defines model for ObjectId.
type ObjectId = string

// ObjectSlug defines model for ObjectSlug.
type ObjectSlug = string

// BadRequest defines model for BadRequest.
type BadRequest = ResponseBodyInfo

// ForbiddenAccess defines model for ForbiddenAccess.
type ForbiddenAccess = ResponseBodyInfo

// NotFound defines model for NotFound.
type NotFound = ResponseBodyInfo

// ServerError defines model for ServerError.
type ServerError = ResponseBodyInfo

// UnauthenticatedAccess defines model for UnauthenticatedAccess.
type UnauthenticatedAccess = ResponseBodyInfo

// GetAppInfoParams defines parameters for GetAppInfo.
type GetAppInfoParams struct {
	// correlation id for tracing purposes
	XCorrelationId *CorrelationId `json:"X-Correlation-Id,omitempty"`
}

// UploadFileParams defines parameters for UploadFile.
type UploadFileParams struct {
	// correlation id for tracing purposes
	XCorrelationId *CorrelationId `json:"X-Correlation-Id,omitempty"`
}

// RetrieveFileBySlugParams defines parameters for RetrieveFileBySlug.
type RetrieveFileBySlugParams struct {
	// correlation id for tracing purposes
	XCorrelationId *CorrelationId `json:"X-Correlation-Id,omitempty"`
}

// CreateAuthClientJSONBody defines parameters for CreateAuthClient.
type CreateAuthClientJSONBody = CreateAuthClientRequest

// CreateAuthClientParams defines parameters for CreateAuthClient.
type CreateAuthClientParams struct {
	// correlation id for tracing purposes
	XCorrelationId *CorrelationId `json:"X-Correlation-Id,omitempty"`
}

// SearchAuthClientJSONBody defines parameters for SearchAuthClient.
type SearchAuthClientJSONBody = SearchAuthClientRequest

// SearchAuthClientParams defines parameters for SearchAuthClient.
type SearchAuthClientParams struct {
	// correlation id for tracing purposes
	XCorrelationId *CorrelationId `json:"X-Correlation-Id,omitempty"`
}

// GetAuthClientByIdParams defines parameters for GetAuthClientById.
type GetAuthClientByIdParams struct {
	// correlation id for tracing purposes
	XCorrelationId *CorrelationId `json:"X-Correlation-Id,omitempty"`
}

// UpdateAuthClientByIdJSONBody defines parameters for UpdateAuthClientById.
type UpdateAuthClientByIdJSONBody = UpdateAuthClientByIdRequest

// UpdateAuthClientByIdParams defines parameters for UpdateAuthClientById.
type UpdateAuthClientByIdParams struct {
	// correlation id for tracing purposes
	XCorrelationId *CorrelationId `json:"X-Correlation-Id,omitempty"`
}

// CreateBarrelJSONBody defines parameters for CreateBarrel.
type CreateBarrelJSONBody = CreateBarrelRequest

// CreateBarrelParams defines parameters for CreateBarrel.
type CreateBarrelParams struct {
	// correlation id for tracing purposes
	XCorrelationId *CorrelationId `json:"X-Correlation-Id,omitempty"`
}

// SearchBarrelJSONBody defines parameters for SearchBarrel.
type SearchBarrelJSONBody = SearchBarrelRequest

// SearchBarrelParams defines parameters for SearchBarrel.
type SearchBarrelParams struct {
	// correlation id for tracing purposes
	XCorrelationId *CorrelationId `json:"X-Correlation-Id,omitempty"`
}

// GetBarrelByIdParams defines parameters for GetBarrelById.
type GetBarrelByIdParams struct {
	// correlation id for tracing purposes
	XCorrelationId *CorrelationId `json:"X-Correlation-Id,omitempty"`
}

// UpdateBarrelByIdJSONBody defines parameters for UpdateBarrelById.
type UpdateBarrelByIdJSONBody = UpdateBarrelByIdRequest

// UpdateBarrelByIdParams defines parameters for UpdateBarrelById.
type UpdateBarrelByIdParams struct {
	// correlation id for tracing purposes
	XCorrelationId *CorrelationId `json:"X-Correlation-Id,omitempty"`
}

// SearchFileJSONBody defines parameters for SearchFile.
type SearchFileJSONBody = SearchFileRequest

// SearchFileParams defines parameters for SearchFile.
type SearchFileParams struct {
	// correlation id for tracing purposes
	XCorrelationId *CorrelationId `json:"X-Correlation-Id,omitempty"`
}

// DeleteFileByIdParams defines parameters for DeleteFileById.
type DeleteFileByIdParams struct {
	// correlation id for tracing purposes
	XCorrelationId *CorrelationId `json:"X-Correlation-Id,omitempty"`
}

// GetFileByIdParams defines parameters for GetFileById.
type GetFileByIdParams struct {
	// correlation id for tracing purposes
	XCorrelationId *CorrelationId `json:"X-Correlation-Id,omitempty"`
}

// CreateAuthClientJSONRequestBody defines body for CreateAuthClient for application/json ContentType.
type CreateAuthClientJSONRequestBody = CreateAuthClientJSONBody

// SearchAuthClientJSONRequestBody defines body for SearchAuthClient for application/json ContentType.
type SearchAuthClientJSONRequestBody = SearchAuthClientJSONBody

// UpdateAuthClientByIdJSONRequestBody defines body for UpdateAuthClientById for application/json ContentType.
type UpdateAuthClientByIdJSONRequestBody = UpdateAuthClientByIdJSONBody

// CreateBarrelJSONRequestBody defines body for CreateBarrel for application/json ContentType.
type CreateBarrelJSONRequestBody = CreateBarrelJSONBody

// SearchBarrelJSONRequestBody defines body for SearchBarrel for application/json ContentType.
type SearchBarrelJSONRequestBody = SearchBarrelJSONBody

// UpdateBarrelByIdJSONRequestBody defines body for UpdateBarrelById for application/json ContentType.
type UpdateBarrelByIdJSONRequestBody = UpdateBarrelByIdJSONBody

// SearchFileJSONRequestBody defines body for SearchFile for application/json ContentType.
type SearchFileJSONRequestBody = SearchFileJSONBody

// Getter for additional properties for GetFileByIdData_Meta. Returns the specified
// element and whether it was found
func (a GetFileByIdData_Meta) Get(fieldName string) (value string, found bool) {
	if a.AdditionalProperties != nil {
		value, found = a.AdditionalProperties[fieldName]
	}
	return
}

// Setter for additional properties for GetFileByIdData_Meta
func (a *GetFileByIdData_Meta) Set(fieldName string, value string) {
	if a.AdditionalProperties == nil {
		a.AdditionalProperties = make(map[string]string)
	}
	a.AdditionalProperties[fieldName] = value
}

// Override default JSON handling for GetFileByIdData_Meta to handle AdditionalProperties
func (a *GetFileByIdData_Meta) UnmarshalJSON(b []byte) error {
	object := make(map[string]json.RawMessage)
	err := json.Unmarshal(b, &object)
	if err != nil {
		return err
	}

	if len(object) != 0 {
		a.AdditionalProperties = make(map[string]string)
		for fieldName, fieldBuf := range object {
			var fieldVal string
			err := json.Unmarshal(fieldBuf, &fieldVal)
			if err != nil {
				return fmt.Errorf("error unmarshaling field %s: %w", fieldName, err)
			}
			a.AdditionalProperties[fieldName] = fieldVal
		}
	}
	return nil
}

// Override default JSON handling for GetFileByIdData_Meta to handle AdditionalProperties
func (a GetFileByIdData_Meta) MarshalJSON() ([]byte, error) {
	var err error
	object := make(map[string]json.RawMessage)

	for fieldName, field := range a.AdditionalProperties {
		object[fieldName], err = json.Marshal(field)
		if err != nil {
			return nil, fmt.Errorf("error marshaling '%s': %w", fieldName, err)
		}
	}
	return json.Marshal(object)
}

// Getter for additional properties for SearchFileItem_Meta. Returns the specified
// element and whether it was found
func (a SearchFileItem_Meta) Get(fieldName string) (value string, found bool) {
	if a.AdditionalProperties != nil {
		value, found = a.AdditionalProperties[fieldName]
	}
	return
}

// Setter for additional properties for SearchFileItem_Meta
func (a *SearchFileItem_Meta) Set(fieldName string, value string) {
	if a.AdditionalProperties == nil {
		a.AdditionalProperties = make(map[string]string)
	}
	a.AdditionalProperties[fieldName] = value
}

// Override default JSON handling for SearchFileItem_Meta to handle AdditionalProperties
func (a *SearchFileItem_Meta) UnmarshalJSON(b []byte) error {
	object := make(map[string]json.RawMessage)
	err := json.Unmarshal(b, &object)
	if err != nil {
		return err
	}

	if len(object) != 0 {
		a.AdditionalProperties = make(map[string]string)
		for fieldName, fieldBuf := range object {
			var fieldVal string
			err := json.Unmarshal(fieldBuf, &fieldVal)
			if err != nil {
				return fmt.Errorf("error unmarshaling field %s: %w", fieldName, err)
			}
			a.AdditionalProperties[fieldName] = fieldVal
		}
	}
	return nil
}

// Override default JSON handling for SearchFileItem_Meta to handle AdditionalProperties
func (a SearchFileItem_Meta) MarshalJSON() ([]byte, error) {
	var err error
	object := make(map[string]json.RawMessage)

	for fieldName, field := range a.AdditionalProperties {
		object[fieldName], err = json.Marshal(field)
		if err != nil {
			return nil, fmt.Errorf("error marshaling '%s': %w", fieldName, err)
		}
	}
	return json.Marshal(object)
}

// Getter for additional properties for UploadFileData_Meta. Returns the specified
// element and whether it was found
func (a UploadFileData_Meta) Get(fieldName string) (value string, found bool) {
	if a.AdditionalProperties != nil {
		value, found = a.AdditionalProperties[fieldName]
	}
	return
}

// Setter for additional properties for UploadFileData_Meta
func (a *UploadFileData_Meta) Set(fieldName string, value string) {
	if a.AdditionalProperties == nil {
		a.AdditionalProperties = make(map[string]string)
	}
	a.AdditionalProperties[fieldName] = value
}

// Override default JSON handling for UploadFileData_Meta to handle AdditionalProperties
func (a *UploadFileData_Meta) UnmarshalJSON(b []byte) error {
	object := make(map[string]json.RawMessage)
	err := json.Unmarshal(b, &object)
	if err != nil {
		return err
	}

	if len(object) != 0 {
		a.AdditionalProperties = make(map[string]string)
		for fieldName, fieldBuf := range object {
			var fieldVal string
			err := json.Unmarshal(fieldBuf, &fieldVal)
			if err != nil {
				return fmt.Errorf("error unmarshaling field %s: %w", fieldName, err)
			}
			a.AdditionalProperties[fieldName] = fieldVal
		}
	}
	return nil
}

// Override default JSON handling for UploadFileData_Meta to handle AdditionalProperties
func (a UploadFileData_Meta) MarshalJSON() ([]byte, error) {
	var err error
	object := make(map[string]json.RawMessage)

	for fieldName, field := range a.AdditionalProperties {
		object[fieldName], err = json.Marshal(field)
		if err != nil {
			return nil, fmt.Errorf("error marshaling '%s': %w", fieldName, err)
		}
	}
	return json.Marshal(object)
}
