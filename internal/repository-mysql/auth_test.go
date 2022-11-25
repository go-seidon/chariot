package repository_mysql_test

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-seidon/chariot/internal/repository"
	repository_mysql "github.com/go-seidon/chariot/internal/repository-mysql"
	"github.com/go-seidon/provider/typeconv"
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
			insertStmt = regexp.QuoteMeta("INSERT INTO `auth_client` (`id`,`client_id`,`client_secret`,`name`,`type`,`status`,`created_at`,`updated_at`) VALUES (?,?,?,?,?,?,?,?)")
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

	Context("UpdateClient function", Label("unit"), func() {

		var (
			ctx        context.Context
			currentTs  time.Time
			dbClient   sqlmock.Sqlmock
			authRepo   repository.Auth
			p          repository.UpdateClientParam
			r          *repository.UpdateClientResult
			findStmt   string
			updateStmt string
			checkStmt  string
			findRows   *sqlmock.Rows
			checkRows  *sqlmock.Rows
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

			p = repository.UpdateClientParam{
				Id:        "id",
				ClientId:  "new-client-id",
				Name:      "new-name",
				Type:      "basic",
				Status:    "active",
				UpdatedAt: currentTs,
			}
			r = &repository.UpdateClientResult{
				Id:           "id",
				ClientId:     "new-client-id",
				ClientSecret: "client-secret",
				Name:         "new-name",
				Type:         "basic",
				Status:       "active",
				CreatedAt:    time.UnixMilli(currentTs.UnixMilli()).UTC(),
				UpdatedAt:    time.UnixMilli(currentTs.UnixMilli()).UTC(),
			}
			findStmt = regexp.QuoteMeta("SELECT id, client_id, name, type, status FROM `auth_client` WHERE id = ? ORDER BY `auth_client`.`id` LIMIT 1")
			updateStmt = regexp.QuoteMeta("UPDATE `auth_client` SET `client_id`=?,`name`=?,`status`=?,`type`=?,`updated_at`=? WHERE id = ?")
			checkStmt = regexp.QuoteMeta("SELECT id, client_id, client_secret, name, type, status, created_at, updated_at FROM `auth_client` WHERE id = ? ORDER BY `auth_client`.`id` LIMIT 1")
			findRows = sqlmock.NewRows([]string{
				"id", "client_id",
				"name", "type", "status",
			}).AddRow(
				"id", "old-client-id",
				"old-name", "basic", "inactive",
			)
			checkRows = sqlmock.NewRows([]string{
				"id", "client_id", "client_secret",
				"name", "type", "status",
				"created_at", "updated_at",
			}).AddRow(
				"id", "new-client-id", "client-secret",
				"new-name", "basic", "active",
				currentTs.UnixMilli(), currentTs.UnixMilli(),
			)
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

				res, err := authRepo.UpdateClient(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("begin error")))
			})
		})

		When("failed rollback during find client", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectQuery(findStmt).
					WithArgs(p.Id).
					WillReturnError(fmt.Errorf("network error"))

				dbClient.
					ExpectRollback().
					WillReturnError(fmt.Errorf("rollback error"))

				res, err := authRepo.UpdateClient(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("rollback error")))
			})
		})

		When("failed find client", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectQuery(findStmt).
					WithArgs(p.Id).
					WillReturnError(fmt.Errorf("network error"))

				dbClient.
					ExpectRollback()

				res, err := authRepo.UpdateClient(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("client is not available", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectQuery(findStmt).
					WithArgs(p.Id).
					WillReturnError(gorm.ErrRecordNotFound)

				dbClient.
					ExpectRollback()

				res, err := authRepo.UpdateClient(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(repository.ErrNotFound))
			})
		})

		When("failed rollback during update client", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectQuery(findStmt).
					WithArgs(p.Id).
					WillReturnRows(findRows)

				dbClient.
					ExpectExec(updateStmt).
					WithArgs(
						p.ClientId,
						p.Name,
						p.Status,
						p.Type,
						p.UpdatedAt.UnixMilli(),
						p.Id,
					).
					WillReturnError(fmt.Errorf("network error"))

				dbClient.
					ExpectRollback().
					WillReturnError(fmt.Errorf("rollback error"))

				res, err := authRepo.UpdateClient(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("rollback error")))
			})
		})

		When("failed update client", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectQuery(findStmt).
					WithArgs(p.Id).
					WillReturnRows(findRows)

				dbClient.
					ExpectExec(updateStmt).
					WithArgs(
						p.ClientId,
						p.Name,
						p.Status,
						p.Type,
						p.UpdatedAt.UnixMilli(),
						p.Id,
					).
					WillReturnError(fmt.Errorf("network error"))

				dbClient.
					ExpectRollback()

				res, err := authRepo.UpdateClient(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("failed rollback during check update result", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectQuery(findStmt).
					WithArgs(p.Id).
					WillReturnRows(findRows)

				dbClient.
					ExpectExec(updateStmt).
					WithArgs(
						p.ClientId,
						p.Name,
						p.Status,
						p.Type,
						p.UpdatedAt.UnixMilli(),
						p.Id,
					).
					WillReturnResult(sqlmock.NewResult(1, 1))

				dbClient.
					ExpectQuery(checkStmt).
					WithArgs(p.Id).
					WillReturnError(fmt.Errorf("network error"))

				dbClient.
					ExpectRollback().
					WillReturnError(fmt.Errorf("rollback error"))

				res, err := authRepo.UpdateClient(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("rollback error")))
			})
		})

		When("failed check update result", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectQuery(findStmt).
					WithArgs(p.Id).
					WillReturnRows(findRows)

				dbClient.
					ExpectExec(updateStmt).
					WithArgs(
						p.ClientId,
						p.Name,
						p.Status,
						p.Type,
						p.UpdatedAt.UnixMilli(),
						p.Id,
					).
					WillReturnResult(sqlmock.NewResult(1, 1))

				dbClient.
					ExpectQuery(checkStmt).
					WithArgs(p.Id).
					WillReturnError(fmt.Errorf("network error"))

				dbClient.
					ExpectRollback()

				res, err := authRepo.UpdateClient(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("failed commit trx", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectQuery(findStmt).
					WithArgs(p.Id).
					WillReturnRows(findRows)

				dbClient.
					ExpectExec(updateStmt).
					WithArgs(
						p.ClientId,
						p.Name,
						p.Status,
						p.Type,
						p.UpdatedAt.UnixMilli(),
						p.Id,
					).
					WillReturnResult(sqlmock.NewResult(1, 1))

				dbClient.
					ExpectQuery(checkStmt).
					WithArgs(p.Id).
					WillReturnRows(checkRows)

				dbClient.
					ExpectCommit().
					WillReturnError(fmt.Errorf("commit error"))

				res, err := authRepo.UpdateClient(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("commit error")))
			})
		})

		When("success update client", func() {
			It("should return result", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectQuery(findStmt).
					WithArgs(p.Id).
					WillReturnRows(findRows)

				dbClient.
					ExpectExec(updateStmt).
					WithArgs(
						p.ClientId,
						p.Name,
						p.Status,
						p.Type,
						p.UpdatedAt.UnixMilli(),
						p.Id,
					).
					WillReturnResult(sqlmock.NewResult(1, 1))

				dbClient.
					ExpectQuery(checkStmt).
					WithArgs(p.Id).
					WillReturnRows(checkRows)

				dbClient.
					ExpectCommit()

				res, err := authRepo.UpdateClient(ctx, p)

				Expect(res).To(Equal(r))
				Expect(err).To(BeNil())
			})
		})
	})

	Context("SearchClient function", Label("unit"), func() {

		var (
			ctx        context.Context
			currentTs  time.Time
			dbClient   sqlmock.Sqlmock
			authRepo   repository.Auth
			p          repository.SearchClientParam
			r          *repository.SearchClientResult
			searchStmt string
			countStmt  string
			searchRows *sqlmock.Rows
			countRows  *sqlmock.Rows
		)

		BeforeEach(func() {
			var (
				db  *sql.DB
				err error
			)

			ctx = context.Background()
			currentTs = time.Now().UTC()
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

			p = repository.SearchClientParam{
				Limit:    24,
				Offset:   48,
				Keyword:  "goseidon",
				Statuses: []string{"active"},
			}
			updatedAt := time.UnixMilli(currentTs.UnixMilli()).UTC()
			r = &repository.SearchClientResult{
				Summary: repository.SearchClientSummary{
					TotalItems: 2,
				},
				Items: []repository.SearchClientItem{
					{
						Id:           "id-1",
						ClientId:     "client-id-1",
						ClientSecret: "client-secret-1",
						Name:         "name-1",
						Type:         "basic",
						Status:       "inactive",
						CreatedAt:    time.UnixMilli(currentTs.UnixMilli()).UTC(),
						UpdatedAt:    &updatedAt,
					},
					{
						Id:           "id-2",
						ClientId:     "client-id-2",
						ClientSecret: "client-secret-2",
						Name:         "name-2",
						Type:         "basic",
						Status:       "active",
						CreatedAt:    time.UnixMilli(currentTs.UnixMilli()).UTC(),
						UpdatedAt:    &updatedAt,
					},
				},
			}
			searchStmt = regexp.QuoteMeta(strings.TrimSpace(`
				SELECT id, client_id, client_secret, name, type, status, created_at, updated_at 
				FROM ` + "`auth_client`" + ` 
				WHERE name LIKE ? OR client_id LIKE ?
				AND status IN (?)
				LIMIT 24
				OFFSET 48
			`))
			countStmt = regexp.QuoteMeta(strings.TrimSpace(`
				SELECT count(*)
				FROM ` + "`auth_client`" + ` 
				WHERE name LIKE ? OR client_id LIKE ?
				AND status IN (?)
			`))
			searchRows = sqlmock.NewRows([]string{
				"id", "client_id", "client_secret",
				"name", "type", "status",
				"created_at", "updated_at",
			}).AddRow(
				"id-1", "client-id-1", "client-secret-1",
				"name-1", "basic", "inactive",
				currentTs.UnixMilli(), currentTs.UnixMilli(),
			).AddRow(
				"id-2", "client-id-2", "client-secret-2",
				"name-2", "basic", "active",
				currentTs.UnixMilli(), currentTs.UnixMilli(),
			)
			countRows = sqlmock.
				NewRows([]string{"count(*)"}).
				AddRow(2)
		})

		AfterEach(func() {
			err := dbClient.ExpectationsWereMet()
			if err != nil {
				AbortSuite("some expectations were not met " + err.Error())
			}
		})

		When("failed search client", func() {
			It("should return error", func() {
				dbClient.
					ExpectQuery(searchStmt).
					WithArgs(
						"%"+p.Keyword+"%",
						"%"+p.Keyword+"%",
						p.Statuses[0],
					).
					WillReturnError(fmt.Errorf("network error"))

				res, err := authRepo.SearchClient(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("there are no client", func() {
			It("should return empty result", func() {
				dbClient.
					ExpectQuery(searchStmt).
					WithArgs(
						"%"+p.Keyword+"%",
						"%"+p.Keyword+"%",
						p.Statuses[0],
					).
					WillReturnError(gorm.ErrRecordNotFound)

				res, err := authRepo.SearchClient(ctx, p)

				r := &repository.SearchClientResult{
					Summary: repository.SearchClientSummary{
						TotalItems: 0,
					},
					Items: []repository.SearchClientItem{},
				}
				Expect(res).To(Equal(r))
				Expect(err).To(BeNil())
			})
		})

		When("failed count search client", func() {
			It("should return result", func() {
				searchRows := sqlmock.NewRows([]string{
					"id", "client_id", "client_secret",
					"name", "type", "status",
					"created_at", "updated_at",
				}).AddRow(
					"id-1", "client-id-1", "client-secret-1",
					"name-1", "basic", "inactive",
					currentTs.UnixMilli(), nil,
				)
				dbClient.
					ExpectQuery(searchStmt).
					WithArgs(
						"%"+p.Keyword+"%",
						"%"+p.Keyword+"%",
						p.Statuses[0],
					).
					WillReturnRows(searchRows)

				dbClient.
					ExpectQuery(countStmt).
					WithArgs(
						"%"+p.Keyword+"%",
						"%"+p.Keyword+"%",
						p.Statuses[0],
					).
					WillReturnError(fmt.Errorf("network error"))

				res, err := authRepo.SearchClient(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("there is one client", func() {
			It("should return result", func() {
				searchRows := sqlmock.NewRows([]string{
					"id", "client_id", "client_secret",
					"name", "type", "status",
					"created_at", "updated_at",
				}).AddRow(
					"id-1", "client-id-1", "client-secret-1",
					"name-1", "basic", "inactive",
					currentTs.UnixMilli(), currentTs.UnixMilli(),
				)
				dbClient.
					ExpectQuery(searchStmt).
					WithArgs(
						"%"+p.Keyword+"%",
						"%"+p.Keyword+"%",
						p.Statuses[0],
					).
					WillReturnRows(searchRows)

				countRows = sqlmock.
					NewRows([]string{"count(*)"}).
					AddRow(1)
				dbClient.
					ExpectQuery(countStmt).
					WithArgs(
						"%"+p.Keyword+"%",
						"%"+p.Keyword+"%",
						p.Statuses[0],
					).
					WillReturnRows(countRows)

				res, err := authRepo.SearchClient(ctx, p)

				r := &repository.SearchClientResult{
					Summary: repository.SearchClientSummary{
						TotalItems: 1,
					},
					Items: []repository.SearchClientItem{
						{
							Id:           "id-1",
							ClientId:     "client-id-1",
							ClientSecret: "client-secret-1",
							Name:         "name-1",
							Type:         "basic",
							Status:       "inactive",
							CreatedAt:    time.UnixMilli(currentTs.UnixMilli()).UTC(),
							UpdatedAt:    typeconv.Time(time.UnixMilli(currentTs.UnixMilli()).UTC()),
						},
					},
				}
				Expect(res).To(Equal(r))
				Expect(err).To(BeNil())
			})
		})

		When("there are some clients", func() {
			It("should return result", func() {
				dbClient.
					ExpectQuery(searchStmt).
					WithArgs(
						"%"+p.Keyword+"%",
						"%"+p.Keyword+"%",
						p.Statuses[0],
					).
					WillReturnRows(searchRows)

				dbClient.
					ExpectQuery(countStmt).
					WithArgs(
						"%"+p.Keyword+"%",
						"%"+p.Keyword+"%",
						p.Statuses[0],
					).
					WillReturnRows(countRows)

				res, err := authRepo.SearchClient(ctx, p)

				Expect(res).To(Equal(r))
				Expect(err).To(BeNil())
			})
		})
	})

})
