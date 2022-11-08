package repository_mysql_test

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"

	// "regexp"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-seidon/chariot/internal/repository"
	repository_mysql "github.com/go-seidon/chariot/internal/repository-mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Auth Repository", func() {

	Context("CreateClient function", Label("unit"), func() {
		var (
			ctx        context.Context
			currentTs  time.Time
			dbClient   sqlmock.Sqlmock
			authRepo   repository.Auth
			p          repository.CreateClientParam
			checkStmt  string
			insertStmt string
			findStmt   string
		)

		BeforeEach(func() {
			var (
				db  *sql.DB
				err error
			)

			ctx = context.Background()
			currentTs = time.Now()
			db, dbClient, err = sqlmock.New()
			if err != nil {
				AbortSuite("failed create db mock: " + err.Error())
			}

			gormClient, err := gorm.Open(mysql.New(mysql.Config{
				Conn:                      db,
				SkipInitializeWithVersion: true,
			}), &gorm.Config{
				DisableAutomaticPing: true,
			})
			if err != nil {
				AbortSuite("failed create gorm client: " + err.Error())
			}
			authRepo = repository_mysql.NewAuth(repository_mysql.AuthParam{
				GormClient: gormClient,
			})

			p = repository.CreateClientParam{
				Id:           "id",
				ClientId:     "client-id",
				ClientSecret: "client-secret",
				Name:         "name",
				Type:         "basic",
				Status:       "active",
				CreatedAt:    currentTs,
			}
			checkStmt = regexp.QuoteMeta("SELECT id, client_id FROM `auth_client` WHERE client_id = ? ORDER BY `auth_client`.`id` LIMIT 1")
			insertStmt = regexp.QuoteMeta("INSERT INTO `auth_client` (`id`,`client_id`,`client_secret`,`name`,`type`,`status`,`created_at`) VALUES (?,?,?,?,?,?,?)")
			findStmt = regexp.QuoteMeta("SELECT id, client_id, client_secret, name, type, status, created_at FROM `auth_client` WHERE id = ? ORDER BY `auth_client`.`id` LIMIT 1")
		})

		AfterEach(func() {
			err := dbClient.ExpectationsWereMet()
			if err != nil {
				AbortSuite("some expectations were not met " + err.Error())
			}
		})

		When("failed begin trx", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin().
					WillReturnError(fmt.Errorf("begin error"))

				res, err := authRepo.CreateClient(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("begin error")))
			})
		})

		When("failed rollback during check client", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectQuery(checkStmt).
					WithArgs(p.ClientId).
					WillReturnError(fmt.Errorf("network error"))

				dbClient.
					ExpectRollback().
					WillReturnError(fmt.Errorf("rollback error"))

				res, err := authRepo.CreateClient(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("rollback error")))
			})
		})

		When("failed check client", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectQuery(checkStmt).
					WithArgs(p.ClientId).
					WillReturnError(fmt.Errorf("network error"))

				dbClient.
					ExpectRollback()

				res, err := authRepo.CreateClient(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("client already exists", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				rows := sqlmock.NewRows([]string{
					"id", "client_id",
				}).AddRow(
					p.Id, p.ClientId,
				)

				dbClient.
					ExpectQuery(checkStmt).
					WithArgs(p.ClientId).
					WillReturnRows(rows)

				dbClient.
					ExpectRollback()

				res, err := authRepo.CreateClient(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(repository.ErrExists))
			})
		})

		When("failed rollback during client creation", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectQuery(checkStmt).
					WithArgs(p.ClientId).
					WillReturnError(gorm.ErrRecordNotFound)

				dbClient.
					ExpectExec(insertStmt).
					WithArgs(
						p.Id, p.ClientId, p.ClientSecret,
						p.Name, p.Type, p.Status,
						p.CreatedAt.UnixMilli(),
					).
					WillReturnError(fmt.Errorf("network error"))

				dbClient.
					ExpectRollback().
					WillReturnError(fmt.Errorf("rollback error"))

				res, err := authRepo.CreateClient(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("rollback error")))
			})
		})

		When("failed create client", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectQuery(checkStmt).
					WithArgs(p.ClientId).
					WillReturnError(gorm.ErrRecordNotFound)

				dbClient.
					ExpectExec(insertStmt).
					WithArgs(
						p.Id, p.ClientId, p.ClientSecret,
						p.Name, p.Type, p.Status,
						p.CreatedAt.UnixMilli(),
					).
					WillReturnError(fmt.Errorf("network error"))

				dbClient.
					ExpectRollback()

				res, err := authRepo.CreateClient(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("failed rollback during find client", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectQuery(checkStmt).
					WithArgs(p.ClientId).
					WillReturnError(gorm.ErrRecordNotFound)

				dbClient.
					ExpectExec(insertStmt).
					WithArgs(
						p.Id, p.ClientId, p.ClientSecret,
						p.Name, p.Type, p.Status,
						p.CreatedAt.UnixMilli(),
					).
					WillReturnResult(sqlmock.NewResult(1, 1))

				dbClient.
					ExpectQuery(findStmt).
					WithArgs(p.Id).
					WillReturnError(fmt.Errorf("network error"))

				dbClient.
					ExpectRollback().
					WillReturnError(fmt.Errorf("rollback error"))

				res, err := authRepo.CreateClient(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("rollback error")))
			})
		})

		When("failed find client", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectQuery(checkStmt).
					WithArgs(p.ClientId).
					WillReturnError(gorm.ErrRecordNotFound)

				dbClient.
					ExpectExec(insertStmt).
					WithArgs(
						p.Id, p.ClientId, p.ClientSecret,
						p.Name, p.Type, p.Status,
						p.CreatedAt.UnixMilli(),
					).
					WillReturnResult(sqlmock.NewResult(1, 1))

				dbClient.
					ExpectQuery(findStmt).
					WithArgs(p.Id).
					WillReturnError(fmt.Errorf("network error"))

				dbClient.
					ExpectRollback()

				res, err := authRepo.CreateClient(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("failed commit during success create client", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectQuery(checkStmt).
					WithArgs(p.ClientId).
					WillReturnError(gorm.ErrRecordNotFound)

				dbClient.
					ExpectExec(insertStmt).
					WithArgs(
						p.Id, p.ClientId, p.ClientSecret,
						p.Name, p.Type, p.Status,
						p.CreatedAt.UnixMilli(),
					).
					WillReturnResult(sqlmock.NewResult(1, 1))

				rows := sqlmock.NewRows([]string{
					"id", "client_id", "client_secret",
					"name", "type", "status", "created_at",
				}).AddRow(
					p.Id, p.ClientId, p.ClientSecret,
					p.Name, p.Type, p.Status, p.CreatedAt.UnixMilli(),
				)
				dbClient.
					ExpectQuery(findStmt).
					WithArgs(p.Id).
					WillReturnRows(rows)

				dbClient.
					ExpectCommit().
					WillReturnError(fmt.Errorf("commit error"))

				res, err := authRepo.CreateClient(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("commit error")))
			})
		})

		When("success create client", func() {
			It("should return result", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectQuery(checkStmt).
					WithArgs(p.ClientId).
					WillReturnError(gorm.ErrRecordNotFound)

				dbClient.
					ExpectExec(insertStmt).
					WithArgs(
						p.Id, p.ClientId, p.ClientSecret,
						p.Name, p.Type, p.Status,
						p.CreatedAt.UnixMilli(),
					).
					WillReturnResult(sqlmock.NewResult(1, 1))

				rows := sqlmock.NewRows([]string{
					"id", "client_id", "client_secret",
					"name", "type", "status", "created_at",
				}).AddRow(
					p.Id, p.ClientId, p.ClientSecret,
					p.Name, p.Type, p.Status, p.CreatedAt.UnixMilli(),
				)
				dbClient.
					ExpectQuery(findStmt).
					WithArgs(p.Id).
					WillReturnRows(rows)

				dbClient.
					ExpectCommit()

				res, err := authRepo.CreateClient(ctx, p)

				expectedRes := &repository.CreateClientResult{
					Id:           p.Id,
					ClientId:     p.ClientId,
					ClientSecret: p.ClientSecret,
					Name:         p.Name,
					Type:         p.Type,
					Status:       p.Status,
					CreatedAt:    time.UnixMilli(p.CreatedAt.UnixMilli()).UTC(),
				}
				Expect(res).To(Equal(expectedRes))
				Expect(err).To(BeNil())
			})
		})
	})

	Context("FindClient function", Label("unit"), func() {

		var (
			ctx       context.Context
			currentTs time.Time
			dbClient  sqlmock.Sqlmock
			authRepo  repository.Auth
			p         repository.FindClientParam
			r         *repository.FindClientResult
			findStmt  string
		)

		BeforeEach(func() {
			var (
				db  *sql.DB
				err error
			)

			ctx = context.Background()
			currentTs = time.Now()
			db, dbClient, err = sqlmock.New()
			if err != nil {
				AbortSuite("failed create db mock: " + err.Error())
			}

			gormClient, err := gorm.Open(mysql.New(mysql.Config{
				Conn:                      db,
				SkipInitializeWithVersion: true,
			}), &gorm.Config{
				DisableAutomaticPing: true,
			})
			if err != nil {
				AbortSuite("failed create gorm client: " + err.Error())
			}
			authRepo = repository_mysql.NewAuth(repository_mysql.AuthParam{
				GormClient: gormClient,
			})

			p = repository.FindClientParam{
				Id: "id",
			}
			r = &repository.FindClientResult{
				Id:           "id",
				ClientId:     "client-id",
				ClientSecret: "client-secret",
				Name:         "name",
				Type:         "basic",
				Status:       "active",
				CreatedAt:    time.UnixMilli(currentTs.UnixMilli()).UTC(),
			}
			findStmt = regexp.QuoteMeta("SELECT id, client_id, client_secret, name, type, status, created_at, updated_at FROM `auth_client` WHERE id = ? ORDER BY `auth_client`.`id` LIMIT 1")
		})

		AfterEach(func() {
			err := dbClient.ExpectationsWereMet()
			if err != nil {
				AbortSuite("some expectations were not met " + err.Error())
			}
		})

		When("failed check client", func() {
			It("should return error", func() {
				dbClient.
					ExpectQuery(findStmt).
					WithArgs(p.Id).
					WillReturnError(fmt.Errorf("network error"))

				res, err := authRepo.FindClient(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("client is not available", func() {
			It("should return error", func() {
				dbClient.
					ExpectQuery(findStmt).
					WithArgs(p.Id).
					WillReturnError(gorm.ErrRecordNotFound)

				res, err := authRepo.FindClient(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(repository.ErrNotFound))
			})
		})

		When("success find client", func() {
			It("should return result", func() {
				rows := sqlmock.NewRows([]string{
					"id", "client_id", "client_secret",
					"name", "type", "status",
					"created_at", "updated_at",
				}).AddRow(
					r.Id, r.ClientId, r.ClientSecret,
					r.Name, r.Type, r.Status,
					currentTs.UnixMilli(), nil,
				)

				dbClient.
					ExpectQuery(findStmt).
					WithArgs(p.Id).
					WillReturnRows(rows)

				res, err := authRepo.FindClient(ctx, p)

				Expect(res).To(Equal(r))
				Expect(err).To(BeNil())
			})
		})

		When("success find updated client", func() {
			It("should return result", func() {
				rows := sqlmock.NewRows([]string{
					"id", "client_id", "client_secret",
					"name", "type", "status",
					"created_at", "updated_at",
				}).AddRow(
					r.Id, r.ClientId, r.ClientSecret,
					r.Name, r.Type, r.Status,
					currentTs.UnixMilli(), currentTs.UnixMilli(),
				)

				dbClient.
					ExpectQuery(findStmt).
					WithArgs(p.Id).
					WillReturnRows(rows)

				res, err := authRepo.FindClient(ctx, p)

				updatedAt := time.UnixMilli(currentTs.UnixMilli()).UTC()
				r := &repository.FindClientResult{
					Id:           "id",
					ClientId:     "client-id",
					ClientSecret: "client-secret",
					Name:         "name",
					Type:         "basic",
					Status:       "active",
					CreatedAt:    time.UnixMilli(currentTs.UnixMilli()).UTC(),
					UpdatedAt:    &updatedAt,
				}
				Expect(res).To(Equal(r))
				Expect(err).To(BeNil())
			})
		})
	})

})
