package mysql_test

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/go-seidon/chariot/internal/repository"
	mysql "github.com/go-seidon/chariot/internal/repository/mysql"
	mock_mysql "github.com/go-seidon/provider/mysql/mock"
)

var _ = Describe("Repository Provider", func() {
	Context("NewRepository function", Label("unit"), func() {
		When("db client is not specified", func() {
			It("should return error", func() {
				res, err := mysql.NewRepository()

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("invalid db client")))
			})
		})

		When("required parameters are specified", func() {
			It("should return result", func() {
				mOpt := mysql.WithDbClient(&sql.DB{})
				res, err := mysql.NewRepository(mOpt)

				Expect(res).ToNot(BeNil())
				Expect(err).To(BeNil())
			})
		})
	})

	Context("GetAuth function", Label("unit"), func() {
		var (
			provider repository.Repository
		)

		BeforeEach(func() {
			mOpt := mysql.WithDbClient(&sql.DB{})
			provider, _ = mysql.NewRepository(mOpt)
		})

		When("function is called", func() {
			It("should return result", func() {
				res := provider.GetAuth()

				Expect(res).ToNot(BeNil())
			})
		})
	})

	Context("GetBarrel function", Label("unit"), func() {
		var (
			provider repository.Repository
		)

		BeforeEach(func() {
			mOpt := mysql.WithDbClient(&sql.DB{})
			provider, _ = mysql.NewRepository(mOpt)
		})

		When("function is called", func() {
			It("should return result", func() {
				res := provider.GetBarrel()

				Expect(res).ToNot(BeNil())
			})
		})
	})

	Context("GetFile function", Label("unit"), func() {
		var (
			provider repository.Repository
		)

		BeforeEach(func() {
			mOpt := mysql.WithDbClient(&sql.DB{})
			provider, _ = mysql.NewRepository(mOpt)
		})

		When("function is called", func() {
			It("should return result", func() {
				res := provider.GetFile()

				Expect(res).ToNot(BeNil())
			})
		})
	})

	Context("Init function", Label("unit"), func() {
		var (
			provider repository.Repository
			ctx      context.Context
		)

		BeforeEach(func() {
			mOpt := mysql.WithDbClient(&sql.DB{})
			provider, _ = mysql.NewRepository(mOpt)
			ctx = context.Background()
		})

		When("success init", func() {
			It("should return result", func() {
				res := provider.Init(ctx)

				Expect(res).To(BeNil())
			})
		})
	})

	Context("Ping function", Label("unit"), func() {
		var (
			provider repository.Repository
			ctx      context.Context
			client   *mock_mysql.MockClient
		)

		BeforeEach(func() {
			ctx = context.Background()
			t := GinkgoT()
			ctrl := gomock.NewController(t)
			client = mock_mysql.NewMockClient(ctrl)
			provider, _ = mysql.NewRepository(
				mysql.WithDbClient(client),
			)
		})

		When("success ping", func() {
			It("should return result", func() {
				client.
					EXPECT().
					PingContext(gomock.Eq(ctx)).
					Return(nil).
					Times(1)

				err := provider.Ping(ctx)

				Expect(err).To(BeNil())
			})
		})

		When("failed ping", func() {
			It("should return error", func() {
				client.
					EXPECT().
					PingContext(gomock.Eq(ctx)).
					Return(fmt.Errorf("ping error")).
					Times(1)

				err := provider.Ping(ctx)

				Expect(err).To(Equal(fmt.Errorf("ping error")))
			})
		})
	})
})
