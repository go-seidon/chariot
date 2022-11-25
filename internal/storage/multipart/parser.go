package multipart

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
)

type Parser = func(h *multipart.FileHeader) (*FileInfo, error)

type FileInfo struct {
	Name      string
	Size      int64
	Extension string
	Mimetype  string
	Data      multipart.File
}

func FileParser(h *multipart.FileHeader) (*FileInfo, error) {
	if h == nil {
		return nil, fmt.Errorf("invalid header")
	}

	data, err := h.Open()
	if err != nil {
		return nil, err
	}

	buff := make([]byte, 512)
	n, err := data.Read(buff)
	if err != nil && err != io.EOF {
		return nil, err
	}
	buff = buff[:n]

	_, err = data.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	info := &FileInfo{
		Size:      h.Size,
		Name:      FileName(h),
		Extension: FileExtension(h),
		Mimetype:  http.DetectContentType(buff),
		Data:      data,
	}
	return info, nil
}

func FileName(fh *multipart.FileHeader) string {
	names := strings.Split(fh.Filename, ".")
	return names[0]
}

func FileExtension(fh *multipart.FileHeader) string {
	names := strings.Split(fh.Filename, ".")
	if len(names) == 1 {
		return ""
	}
	return names[len(names)-1]
}
