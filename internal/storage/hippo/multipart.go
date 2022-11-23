package hippo

import (
	"io"
	"mime/multipart"
)

type Writer = func(p WriterParam) (*multipart.Writer, error)

type WriterParam struct {
	Writer    io.Writer
	Reader    io.Reader
	FieldName string
	FileName  string
}

// @todo: add test if necessary
func FileWriter(p WriterParam) (*multipart.Writer, error) {
	writer := multipart.NewWriter(p.Writer)
	part, err := writer.CreateFormFile(p.FieldName, p.FileName)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(part, p.Reader)
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	return writer, nil
}
