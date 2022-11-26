package resthandler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/go-seidon/chariot/generated/restapp"
	"github.com/go-seidon/chariot/internal/resthandler"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Basic Handler", func() {
	Context("GetAppInfo function", Label("unit"), func() {
		var (
			ctx echo.Context
			h   func(ctx echo.Context) error
			rec *httptest.ResponseRecorder
		)

		BeforeEach(func() {
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec = httptest.NewRecorder()

			e := echo.New()
			ctx = e.NewContext(req, rec)
			basicHandler := resthandler.NewBasic(resthandler.BasicParam{
				Config: &resthandler.BasicConfig{
					AppName:    "name",
					AppVersion: "version",
				},
			})
			h = basicHandler.GetAppInfo
		})

		When("success get app info", func() {
			It("should return result", func() {

				err := h(ctx)

				res := &restapp.GetAppInfoResponse{}
				json.Unmarshal(rec.Body.Bytes(), res)

				Expect(err).To(BeNil())
				Expect(rec.Code).To(Equal(http.StatusOK))
				Expect(res.Code).To(Equal(int32(1000)))
				Expect(res.Message).To(Equal("success get app info"))
				Expect(res.Data).To(Equal(restapp.GetAppInfoData{
					AppName:    "name",
					AppVersion: "version",
				}))
			})
		})
	})
})
