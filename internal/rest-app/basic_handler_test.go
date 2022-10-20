package rest_app_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	rest_v1 "github.com/go-seidon/chariot/generated/rest-v1"
	rest_app "github.com/go-seidon/chariot/internal/rest-app"
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
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec = httptest.NewRecorder()

			ctx = e.NewContext(req, rec)
			basicHandler := rest_app.NewBasicHandler(rest_app.BasicHandlerParam{
				Config: &rest_app.RestAppConfig{
					AppName:    "name",
					AppVersion: "version",
				},
			})
			h = basicHandler.GetAppInfo
		})

		When("success get app info", func() {
			It("should return result", func() {

				err := h(ctx)

				res := &rest_v1.GetAppInfoResponse{}
				json.Unmarshal(rec.Body.Bytes(), res)

				Expect(err).To(BeNil())
				Expect(rec.Code).To(Equal(http.StatusOK))
				Expect(res.Code).To(Equal(int32(1000)))
				Expect(res.Message).To(Equal("success get app info"))
				Expect(res.Data).To(Equal(rest_v1.GetAppInfoData{
					AppName:    "name",
					AppVersion: "version",
				}))
			})
		})
	})
})
