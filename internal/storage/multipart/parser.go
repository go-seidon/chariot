package multipart

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
)

type Parser = func(p ParserParam) (*FileInfo, error)

type ParserParam struct {
	Header *multipart.FileHeader
	Data   io.ReadSeeker
}

type FileInfo struct {
	Name      string
	Size      int64
	Extension string
	Mimetype  string
}

func FileParser(p ParserParam) (*FileInfo, error) {
	if p.Header == nil {
		return nil, fmt.Errorf("invalid header")
	}
	if p.Data == nil {
		return nil, fmt.Errorf("invalid data")
	}

	buff := make([]byte, 512)
	n, err := p.Data.Read(buff)
	if err != nil && err != io.EOF {
		return nil, err
	}
	buff = buff[:n]

	_, err = p.Data.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	info := &FileInfo{
		Size:      p.Header.Size,
		Name:      FileName(p.Header),
		Extension: FileExtension(p.Header),
		Mimetype:  http.DetectContentType(buff),
	}
	return info, nil
}

func FileName(fh *multipart.FileHeader) string {
	names := strings.Split(fh.Filename, ".")
	if len(names) == 0 {
		return ""
	}
	return names[0]
}

func FileExtension(fh *multipart.FileHeader) string {
	names := strings.Split(fh.Filename, ".")
	if len(names) == 1 {
		return ""
	}
	return names[len(names)-1]
}
