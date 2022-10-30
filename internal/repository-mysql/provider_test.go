package repository_mysql_test

// import (
// 	"context"
// 	"database/sql"
// 	"fmt"

// 	"github.com/golang/mock/gomock"
// 	. "github.com/onsi/ginkgo/v2"
// 	. "github.com/onsi/gomega"

// 	mock_db_mysql "github.com/go-seidon/chariot/internal/db-mysql/mock"
// 	"github.com/go-seidon/chariot/internal/repository"
// 	repository_mysql "github.com/go-seidon/chariot/internal/repository-mysql"
// )

// var _ = Describe("Repository Provider", func() {
// 	Context("NewRepository function", Label("unit"), func() {
// 		When("master db client is not specified", func() {
// 			It("should return error", func() {
// 				res, err := repository_mysql.NewRepository()

// 				Expect(res).To(BeNil())
// 				Expect(err).To(Equal(fmt.Errorf("invalid db client specified")))
// 			})
// 		})

// 		When("replica db client is not specified", func() {
// 			It("should return error", func() {
// 				mOpt := repository_mysql.WithDbMaster(&sql.DB{})
// 				res, err := repository_mysql.NewRepository(mOpt)

// 				Expect(res).To(BeNil())
// 				Expect(err).To(Equal(fmt.Errorf("invalid db client specified")))
// 			})
// 		})

// 		When("required parameters are specified", func() {
// 			It("should return result", func() {
// 				mOpt := repository_mysql.WithDbMaster(&sql.DB{})
// 				rOpt := repository_mysql.WithDbReplica(&sql.DB{})
// 				res, err := repository_mysql.NewRepository(mOpt, rOpt)

// 				Expect(res).ToNot(BeNil())
// 				Expect(err).To(BeNil())
// 			})
// 		})
// 	})

// 	Context("GetAuthRepo function", Label("unit"), func() {
// 		var (
// 			provider repository.Provider
// 		)

// 		BeforeEach(func() {
// 			mOpt := repository_mysql.WithDbMaster(&sql.DB{})
// 			rOpt := repository_mysql.WithDbReplica(&sql.DB{})
// 			provider, _ = repository_mysql.NewRepository(mOpt, rOpt)
// 		})

// 		When("function is called", func() {
// 			It("should return result", func() {
// 				res := provider.GetAuthRepo()

// 				Expect(res).ToNot(BeNil())
// 			})
// 		})
// 	})

// 	Context("Init function", Label("unit"), func() {
// 		var (
// 			provider repository.Provider
// 			ctx      context.Context
// 		)

// 		BeforeEach(func() {
// 			mOpt := repository_mysql.WithDbMaster(&sql.DB{})
// 			rOpt := repository_mysql.WithDbReplica(&sql.DB{})
// 			provider, _ = repository_mysql.NewRepository(mOpt, rOpt)
// 			ctx = context.Background()
// 		})

// 		When("success init", func() {
// 			It("should return result", func() {
// 				res := provider.Init(ctx)

// 				Expect(res).To(BeNil())
// 			})
// 		})
// 	})

// 	Context("Ping function", Label("unit"), func() {
// 		var (
// 			provider repository.Provider
// 			ctx      context.Context
// 			client   *mock_db_mysql.MockClient
// 		)

// 		BeforeEach(func() {
// 			t := GinkgoT()
// 			ctrl := gomock.NewController(t)
// 			client = mock_db_mysql.NewMockClient(ctrl)
// 			mOpt := repository_mysql.WithDbMaster(client)
// 			rOpt := repository_mysql.WithDbReplica(client)
// 			provider, _ = repository_mysql.NewRepository(mOpt, rOpt)
// 			ctx = context.Background()
// 		})

// 		When("success ping", func() {
// 			It("should return result", func() {
// 				client.
// 					EXPECT().
// 					PingContext(gomock.Eq(ctx)).
// 					Return(nil).
// 					Times(1)

// 				res := provider.Ping(ctx)

// 				Expect(res).To(BeNil())
// 			})
// 		})

// 		When("failed ping", func() {
// 			It("should return error", func() {
// 				client.
// 					EXPECT().
// 					PingContext(gomock.Eq(ctx)).
// 					Return(fmt.Errorf("ping error")).
// 					Times(1)

// 				res := provider.Ping(ctx)

// 				Expect(res).To(Equal(fmt.Errorf("ping error")))
// 			})
// 		})
// 	})
// })
