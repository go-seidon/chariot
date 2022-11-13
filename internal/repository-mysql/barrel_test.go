package repository_mysql_test

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
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

		When("failed rollback during check client", func() {
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

		When("failed check client", func() {
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

		When("client already exists", func() {
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

		When("failed rollback during client creation", func() {
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

		When("failed create client", func() {
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

		When("failed rollback during find client", func() {
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

		When("failed find client", func() {
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

		When("failed commit during success create client", func() {
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

		When("success create client", func() {
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
})
