package multipart_test

import (
	"fmt"
	"io"
	mime_multipart "mime/multipart"

	"github.com/go-seidon/chariot/internal/storage/multipart"
	mock_io "github.com/go-seidon/provider/io/mock"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Parser Package", func() {

	Context("FileName function", Label("unit"), func() {
		var (
			fh *mime_multipart.FileHeader
		)

		BeforeEach(func() {
			fh = &mime_multipart.FileHeader{
				Filename: "dolpin.jpeg",
			}
		})

		When("file name is empty", func() {
			It("should return empty", func() {
				fh.Filename = ""
				res := multipart.FileName(fh)

				Expect(res).To(Equal(""))
			})
		})

		When("name contain extension", func() {
			It("should return only name", func() {
				res := multipart.FileName(fh)

				Expect(res).To(Equal("dolpin"))
			})
		})

		When("name not contain extension", func() {
			It("should return only name", func() {
				fh.Filename = "dolpin"
				res := multipart.FileName(fh)

				Expect(res).To(Equal("dolpin"))
			})
		})

		When("name contain multiple dot", func() {
			It("should return only name", func() {
				fh.Filename = "dolpin.new.jpeg"
				res := multipart.FileName(fh)

				Expect(res).To(Equal("dolpin"))
			})
		})
	})

	Context("FileExtension function", Label("unit"), func() {
		var (
			fh *mime_multipart.FileHeader
		)

		BeforeEach(func() {
			fh = &mime_multipart.FileHeader{
				Filename: "dolpin.jpeg",
			}
		})

		When("file name is empty", func() {
			It("should return empty", func() {
				fh.Filename = ""
				res := multipart.FileExtension(fh)

				Expect(res).To(Equal(""))
			})
		})

		When("name contain extension", func() {
			It("should return extension", func() {
				res := multipart.FileExtension(fh)

				Expect(res).To(Equal("jpeg"))
			})
		})

		When("name not contain extension", func() {
			It("should return empty", func() {
				fh.Filename = "dolpin"
				res := multipart.FileExtension(fh)

				Expect(res).To(Equal(""))
			})
		})

		When("name contain multiple dot", func() {
			It("should return only extension", func() {
				fh.Filename = "dolpin.new.jpeg"
				res := multipart.FileExtension(fh)

				Expect(res).To(Equal("jpeg"))
			})
		})
	})

	Context("FileParser function", Label("unit"), func() {
		var (
			f  *mock_io.MockReadSeeker
			fh *mime_multipart.FileHeader
		)

		BeforeEach(func() {
			t := GinkgoT()
			ctrl := gomock.NewController(t)
			f = mock_io.NewMockReadSeeker(ctrl)
			fh = &mime_multipart.FileHeader{
				Filename: "dolpin.jpeg",
				Size:     200,
			}
		})

		When("header is not specified", func() {
			It("should return error", func() {
				res, err := multipart.FileParser(multipart.ParserParam{
					Header: nil,
					Data:   f,
				})

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("invalid header")))
			})
		})

		When("data is not specified", func() {
			It("should return error", func() {
				res, err := multipart.FileParser(multipart.ParserParam{
					Header: fh,
					Data:   nil,
				})

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("invalid data")))
			})
		})

		When("failed read file", func() {
			It("should return error", func() {
				buff := make([]byte, 512)
				f.
					EXPECT().
					Read(gomock.Eq(buff)).
					Return(0, fmt.Errorf("disk error")).
					Times(1)

				res, err := multipart.FileParser(multipart.ParserParam{
					Header: fh,
					Data:   f,
				})

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("disk error")))
			})
		})

		When("failed seek to the end of file", func() {
			It("should return error", func() {
				buff := make([]byte, 512)
				f.
					EXPECT().
					Read(gomock.Eq(buff)).
					Return(512, nil).
					Times(1)

				f.
					EXPECT().
					Seek(gomock.Eq(int64(0)), gomock.Eq(0)).
					Return(int64(0), fmt.Errorf("disk error")).
					Times(1)

				res, err := multipart.FileParser(multipart.ParserParam{
					Header: fh,
					Data:   f,
				})

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("disk error")))
			})
		})

		When("reach end of file", func() {
			It("should return result", func() {
				buff := make([]byte, 512)
				f.
					EXPECT().
					Read(gomock.Eq(buff)).
					Return(200, io.EOF).
					Times(1)

				f.
					EXPECT().
					Seek(gomock.Eq(int64(0)), gomock.Eq(0)).
					Return(int64(1), nil).
					Times(1)

				res, err := multipart.FileParser(multipart.ParserParam{
					Header: fh,
					Data:   f,
				})

				expectedRes := &multipart.FileInfo{
					Name:      "dolpin",
					Extension: "jpeg",
					Size:      200,
					Mimetype:  "application/octet-stream",
				}
				Expect(res).To(Equal(expectedRes))
				Expect(err).To(BeNil())
			})
		})

		When("success parse file", func() {
			It("should return result", func() {
				buff := make([]byte, 512)
				f.
					EXPECT().
					Read(gomock.Eq(buff)).
					Return(512, nil).
					Times(1)

				f.
					EXPECT().
					Seek(gomock.Eq(int64(0)), gomock.Eq(0)).
					Return(int64(1), nil).
					Times(1)

				res, err := multipart.FileParser(multipart.ParserParam{
					Header: fh,
					Data:   f,
				})

				expectedRes := &multipart.FileInfo{
					Name:      "dolpin",
					Extension: "jpeg",
					Size:      200,
					Mimetype:  "application/octet-stream",
				}
				Expect(res).To(Equal(expectedRes))
				Expect(err).To(BeNil())
			})
		})

	})
})
