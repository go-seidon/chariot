package hippo

type ResponseBodyInfo struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

type UploadObjectResponseBody struct {
	Code    int32                    `json:"code"`
	Message string                   `json:"message"`
	Data    UploadObjectResponseData `json:"data"`
}

type UploadObjectResponseData struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	Extension  string `json:"extension"`
	Mimetype   string `json:"mimetype"`
	Size       int64  `json:"size"`
	UploadedAt int64  `json:"uploaded_at"`
}
