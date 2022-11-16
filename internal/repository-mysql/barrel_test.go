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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var _ = Describe("Barrel Repository", func() {
	Context("CreateBarrel function", Label("unit"), func() {
		var (
			ctx        context.Context
			currentTs  time.Time
			dbClient   sqlmock.Sqlmock
			barrelRepo repository.Barrel
			p          repository.CreateBarrelParam
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
			barrelRepo = repository_mysql.NewBarrel(repository_mysql.BarrelParam{
				GormClient: gormClient,
			})

			p = repository.CreateBarrelParam{
				Id:        "id",
				Code:      "code",
				Name:      "name",
				Provider:  "goseidon_hippo",
				Status:    "active",
				CreatedAt: currentTs,
			}
			checkStmt = regexp.QuoteMeta("SELECT id, code FROM `barrel` WHERE code = ? ORDER BY `barrel`.`id` LIMIT 1")
			insertStmt = regexp.QuoteMeta("INSERT INTO `barrel` (`id`,`code`,`name`,`provider`,`status`,`created_at`) VALUES (?,?,?,?,?,?)")
			findStmt = regexp.QuoteMeta("SELECT id, code, name, provider, status, created_at FROM `barrel` WHERE id = ? ORDER BY `barrel`.`id` LIMIT 1")
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

				res, err := barrelRepo.CreateBarrel(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("begin error")))
			})
		})

		When("failed rollback during check barrel", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectQuery(checkStmt).
					WithArgs(p.Code).
					WillReturnError(fmt.Errorf("network error"))

				dbClient.
					ExpectRollback().
					WillReturnError(fmt.Errorf("rollback error"))

				res, err := barrelRepo.CreateBarrel(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("rollback error")))
			})
		})

		When("failed check barrel", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectQuery(checkStmt).
					WithArgs(p.Code).
					WillReturnError(fmt.Errorf("network error"))

				dbClient.
					ExpectRollback()

				res, err := barrelRepo.CreateBarrel(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("barrel already exists", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				rows := sqlmock.NewRows([]string{
					"id", "code",
				}).AddRow(
					p.Id, p.Code,
				)

				dbClient.
					ExpectQuery(checkStmt).
					WithArgs(p.Code).
					WillReturnRows(rows)

				dbClient.
					ExpectRollback()

				res, err := barrelRepo.CreateBarrel(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(repository.ErrExists))
			})
		})

		When("failed rollback during barrel creation", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectQuery(checkStmt).
					WithArgs(p.Code).
					WillReturnError(gorm.ErrRecordNotFound)

				dbClient.
					ExpectExec(insertStmt).
					WithArgs(
						p.Id, p.Code,
						p.Name, p.Provider, p.Status,
						p.CreatedAt.UnixMilli(),
					).
					WillReturnError(fmt.Errorf("network error"))

				dbClient.
					ExpectRollback().
					WillReturnError(fmt.Errorf("rollback error"))

				res, err := barrelRepo.CreateBarrel(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("rollback error")))
			})
		})

		When("failed create barrel", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectQuery(checkStmt).
					WithArgs(p.Code).
					WillReturnError(gorm.ErrRecordNotFound)

				dbClient.
					ExpectExec(insertStmt).
					WithArgs(
						p.Id, p.Code,
						p.Name, p.Provider, p.Status,
						p.CreatedAt.UnixMilli(),
					).
					WillReturnError(fmt.Errorf("network error"))

				dbClient.
					ExpectRollback()

				res, err := barrelRepo.CreateBarrel(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("failed rollback during find barrel", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectQuery(checkStmt).
					WithArgs(p.Code).
					WillReturnError(gorm.ErrRecordNotFound)

				dbClient.
					ExpectExec(insertStmt).
					WithArgs(
						p.Id, p.Code,
						p.Name, p.Provider, p.Status,
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

				res, err := barrelRepo.CreateBarrel(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("rollback error")))
			})
		})

		When("failed find barrel", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectQuery(checkStmt).
					WithArgs(p.Code).
					WillReturnError(gorm.ErrRecordNotFound)

				dbClient.
					ExpectExec(insertStmt).
					WithArgs(
						p.Id, p.Code,
						p.Name, p.Provider, p.Status,
						p.CreatedAt.UnixMilli(),
					).
					WillReturnResult(sqlmock.NewResult(1, 1))

				dbClient.
					ExpectQuery(findStmt).
					WithArgs(p.Id).
					WillReturnError(fmt.Errorf("network error"))

				dbClient.
					ExpectRollback()

				res, err := barrelRepo.CreateBarrel(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("failed commit during success create barrel", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectQuery(checkStmt).
					WithArgs(p.Code).
					WillReturnError(gorm.ErrRecordNotFound)

				dbClient.
					ExpectExec(insertStmt).
					WithArgs(
						p.Id, p.Code,
						p.Name, p.Provider, p.Status,
						p.CreatedAt.UnixMilli(),
					).
					WillReturnResult(sqlmock.NewResult(1, 1))

				rows := sqlmock.NewRows([]string{
					"id", "code",
					"name", "provider",
					"status", "created_at",
				}).AddRow(
					p.Id, p.Code,
					p.Name, p.Provider,
					p.Status, p.CreatedAt.UnixMilli(),
				)
				dbClient.
					ExpectQuery(findStmt).
					WithArgs(p.Id).
					WillReturnRows(rows)

				dbClient.
					ExpectCommit().
					WillReturnError(fmt.Errorf("commit error"))

				res, err := barrelRepo.CreateBarrel(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("commit error")))
			})
		})

		When("success create barrel", func() {
			It("should return result", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectQuery(checkStmt).
					WithArgs(p.Code).
					WillReturnError(gorm.ErrRecordNotFound)

				dbClient.
					ExpectExec(insertStmt).
					WithArgs(
						p.Id, p.Code,
						p.Name, p.Provider, p.Status,
						p.CreatedAt.UnixMilli(),
					).
					WillReturnResult(sqlmock.NewResult(1, 1))

				rows := sqlmock.NewRows([]string{
					"id", "code",
					"name", "provider",
					"status", "created_at",
				}).AddRow(
					p.Id, p.Code,
					p.Name, p.Provider,
					p.Status, p.CreatedAt.UnixMilli(),
				)
				dbClient.
					ExpectQuery(findStmt).
					WithArgs(p.Id).
					WillReturnRows(rows)

				dbClient.
					ExpectCommit()

				res, err := barrelRepo.CreateBarrel(ctx, p)

				expectedRes := &repository.CreateBarrelResult{
					Id:        p.Id,
					Code:      p.Code,
					Name:      p.Name,
					Provider:  p.Provider,
					Status:    p.Status,
					CreatedAt: time.UnixMilli(p.CreatedAt.UnixMilli()).UTC(),
				}
				Expect(res).To(Equal(expectedRes))
				Expect(err).To(BeNil())
			})
		})
	})

	Context("FindBarrel function", Label("unit"), func() {

		var (
			ctx        context.Context
			currentTs  time.Time
			dbClient   sqlmock.Sqlmock
			barrelRepo repository.Barrel
			p          repository.FindBarrelParam
			r          *repository.FindBarrelResult
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
			barrelRepo = repository_mysql.NewBarrel(repository_mysql.BarrelParam{
				GormClient: gormClient,
			})

			p = repository.FindBarrelParam{
				Id: "id",
			}
			r = &repository.FindBarrelResult{
				Id:        "id",
				Code:      "code",
				Name:      "name",
				Provider:  "goseidon_hippo",
				Status:    "active",
				CreatedAt: time.UnixMilli(currentTs.UnixMilli()).UTC(),
			}
			findStmt = regexp.QuoteMeta("SELECT id, code, name, provider, status, created_at, updated_at FROM `barrel` WHERE id = ? ORDER BY `barrel`.`id` LIMIT 1")
		})

		AfterEach(func() {
			err := dbClient.ExpectationsWereMet()
			if err != nil {
				AbortSuite("some expectations were not met " + err.Error())
			}
		})

		When("failed check barrel", func() {
			It("should return error", func() {
				dbClient.
					ExpectQuery(findStmt).
					WithArgs(p.Id).
					WillReturnError(fmt.Errorf("network error"))

				res, err := barrelRepo.FindBarrel(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("barrel is not available", func() {
			It("should return error", func() {
				dbClient.
					ExpectQuery(findStmt).
					WithArgs(p.Id).
					WillReturnError(gorm.ErrRecordNotFound)

				res, err := barrelRepo.FindBarrel(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(repository.ErrNotFound))
			})
		})

		When("success find barrel", func() {
			It("should return result", func() {
				rows := sqlmock.NewRows([]string{
					"id", "code",
					"name", "provider", "status",
					"created_at", "updated_at",
				}).AddRow(
					r.Id, r.Code,
					r.Name, r.Provider, r.Status,
					currentTs.UnixMilli(), nil,
				)

				dbClient.
					ExpectQuery(findStmt).
					WithArgs(p.Id).
					WillReturnRows(rows)

				res, err := barrelRepo.FindBarrel(ctx, p)

				Expect(res).To(Equal(r))
				Expect(err).To(BeNil())
			})
		})

		When("success find updated barrel", func() {
			It("should return result", func() {
				rows := sqlmock.NewRows([]string{
					"id", "code",
					"name", "provider", "status",
					"created_at", "updated_at",
				}).AddRow(
					r.Id, r.Code,
					r.Name, r.Provider, r.Status,
					currentTs.UnixMilli(),
					currentTs.UnixMilli(),
				)

				dbClient.
					ExpectQuery(findStmt).
					WithArgs(p.Id).
					WillReturnRows(rows)

				res, err := barrelRepo.FindBarrel(ctx, p)

				updatedAt := time.UnixMilli(currentTs.UnixMilli()).UTC()
				r := &repository.FindBarrelResult{
					Id:        "id",
					Code:      "code",
					Name:      "name",
					Provider:  "goseidon_hippo",
					Status:    "active",
					CreatedAt: time.UnixMilli(currentTs.UnixMilli()).UTC(),
					UpdatedAt: &updatedAt,
				}
				Expect(res).To(Equal(r))
				Expect(err).To(BeNil())
			})
		})
	})

	Context("UpdateBarrel function", Label("unit"), func() {

		var (
			ctx        context.Context
			currentTs  time.Time
			dbClient   sqlmock.Sqlmock
			barrelRepo repository.Barrel
			p          repository.UpdateBarrelParam
			r          *repository.UpdateBarrelResult
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
			barrelRepo = repository_mysql.NewBarrel(repository_mysql.BarrelParam{
				GormClient: gormClient,
			})

			p = repository.UpdateBarrelParam{
				Id:        "id",
				Code:      "new-code",
				Name:      "new-name",
				Provider:  "goseidon_hippo",
				Status:    "active",
				UpdatedAt: currentTs,
			}
			r = &repository.UpdateBarrelResult{
				Id:        "id",
				Code:      "new-code",
				Name:      "new-name",
				Provider:  "goseidon_hippo",
				Status:    "active",
				CreatedAt: time.UnixMilli(currentTs.UnixMilli()).UTC(),
				UpdatedAt: time.UnixMilli(currentTs.UnixMilli()).UTC(),
			}
			findStmt = regexp.QuoteMeta("SELECT id, code, name, provider, status FROM `barrel` WHERE id = ? ORDER BY `barrel`.`id` LIMIT 1")
			updateStmt = regexp.QuoteMeta("UPDATE `barrel` SET `code`=?,`name`=?,`provider`=?,`status`=?,`updated_at`=? WHERE id = ?")
			checkStmt = regexp.QuoteMeta("SELECT id, code, name, provider, status, created_at, updated_at FROM `barrel` WHERE id = ? ORDER BY `barrel`.`id` LIMIT 1")
			findRows = sqlmock.NewRows([]string{
				"id", "code",
				"name", "provider", "status",
			}).AddRow(
				"id", "old-code",
				"old-name", "goseidon_hippo", "inactive",
			)
			checkRows = sqlmock.NewRows([]string{
				"id", "code",
				"name", "provider",
				"status",
				"created_at", "updated_at",
			}).AddRow(
				"id", "new-code",
				"new-name", "goseidon_hippo", "active",
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

				res, err := barrelRepo.UpdateBarrel(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("begin error")))
			})
		})

		When("failed rollback during find barrel", func() {
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

				res, err := barrelRepo.UpdateBarrel(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("rollback error")))
			})
		})

		When("failed find barrel", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectQuery(findStmt).
					WithArgs(p.Id).
					WillReturnError(fmt.Errorf("network error"))

				dbClient.
					ExpectRollback()

				res, err := barrelRepo.UpdateBarrel(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("barrel is not available", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectQuery(findStmt).
					WithArgs(p.Id).
					WillReturnError(gorm.ErrRecordNotFound)

				dbClient.
					ExpectRollback()

				res, err := barrelRepo.UpdateBarrel(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(repository.ErrNotFound))
			})
		})

		When("failed rollback during update barrel", func() {
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
						p.Code,
						p.Name,
						p.Provider,
						p.Status,
						p.UpdatedAt.UnixMilli(),
						p.Id,
					).
					WillReturnError(fmt.Errorf("network error"))

				dbClient.
					ExpectRollback().
					WillReturnError(fmt.Errorf("rollback error"))

				res, err := barrelRepo.UpdateBarrel(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("rollback error")))
			})
		})

		When("failed update barrel", func() {
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
						p.Code,
						p.Name,
						p.Provider,
						p.Status,
						p.UpdatedAt.UnixMilli(),
						p.Id,
					).
					WillReturnError(fmt.Errorf("network error"))

				dbClient.
					ExpectRollback()

				res, err := barrelRepo.UpdateBarrel(ctx, p)

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
						p.Code,
						p.Name,
						p.Provider,
						p.Status,
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

				res, err := barrelRepo.UpdateBarrel(ctx, p)

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
						p.Code,
						p.Name,
						p.Provider,
						p.Status,
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

				res, err := barrelRepo.UpdateBarrel(ctx, p)

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
						p.Code,
						p.Name,
						p.Provider,
						p.Status,
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

				res, err := barrelRepo.UpdateBarrel(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("commit error")))
			})
		})

		When("success update barrel", func() {
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
						p.Code,
						p.Name,
						p.Provider,
						p.Status,
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

				res, err := barrelRepo.UpdateBarrel(ctx, p)

				Expect(res).To(Equal(r))
				Expect(err).To(BeNil())
			})
		})
	})

	Context("SearchBarrel function", Label("unit"), func() {

		var (
			ctx        context.Context
			currentTs  time.Time
			dbClient   sqlmock.Sqlmock
			barrelRepo repository.Barrel
			p          repository.SearchBarrelParam
			r          *repository.SearchBarrelResult
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
			barrelRepo = repository_mysql.NewBarrel(repository_mysql.BarrelParam{
				GormClient: gormClient,
			})

			p = repository.SearchBarrelParam{
				Limit:     24,
				Offset:    48,
				Keyword:   "goseidon",
				Statuses:  []string{"active"},
				Providers: []string{"goseidon_hippo"},
				Codes:     []string{"hippo1"},
			}
			updatedAt := time.UnixMilli(currentTs.UnixMilli()).UTC()
			r = &repository.SearchBarrelResult{
				Summary: repository.SearchBarrelSummary{
					TotalItems: 2,
				},
				Items: []repository.SearchBarrelItem{
					{
						Id:        "id-1",
						Code:      "code-1",
						Name:      "name-1",
						Provider:  "goseidon_hippo",
						Status:    "inactive",
						CreatedAt: time.UnixMilli(currentTs.UnixMilli()).UTC(),
					},
					{
						Id:        "id-2",
						Code:      "code-2",
						Name:      "name-2",
						Provider:  "goseidon_hippo",
						Status:    "active",
						CreatedAt: time.UnixMilli(currentTs.UnixMilli()).UTC(),
						UpdatedAt: &updatedAt,
					},
				},
			}
			searchStmt = regexp.QuoteMeta(strings.TrimSpace(`
				SELECT id, code, name, provider, status, created_at, updated_at 
				FROM ` + "`barrel`" + ` 
				WHERE name LIKE ? OR code LIKE ?
				AND status IN (?)
				AND provider IN (?)
				AND code IN (?)
				LIMIT 24
				OFFSET 48
			`))
			countStmt = regexp.QuoteMeta(strings.TrimSpace(`
				SELECT count(*)
				FROM ` + "`barrel`" + ` 
				WHERE name LIKE ? OR code LIKE ?
				AND status IN (?)
				AND provider IN (?)
				AND code IN (?)
			`))
			searchRows = sqlmock.NewRows([]string{
				"id", "code",
				"name", "provider", "status",
				"created_at", "updated_at",
			}).AddRow(
				"id-1", "code-1",
				"name-1", "goseidon_hippo", "inactive",
				currentTs.UnixMilli(), nil,
			).AddRow(
				"id-2", "code-2",
				"name-2", "goseidon_hippo", "active",
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
						p.Providers[0],
						p.Codes[0],
					).
					WillReturnError(fmt.Errorf("network error"))

				res, err := barrelRepo.SearchBarrel(ctx, p)

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
						p.Providers[0],
						p.Codes[0],
					).
					WillReturnError(gorm.ErrRecordNotFound)

				res, err := barrelRepo.SearchBarrel(ctx, p)

				r := &repository.SearchBarrelResult{
					Summary: repository.SearchBarrelSummary{
						TotalItems: 0,
					},
					Items: []repository.SearchBarrelItem{},
				}
				Expect(res).To(Equal(r))
				Expect(err).To(BeNil())
			})
		})

		When("failed count search client", func() {
			It("should return result", func() {
				searchRows := sqlmock.NewRows([]string{
					"id", "code",
					"name", "provider", "status",
					"created_at", "updated_at",
				}).AddRow(
					"id-1", "code-1",
					"name-1", "goseidon_hippo", "inactive",
					currentTs.UnixMilli(), nil,
				)
				dbClient.
					ExpectQuery(searchStmt).
					WithArgs(
						"%"+p.Keyword+"%",
						"%"+p.Keyword+"%",
						p.Statuses[0],
						p.Providers[0],
						p.Codes[0],
					).
					WillReturnRows(searchRows)

				dbClient.
					ExpectQuery(countStmt).
					WithArgs(
						"%"+p.Keyword+"%",
						"%"+p.Keyword+"%",
						p.Statuses[0],
						p.Providers[0],
						p.Codes[0],
					).
					WillReturnError(fmt.Errorf("network error"))

				res, err := barrelRepo.SearchBarrel(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("there is one client", func() {
			It("should return result", func() {
				searchRows := sqlmock.NewRows([]string{
					"id", "code",
					"name", "provider", "status",
					"created_at", "updated_at",
				}).AddRow(
					"id-1", "code-1",
					"name-1", "goseidon_hippo", "inactive",
					currentTs.UnixMilli(), nil,
				)
				dbClient.
					ExpectQuery(searchStmt).
					WithArgs(
						"%"+p.Keyword+"%",
						"%"+p.Keyword+"%",
						p.Statuses[0],
						p.Providers[0],
						p.Codes[0],
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
						p.Providers[0],
						p.Codes[0],
					).
					WillReturnRows(countRows)

				res, err := barrelRepo.SearchBarrel(ctx, p)

				r := &repository.SearchBarrelResult{
					Summary: repository.SearchBarrelSummary{
						TotalItems: 1,
					},
					Items: []repository.SearchBarrelItem{
						{
							Id:        "id-1",
							Code:      "code-1",
							Name:      "name-1",
							Provider:  "goseidon_hippo",
							Status:    "inactive",
							CreatedAt: time.UnixMilli(currentTs.UnixMilli()).UTC(),
							UpdatedAt: nil,
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
						p.Providers[0],
						p.Codes[0],
					).
					WillReturnRows(searchRows)

				dbClient.
					ExpectQuery(countStmt).
					WithArgs(
						"%"+p.Keyword+"%",
						"%"+p.Keyword+"%",
						p.Statuses[0],
						p.Providers[0],
						p.Codes[0],
					).
					WillReturnRows(countRows)

				res, err := barrelRepo.SearchBarrel(ctx, p)

				Expect(res).To(Equal(r))
				Expect(err).To(BeNil())
			})
		})
	})
})
