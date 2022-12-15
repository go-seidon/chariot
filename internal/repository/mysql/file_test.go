package mysql_test

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-seidon/chariot/internal/repository"
	"github.com/go-seidon/chariot/internal/repository/mysql"
	random "github.com/go-seidon/provider/random/mock"
	"github.com/go-seidon/provider/typeconv"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gorm_mysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var _ = Describe("File Repository", func() {
	Context("CreateFile function", Label("unit"), func() {
		var (
			ctx                context.Context
			currentTs          time.Time
			dbClient           sqlmock.Sqlmock
			randomizer         *random.MockRandomizer
			fileRepo           repository.File
			p                  repository.CreateFileParam
			checkStmt          string
			createFileStmt     string
			createMetaStmt     string
			createLocationStmt string
			findFileStmt       string
			findMetaStmt       string
			findLocationStmt   string
			findFileRows       *sqlmock.Rows
			findLocationRows   *sqlmock.Rows
			findMetaRows       *sqlmock.Rows
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

			gormClient, err := gorm.Open(gorm_mysql.New(gorm_mysql.Config{
				Conn:                      db,
				SkipInitializeWithVersion: true,
			}), &gorm.Config{
				DisableAutomaticPing: true,
			})
			if err != nil {
				AbortSuite("failed create gorm client: " + err.Error())
			}

			t := GinkgoT()
			ctrl := gomock.NewController(t)
			randomizer = random.NewMockRandomizer(ctrl)
			fileRepo = mysql.NewFile(mysql.FileParam{
				GormClient: gormClient,
				Randomizer: randomizer,
			})

			p = repository.CreateFileParam{
				Id:         "id",
				Slug:       "dolphin-22.jpg",
				Name:       "Dolphin 22",
				Mimetype:   "image/jpeg",
				Extension:  "jpg",
				Size:       23442,
				Visibility: "public",
				Status:     "active",
				Meta: map[string]string{
					"feature": "profile",
					"user_id": "8c7ffa05-70c7-437e-8166-0f6a651a9575",
				},
				Locations: []repository.CreateFileLocation{
					{
						Id:         "i1",
						BarrelId:   "h1",
						ExternalId: typeconv.String("e1"),
						Priority:   1,
						Status:     "available",
						CreatedAt:  currentTs,
						UploadedAt: &currentTs,
					},
					{
						Id:         "i2",
						BarrelId:   "b1",
						ExternalId: typeconv.String("e2"),
						Priority:   2,
						Status:     "uploading",
						CreatedAt:  currentTs,
						UploadedAt: nil,
					},
				},
				CreatedAt:  currentTs,
				UploadedAt: currentTs,
			}
			checkStmt = regexp.QuoteMeta("SELECT id, slug FROM `file` WHERE slug = ? ORDER BY `file`.`id` LIMIT 1")
			createFileStmt = regexp.QuoteMeta("INSERT INTO `file` (`id`,`slug`,`name`,`mimetype`,`extension`,`size`,`visibility`,`status`,`uploaded_at`,`created_at`,`updated_at`) VALUES (?,?,?,?,?,?,?,?,?,?,?)")
			createMetaStmt = regexp.QuoteMeta("INSERT INTO `file_meta` (`file_id`,`key`,`value`) VALUES (?,?,?),(?,?,?) ON DUPLICATE KEY UPDATE `file_id`=VALUES(`file_id`)")
			createLocationStmt = regexp.QuoteMeta("INSERT INTO `file_location` (`id`,`file_id`,`barrel_id`,`external_id`,`priority`,`status`,`uploaded_at`,`created_at`,`updated_at`) VALUES (?,?,?,?,?,?,?,?,?),(?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE `file_id`=VALUES(`file_id`)")
			findFileStmt = regexp.QuoteMeta("SELECT id, slug, name, mimetype, extension, size, visibility, status, created_at, uploaded_at FROM `file` WHERE id = ? ORDER BY `file`.`id` LIMIT 1")
			findMetaStmt = regexp.QuoteMeta("SELECT file_id, `key`, value FROM `file_meta` WHERE `file_meta`.`file_id` = ?")
			findLocationStmt = regexp.QuoteMeta("SELECT file_id, barrel_id, external_id, priority, status, created_at, uploaded_at FROM `file_location` WHERE `file_location`.`file_id` = ?")

			findFileRows = sqlmock.NewRows([]string{
				"id", "slug", "name", "mimetype",
				"extension", "size", "visibility",
				"status", "created_at", "uploaded_at",
			}).AddRow(
				p.Id, p.Slug, p.Name, p.Mimetype,
				p.Extension, p.Size, p.Visibility,
				p.Status, p.CreatedAt.UnixMilli(), p.UploadedAt.UnixMilli(),
			)

			findLocationRows = sqlmock.NewRows([]string{
				"file_id", "barrel_id", "external_id",
				"priority", "status",
				"created_at", "uploaded_at",
			})
			for _, location := range p.Locations {
				var uploadedAt *int64
				if location.UploadedAt != nil {
					uploadedAt = typeconv.Int64(location.UploadedAt.UnixMilli())
				}
				findLocationRows.AddRow(
					p.Id,
					location.BarrelId,
					location.ExternalId,
					location.Priority,
					location.Status,
					location.CreatedAt.UnixMilli(),
					uploadedAt,
				)
			}

			findMetaRows = sqlmock.NewRows([]string{
				"file_id", "key", "value",
			})
			for key, value := range p.Meta {
				findMetaRows.AddRow(p.Id, key, value)
			}
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

				res, err := fileRepo.CreateFile(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("begin error")))
			})
		})

		When("failed rollback during slug checking", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectQuery(checkStmt).
					WithArgs(p.Slug).
					WillReturnError(fmt.Errorf("network error"))

				dbClient.
					ExpectRollback().
					WillReturnError(fmt.Errorf("rollback error"))

				res, err := fileRepo.CreateFile(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("rollback error")))
			})
		})

		When("failed check slug", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectQuery(checkStmt).
					WithArgs(p.Slug).
					WillReturnError(fmt.Errorf("network error"))

				dbClient.
					ExpectRollback()

				res, err := fileRepo.CreateFile(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("failed rollback during generate slug prefix", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				checkRows := sqlmock.
					NewRows([]string{"id", "slug"}).
					AddRow("id", "slug")
				dbClient.
					ExpectQuery(checkStmt).
					WithArgs(p.Slug).
					WillReturnRows(checkRows)

				randomizer.
					EXPECT().
					String(gomock.Eq(7)).
					Return("", fmt.Errorf("disk error")).
					Times(1)

				dbClient.
					ExpectRollback().
					WillReturnError(fmt.Errorf("network error"))

				res, err := fileRepo.CreateFile(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("failed generate slug prefix", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				checkRows := sqlmock.
					NewRows([]string{"id", "slug"}).
					AddRow("id", "slug")
				dbClient.
					ExpectQuery(checkStmt).
					WithArgs(p.Slug).
					WillReturnRows(checkRows)

				randomizer.
					EXPECT().
					String(gomock.Eq(7)).
					Return("", fmt.Errorf("disk error")).
					Times(1)

				dbClient.
					ExpectRollback()

				res, err := fileRepo.CreateFile(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("disk error")))
			})
		})

		When("failed rollback during create file", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectQuery(checkStmt).
					WithArgs(p.Slug).
					WillReturnError(gorm.ErrRecordNotFound)

				dbClient.
					ExpectExec(createFileStmt).
					WithArgs(
						p.Id, p.Slug, p.Name,
						p.Mimetype, p.Extension, p.Size,
						p.Visibility, p.Status,
						p.CreatedAt.UnixMilli(),
						p.CreatedAt.UnixMilli(),
						p.UploadedAt.UnixMilli(),
					).
					WillReturnError(fmt.Errorf("network error"))

				dbClient.
					ExpectRollback().
					WillReturnError(fmt.Errorf("rollback error"))

				res, err := fileRepo.CreateFile(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("rollback error")))
			})
		})

		When("failed create file", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectQuery(checkStmt).
					WithArgs(p.Slug).
					WillReturnError(gorm.ErrRecordNotFound)

				dbClient.
					ExpectExec(createFileStmt).
					WithArgs(
						p.Id, p.Slug, p.Name,
						p.Mimetype, p.Extension, p.Size,
						p.Visibility, p.Status,
						p.CreatedAt.UnixMilli(),
						p.CreatedAt.UnixMilli(),
						p.UploadedAt.UnixMilli(),
					).
					WillReturnError(fmt.Errorf("network error"))

				dbClient.
					ExpectRollback()

				res, err := fileRepo.CreateFile(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("failed rollback during check uploaded file", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectQuery(checkStmt).
					WithArgs(p.Slug).
					WillReturnError(gorm.ErrRecordNotFound)

				dbClient.
					ExpectExec(createFileStmt).
					WithArgs(
						p.Id, p.Slug, p.Name,
						p.Mimetype, p.Extension, p.Size,
						p.Visibility, p.Status,
						p.CreatedAt.UnixMilli(),
						p.CreatedAt.UnixMilli(),
						p.UploadedAt.UnixMilli(),
					).
					WillReturnResult(sqlmock.NewResult(1, 1))

				dbClient.
					ExpectExec(createMetaStmt).
					WillReturnResult(sqlmock.NewResult(1, 1))

				dbClient.
					ExpectExec(createLocationStmt).
					WillReturnResult(sqlmock.NewResult(1, 1))

				dbClient.
					ExpectQuery(findFileStmt).
					WithArgs(p.Id).
					WillReturnError(fmt.Errorf("network error"))

				dbClient.
					ExpectRollback().
					WillReturnError(fmt.Errorf("rollback error"))

				res, err := fileRepo.CreateFile(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("rollback error")))
			})
		})

		When("failed check uploaded file", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectQuery(checkStmt).
					WithArgs(p.Slug).
					WillReturnError(gorm.ErrRecordNotFound)

				dbClient.
					ExpectExec(createFileStmt).
					WithArgs(
						p.Id, p.Slug, p.Name,
						p.Mimetype, p.Extension, p.Size,
						p.Visibility, p.Status,
						p.CreatedAt.UnixMilli(),
						p.CreatedAt.UnixMilli(),
						p.UploadedAt.UnixMilli(),
					).
					WillReturnResult(sqlmock.NewResult(1, 1))

				dbClient.
					ExpectExec(createMetaStmt).
					WillReturnResult(sqlmock.NewResult(1, 1))

				dbClient.
					ExpectExec(createLocationStmt).
					WillReturnResult(sqlmock.NewResult(1, 1))

				dbClient.
					ExpectQuery(findFileStmt).
					WithArgs(p.Id).
					WillReturnError(fmt.Errorf("network error"))

				dbClient.
					ExpectRollback()

				res, err := fileRepo.CreateFile(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("failed commit trx", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectQuery(checkStmt).
					WithArgs(p.Slug).
					WillReturnError(gorm.ErrRecordNotFound)

				dbClient.
					ExpectExec(createFileStmt).
					WithArgs(
						p.Id, p.Slug, p.Name,
						p.Mimetype, p.Extension, p.Size,
						p.Visibility, p.Status,
						p.CreatedAt.UnixMilli(),
						p.CreatedAt.UnixMilli(),
						p.UploadedAt.UnixMilli(),
					).
					WillReturnResult(sqlmock.NewResult(1, 1))

				dbClient.
					ExpectExec(createMetaStmt).
					WillReturnResult(sqlmock.NewResult(1, 1))

				dbClient.
					ExpectExec(createLocationStmt).
					WillReturnResult(sqlmock.NewResult(1, 1))

				dbClient.
					ExpectQuery(findFileStmt).
					WithArgs(p.Id).
					WillReturnRows(findFileRows)

				dbClient.
					ExpectQuery(findLocationStmt).
					WithArgs(p.Id).
					WillReturnRows(findLocationRows)

				dbClient.
					ExpectQuery(findMetaStmt).
					WithArgs(p.Id).
					WillReturnRows(findMetaRows)

				dbClient.
					ExpectCommit().
					WillReturnError(fmt.Errorf("commit error"))

				res, err := fileRepo.CreateFile(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("commit error")))
			})
		})

		When("success commit trx", func() {
			It("should return result", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectQuery(checkStmt).
					WithArgs(p.Slug).
					WillReturnError(gorm.ErrRecordNotFound)

				dbClient.
					ExpectExec(createFileStmt).
					WithArgs(
						p.Id, p.Slug, p.Name,
						p.Mimetype, p.Extension, p.Size,
						p.Visibility, p.Status,
						p.CreatedAt.UnixMilli(),
						p.CreatedAt.UnixMilli(),
						p.UploadedAt.UnixMilli(),
					).
					WillReturnResult(sqlmock.NewResult(1, 1))

				dbClient.
					ExpectExec(createMetaStmt).
					WillReturnResult(sqlmock.NewResult(1, 1))

				dbClient.
					ExpectExec(createLocationStmt).
					WillReturnResult(sqlmock.NewResult(1, 1))

				dbClient.
					ExpectQuery(findFileStmt).
					WithArgs(p.Id).
					WillReturnRows(findFileRows)

				dbClient.
					ExpectQuery(findLocationStmt).
					WithArgs(p.Id).
					WillReturnRows(findLocationRows)

				dbClient.
					ExpectQuery(findMetaStmt).
					WithArgs(p.Id).
					WillReturnRows(findMetaRows)

				dbClient.
					ExpectCommit()

				res, err := fileRepo.CreateFile(ctx, p)

				Expect(res.Id).To(Equal(p.Id))
				Expect(res.Slug).To(Equal(p.Slug))
				Expect(res.Name).To(Equal(p.Name))
				Expect(res.Mimetype).To(Equal(p.Mimetype))
				Expect(res.Extension).To(Equal(p.Extension))
				Expect(res.Size).To(Equal(p.Size))
				Expect(res.Visibility).To(Equal(p.Visibility))
				Expect(res.Status).To(Equal(p.Status))
				Expect(res.CreatedAt).To(Equal(time.UnixMilli(p.CreatedAt.UnixMilli()).UTC()))
				Expect(res.UploadedAt).To(Equal(time.UnixMilli(p.UploadedAt.UnixMilli()).UTC()))
				Expect(res.Meta).To(Equal(p.Meta))
				Expect(len(res.Locations)).To(Equal(2))
				Expect(err).To(BeNil())
			})
		})

		When("slug is already used", func() {
			It("should return result", func() {
				dbClient.
					ExpectBegin()

				checkRows := sqlmock.
					NewRows([]string{"id", "slug"}).
					AddRow("id", "slug")
				dbClient.
					ExpectQuery(checkStmt).
					WithArgs(p.Slug).
					WillReturnRows(checkRows)

				randomizer.
					EXPECT().
					String(gomock.Eq(7)).
					Return("abcdefg", nil).
					Times(1)

				slug := "dolphin-22-abcdefg.jpg"
				dbClient.
					ExpectExec(createFileStmt).
					WithArgs(
						p.Id, slug, p.Name,
						p.Mimetype, p.Extension, p.Size,
						p.Visibility, p.Status,
						p.CreatedAt.UnixMilli(),
						p.CreatedAt.UnixMilli(),
						p.UploadedAt.UnixMilli(),
					).
					WillReturnResult(sqlmock.NewResult(1, 1))

				dbClient.
					ExpectExec(createMetaStmt).
					WillReturnResult(sqlmock.NewResult(1, 1))

				dbClient.
					ExpectExec(createLocationStmt).
					WillReturnResult(sqlmock.NewResult(1, 1))

				findFileRows := sqlmock.NewRows([]string{
					"id", "slug", "name", "mimetype",
					"extension", "size", "visibility",
					"status", "created_at", "uploaded_at",
				}).AddRow(
					p.Id, slug, p.Name, p.Mimetype,
					p.Extension, p.Size, p.Visibility,
					p.Status, p.CreatedAt.UnixMilli(), p.UploadedAt.UnixMilli(),
				)
				dbClient.
					ExpectQuery(findFileStmt).
					WithArgs(p.Id).
					WillReturnRows(findFileRows)

				dbClient.
					ExpectQuery(findLocationStmt).
					WithArgs(p.Id).
					WillReturnRows(findLocationRows)

				dbClient.
					ExpectQuery(findMetaStmt).
					WithArgs(p.Id).
					WillReturnRows(findMetaRows)

				dbClient.
					ExpectCommit()

				res, err := fileRepo.CreateFile(ctx, p)

				Expect(res.Id).To(Equal(p.Id))
				Expect(res.Slug).To(Equal(slug))
				Expect(res.Name).To(Equal(p.Name))
				Expect(res.Mimetype).To(Equal(p.Mimetype))
				Expect(res.Extension).To(Equal(p.Extension))
				Expect(res.Size).To(Equal(p.Size))
				Expect(res.Visibility).To(Equal(p.Visibility))
				Expect(res.Status).To(Equal(p.Status))
				Expect(res.CreatedAt).To(Equal(time.UnixMilli(p.CreatedAt.UnixMilli()).UTC()))
				Expect(res.UploadedAt).To(Equal(time.UnixMilli(p.UploadedAt.UnixMilli()).UTC()))
				Expect(res.Meta).To(Equal(p.Meta))
				Expect(len(res.Locations)).To(Equal(2))
				Expect(err).To(BeNil())
			})
		})
	})

	Context("FindFile function", Label("unit"), func() {

		var (
			ctx              context.Context
			currentTs        time.Time
			dbClient         sqlmock.Sqlmock
			fileRepo         repository.File
			p                repository.FindFileParam
			r                *repository.FindFileResult
			findStmt         string
			findMetaStmt     string
			findLocationStmt string
			findBarrelStmt   string
			findRows         *sqlmock.Rows
			findMetaRows     *sqlmock.Rows
			findLocationRows *sqlmock.Rows
			findBarrelRows   *sqlmock.Rows
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

			gormClient, err := gorm.Open(gorm_mysql.New(gorm_mysql.Config{
				Conn:                      db,
				SkipInitializeWithVersion: true,
			}), &gorm.Config{
				DisableAutomaticPing: true,
			})
			if err != nil {
				AbortSuite("failed create gorm client: " + err.Error())
			}
			fileRepo = mysql.NewFile(mysql.FileParam{
				GormClient: gormClient,
			})

			p = repository.FindFileParam{
				Id: "id",
			}
			r = &repository.FindFileResult{
				Id:         "id",
				Slug:       "dolphin-22.jpg",
				Name:       "Dolhpin 22",
				Mimetype:   "image/jpeg",
				Extension:  "jpg",
				Size:       23343,
				Visibility: "public",
				Status:     "available",
				CreatedAt:  time.UnixMilli(currentTs.UnixMilli()).UTC(),
				UpdatedAt:  typeconv.Time(time.UnixMilli(currentTs.UnixMilli()).UTC()),
				UploadedAt: time.UnixMilli(currentTs.UnixMilli()).UTC(),
				DeletedAt:  nil,
				Meta: map[string]string{
					"feature": "profile",
					"user_id": "123",
				},
				Locations: []repository.FindFileLocation{
					{
						ExternalId: typeconv.String("e1"),
						Priority:   1,
						Status:     "available",
						CreatedAt:  time.UnixMilli(currentTs.UnixMilli()).UTC(),
						UpdatedAt:  typeconv.Time(time.UnixMilli(currentTs.UnixMilli()).UTC()),
						UploadedAt: typeconv.Time(time.UnixMilli(currentTs.UnixMilli()).UTC()),
						Barrel: repository.FindFileBarrel{
							Id:       "b1",
							Code:     "b1",
							Provider: "goseidon_hippo",
							Status:   "active",
						},
					},
					{
						ExternalId: nil,
						Priority:   2,
						Status:     "uploading",
						CreatedAt:  time.UnixMilli(currentTs.UnixMilli()).UTC(),
						UpdatedAt:  typeconv.Time(time.UnixMilli(currentTs.UnixMilli()).UTC()),
						UploadedAt: nil,
						Barrel: repository.FindFileBarrel{
							Id:       "b2",
							Code:     "b2",
							Provider: "goseidon_hippo",
							Status:   "active",
						},
					},
				},
			}
			findStmt = regexp.QuoteMeta("SELECT file.id, file.slug, file.name, file.mimetype, file.extension, file.size, file.visibility, file.status, file.created_at, file.updated_at, file.deleted_at, file.uploaded_at FROM `file` WHERE file.id = ? ORDER BY `file`.`id` LIMIT 1")
			findMetaStmt = regexp.QuoteMeta("SELECT file_id, `key`, value FROM `file_meta` WHERE `file_meta`.`file_id` = ?")
			findLocationStmt = regexp.QuoteMeta("SELECT id, file_id, barrel_id, external_id, priority, status, created_at, updated_at, uploaded_at FROM `file_location` WHERE `file_location`.`file_id` = ?")
			findBarrelStmt = regexp.QuoteMeta("SELECT id, code, provider, status FROM `barrel` WHERE `barrel`.`id` IN (?,?)")

			findRows = sqlmock.NewRows([]string{
				"id", "slug", "name", "mimetype",
				"extension", "size", "visibility", "status",
				"created_at", "updated_at", "deleted_at", "uploaded_at",
			}).AddRow(
				r.Id, r.Slug, r.Name, r.Mimetype,
				r.Extension, r.Size, r.Visibility, r.Status,
				currentTs.UnixMilli(), currentTs.UnixMilli(),
				nil, currentTs.UnixMilli(),
			)

			findLocationRows = sqlmock.NewRows([]string{
				"file_id", "barrel_id", "external_id",
				"priority", "status",
				"created_at", "updated_at", "uploaded_at",
			}).AddRow(
				"id", "b1", "e1",
				1, "available",
				currentTs.UnixMilli(),
				currentTs.UnixMilli(),
				currentTs.UnixMilli(),
			).AddRow(
				"id", "b2", nil,
				2, "uploading",
				currentTs.UnixMilli(),
				currentTs.UnixMilli(),
				nil,
			)

			findMetaRows = sqlmock.NewRows([]string{
				"file_id", "key", "value",
			}).
				AddRow("id", "user_id", "123").
				AddRow("id", "feature", "profile")

			findBarrelRows = sqlmock.NewRows([]string{
				"id", "code", "provider", "status",
			}).
				AddRow("b1", "b1", "goseidon_hippo", "active").
				AddRow("b2", "b2", "goseidon_hippo", "active")
		})

		AfterEach(func() {
			err := dbClient.ExpectationsWereMet()
			if err != nil {
				AbortSuite("some expectations were not met " + err.Error())
			}
		})

		When("failed find file", func() {
			It("should return error", func() {
				dbClient.
					ExpectQuery(findStmt).
					WithArgs(p.Id).
					WillReturnError(fmt.Errorf("network error"))

				res, err := fileRepo.FindFile(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("file is not available", func() {
			It("should return error", func() {
				dbClient.
					ExpectQuery(findStmt).
					WithArgs(p.Id).
					WillReturnError(gorm.ErrRecordNotFound)

				res, err := fileRepo.FindFile(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(repository.ErrNotFound))
			})
		})

		When("success find file", func() {
			It("should return result", func() {
				dbClient.
					ExpectQuery(findStmt).
					WithArgs(p.Id).
					WillReturnRows(findRows)

				dbClient.
					ExpectQuery(findLocationStmt).
					WithArgs(p.Id).
					WillReturnRows(findLocationRows)

				dbClient.
					ExpectQuery(findBarrelStmt).
					WithArgs("b1", "b2").
					WillReturnRows(findBarrelRows)

				dbClient.
					ExpectQuery(findMetaStmt).
					WithArgs(p.Id).
					WillReturnRows(findMetaRows)

				res, err := fileRepo.FindFile(ctx, p)

				Expect(res).To(Equal(r))
				Expect(err).To(BeNil())
			})
		})

		When("success find deleted file", func() {
			It("should return result", func() {
				findRows := sqlmock.
					NewRows([]string{
						"id", "slug", "name", "mimetype",
						"extension", "size", "visibility", "status",
						"created_at", "updated_at", "deleted_at", "uploaded_at",
					}).
					AddRow(
						r.Id, r.Slug, r.Name, r.Mimetype,
						r.Extension, r.Size, r.Visibility, r.Status,
						currentTs.UnixMilli(), currentTs.UnixMilli(),
						currentTs.UnixMilli(), currentTs.UnixMilli(),
					)
				dbClient.
					ExpectQuery(findStmt).
					WithArgs(p.Id).
					WillReturnRows(findRows)

				findLocationRows := sqlmock.NewRows([]string{
					"file_id", "barrel_id", "external_id",
					"priority", "status",
					"created_at", "updated_at", "uploaded_at",
				})
				dbClient.
					ExpectQuery(findLocationStmt).
					WithArgs(p.Id).
					WillReturnRows(findLocationRows)

				findMetaRows := sqlmock.NewRows([]string{
					"file_id", "key", "value",
				})
				dbClient.
					ExpectQuery(findMetaStmt).
					WithArgs(p.Id).
					WillReturnRows(findMetaRows)

				res, err := fileRepo.FindFile(ctx, p)

				r := &repository.FindFileResult{
					Id:         "id",
					Slug:       "dolphin-22.jpg",
					Name:       "Dolhpin 22",
					Mimetype:   "image/jpeg",
					Extension:  "jpg",
					Size:       23343,
					Visibility: "public",
					Status:     "available",
					CreatedAt:  time.UnixMilli(currentTs.UnixMilli()).UTC(),
					UpdatedAt:  typeconv.Time(time.UnixMilli(currentTs.UnixMilli()).UTC()),
					UploadedAt: time.UnixMilli(currentTs.UnixMilli()).UTC(),
					DeletedAt:  typeconv.Time(time.UnixMilli(currentTs.UnixMilli()).UTC()),
					Meta:       map[string]string{},
					Locations:  []repository.FindFileLocation{},
				}
				Expect(res).To(Equal(r))
				Expect(err).To(BeNil())
			})
		})

		When("success find using slug", func() {
			It("should return result", func() {
				p := repository.FindFileParam{
					Slug: "dolphin-22.jpg",
				}

				findRows := sqlmock.
					NewRows([]string{
						"id", "slug", "name", "mimetype",
						"extension", "size", "visibility", "status",
						"created_at", "updated_at", "deleted_at", "uploaded_at",
					}).
					AddRow(
						r.Id, r.Slug, r.Name, r.Mimetype,
						r.Extension, r.Size, r.Visibility, r.Status,
						currentTs.UnixMilli(), currentTs.UnixMilli(),
						currentTs.UnixMilli(), currentTs.UnixMilli(),
					)
				findStmt := regexp.QuoteMeta("SELECT file.id, file.slug, file.name, file.mimetype, file.extension, file.size, file.visibility, file.status, file.created_at, file.updated_at, file.deleted_at, file.uploaded_at FROM `file` WHERE file.slug = ? ORDER BY `file`.`id` LIMIT 1")
				dbClient.
					ExpectQuery(findStmt).
					WithArgs(p.Slug).
					WillReturnRows(findRows)

				findLocationRows := sqlmock.NewRows([]string{
					"file_id", "barrel_id", "external_id",
					"priority", "status",
					"created_at", "updated_at", "uploaded_at",
				})
				dbClient.
					ExpectQuery(findLocationStmt).
					WithArgs("id").
					WillReturnRows(findLocationRows)

				findMetaRows := sqlmock.NewRows([]string{
					"file_id", "key", "value",
				})
				dbClient.
					ExpectQuery(findMetaStmt).
					WithArgs("id").
					WillReturnRows(findMetaRows)

				res, err := fileRepo.FindFile(ctx, p)

				r := &repository.FindFileResult{
					Id:         "id",
					Slug:       "dolphin-22.jpg",
					Name:       "Dolhpin 22",
					Mimetype:   "image/jpeg",
					Extension:  "jpg",
					Size:       23343,
					Visibility: "public",
					Status:     "available",
					CreatedAt:  time.UnixMilli(currentTs.UnixMilli()).UTC(),
					UpdatedAt:  typeconv.Time(time.UnixMilli(currentTs.UnixMilli()).UTC()),
					UploadedAt: time.UnixMilli(currentTs.UnixMilli()).UTC(),
					DeletedAt:  typeconv.Time(time.UnixMilli(currentTs.UnixMilli()).UTC()),
					Meta:       map[string]string{},
					Locations:  []repository.FindFileLocation{},
				}
				Expect(res).To(Equal(r))
				Expect(err).To(BeNil())
			})
		})

		When("success find using location id", func() {
			It("should return result", func() {
				p := repository.FindFileParam{
					LocationId: "loc-id",
				}

				findRows := sqlmock.
					NewRows([]string{
						"id", "slug", "name", "mimetype",
						"extension", "size", "visibility", "status",
						"created_at", "updated_at", "deleted_at", "uploaded_at",
					}).
					AddRow(
						r.Id, r.Slug, r.Name, r.Mimetype,
						r.Extension, r.Size, r.Visibility, r.Status,
						currentTs.UnixMilli(), currentTs.UnixMilli(),
						currentTs.UnixMilli(), currentTs.UnixMilli(),
					)
				findStmt := regexp.QuoteMeta("SELECT file.id, file.slug, file.name, file.mimetype, file.extension, file.size, file.visibility, file.status, file.created_at, file.updated_at, file.deleted_at, file.uploaded_at FROM `file` LEFT JOIN file_location AS fl ON fl.file_id = file.id WHERE fl.id = ? ORDER BY `file`.`id` LIMIT 1")
				dbClient.
					ExpectQuery(findStmt).
					WithArgs(p.LocationId).
					WillReturnRows(findRows)

				findLocationRows := sqlmock.NewRows([]string{
					"file_id", "barrel_id", "external_id",
					"priority", "status",
					"created_at", "updated_at", "uploaded_at",
				})
				dbClient.
					ExpectQuery(findLocationStmt).
					WithArgs("id").
					WillReturnRows(findLocationRows)

				findMetaRows := sqlmock.NewRows([]string{
					"file_id", "key", "value",
				})
				dbClient.
					ExpectQuery(findMetaStmt).
					WithArgs("id").
					WillReturnRows(findMetaRows)

				res, err := fileRepo.FindFile(ctx, p)

				r := &repository.FindFileResult{
					Id:         "id",
					Slug:       "dolphin-22.jpg",
					Name:       "Dolhpin 22",
					Mimetype:   "image/jpeg",
					Extension:  "jpg",
					Size:       23343,
					Visibility: "public",
					Status:     "available",
					CreatedAt:  time.UnixMilli(currentTs.UnixMilli()).UTC(),
					UpdatedAt:  typeconv.Time(time.UnixMilli(currentTs.UnixMilli()).UTC()),
					UploadedAt: time.UnixMilli(currentTs.UnixMilli()).UTC(),
					DeletedAt:  typeconv.Time(time.UnixMilli(currentTs.UnixMilli()).UTC()),
					Meta:       map[string]string{},
					Locations:  []repository.FindFileLocation{},
				}
				Expect(res).To(Equal(r))
				Expect(err).To(BeNil())
			})
		})
	})

	Context("SearchFile function", Label("unit"), func() {

		var (
			ctx          context.Context
			currentTs    time.Time
			dbClient     sqlmock.Sqlmock
			fileRepo     repository.File
			p            repository.SearchFileParam
			r            *repository.SearchFileResult
			searchStmt   string
			findMetaStmt string
			countStmt    string
			searchRows   *sqlmock.Rows
			countRows    *sqlmock.Rows
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

			gormClient, err := gorm.Open(gorm_mysql.New(gorm_mysql.Config{
				Conn:                      db,
				SkipInitializeWithVersion: true,
			}), &gorm.Config{
				DisableAutomaticPing: true,
			})
			if err != nil {
				AbortSuite("failed create gorm client: " + err.Error())
			}
			fileRepo = mysql.NewFile(mysql.FileParam{
				GormClient: gormClient,
			})

			p = repository.SearchFileParam{
				Limit:         24,
				Offset:        48,
				Keyword:       "sa",
				StatusIn:      []string{"available", "deleted"},
				VisibilityIn:  []string{"public"},
				ExtensionIn:   []string{"jpg", "png"},
				SizeGte:       1024,
				SizeLte:       2048,
				UploadDateGte: 1669638670000,
				UploadDateLte: 1669638670476,
			}
			r = &repository.SearchFileResult{
				Summary: repository.SearchFileSummary{
					TotalItems: 2,
				},
				Items: []repository.SearchFileItem{
					{
						Id:         "id-1",
						Slug:       "dolphin-22.jpg",
						Mimetype:   "image/jpeg",
						Extension:  "jpg",
						Size:       1025,
						Name:       "Dolphin 22",
						Visibility: "public",
						Status:     "available",
						UploadedAt: time.UnixMilli(currentTs.UnixMilli()).UTC(),
						CreatedAt:  time.UnixMilli(currentTs.UnixMilli()).UTC(),
						UpdatedAt:  typeconv.Time(time.UnixMilli(currentTs.UnixMilli()).UTC()),
						DeletedAt:  nil,
						Meta: map[string]string{
							"feature": "profile",
							"user_id": "123",
						},
					},
					{
						Id:         "id-2",
						Slug:       "sakura.png",
						Mimetype:   "image/png",
						Extension:  "png",
						Size:       1026,
						Name:       "Sakura",
						Visibility: "public",
						Status:     "deleted",
						UploadedAt: time.UnixMilli(currentTs.UnixMilli()).UTC(),
						CreatedAt:  time.UnixMilli(currentTs.UnixMilli()).UTC(),
						UpdatedAt:  typeconv.Time(time.UnixMilli(currentTs.UnixMilli()).UTC()),
						DeletedAt:  typeconv.Time(time.UnixMilli(currentTs.UnixMilli()).UTC()),
						Meta: map[string]string{
							"feature": "background",
							"user_id": "456",
						},
					},
				},
			}
			searchStmt = regexp.QuoteMeta(strings.TrimSpace(`
				SELECT id, slug, name, mimetype, extension, size, visibility, status, uploaded_at, created_at, updated_at, deleted_at
				FROM ` + "`file`" + ` 
				WHERE name LIKE ?
				AND status IN (?,?)
				AND visibility IN (?)
				AND extension IN (?,?)
				AND size >= ?
				AND size <= ?
				AND uploaded_at >= ?
				AND uploaded_at <= ?
				LIMIT 24
				OFFSET 48
			`))
			findMetaStmt = regexp.QuoteMeta("SELECT file_id, `key`, value FROM `file_meta` WHERE `file_meta`.`file_id` IN (?,?)")
			countStmt = regexp.QuoteMeta(strings.TrimSpace(`
				SELECT count(*)
				FROM ` + "`file`" + `
				WHERE name LIKE ?
				AND status IN (?,?)
				AND visibility IN (?)
				AND extension IN (?,?)
				AND size >= ?
				AND size <= ?
				AND uploaded_at >= ?
				AND uploaded_at <= ?
				LIMIT 24
				OFFSET 48
			`))
			searchRows = sqlmock.NewRows([]string{
				"id", "slug", "name", "mimetype",
				"extension", "size", "visibility", "status",
				"uploaded_at", "created_at", "updated_at", "deleted_at",
			}).AddRow(
				"id-1", "dolphin-22.jpg", "Dolphin 22", "image/jpeg",
				"jpg", 1025, "public", "available",
				currentTs.UnixMilli(), currentTs.UnixMilli(),
				currentTs.UnixMilli(), nil,
			).AddRow(
				"id-2", "sakura.png", "Sakura", "image/png",
				"png", 1026, "public", "deleted",
				currentTs.UnixMilli(), currentTs.UnixMilli(),
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

		When("failed search file", func() {
			It("should return error", func() {
				dbClient.
					ExpectQuery(searchStmt).
					WithArgs(
						"%"+p.Keyword+"%",
						p.StatusIn[0],
						p.StatusIn[1],
						p.VisibilityIn[0],
						p.ExtensionIn[0],
						p.ExtensionIn[1],
						p.SizeGte,
						p.SizeLte,
						p.UploadDateGte,
						p.UploadDateLte,
					).
					WillReturnError(fmt.Errorf("network error"))

				res, err := fileRepo.SearchFile(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("there are no file", func() {
			It("should return empty result", func() {
				dbClient.
					ExpectQuery(searchStmt).
					WithArgs(
						"%"+p.Keyword+"%",
						p.StatusIn[0],
						p.StatusIn[1],
						p.VisibilityIn[0],
						p.ExtensionIn[0],
						p.ExtensionIn[1],
						p.SizeGte,
						p.SizeLte,
						p.UploadDateGte,
						p.UploadDateLte,
					).
					WillReturnError(gorm.ErrRecordNotFound)

				res, err := fileRepo.SearchFile(ctx, p)

				r := &repository.SearchFileResult{
					Summary: repository.SearchFileSummary{
						TotalItems: 0,
					},
					Items: []repository.SearchFileItem{},
				}
				Expect(res).To(Equal(r))
				Expect(err).To(BeNil())
			})
		})

		When("failed count search file", func() {
			It("should return result", func() {
				searchRows := sqlmock.NewRows([]string{
					"id", "slug", "name", "mimetype",
					"extension", "size", "visibility", "status",
					"uploaded_at", "created_at", "updated_at", "deleted_at",
				}).AddRow(
					"id-1", "dolphin-22.jpg", "Dolphin 22", "image/jpeg",
					"jpg", 1025, "public", "available",
					currentTs.UnixMilli(), currentTs.UnixMilli(),
					currentTs.UnixMilli(), nil,
				)
				dbClient.
					ExpectQuery(searchStmt).
					WithArgs(
						"%"+p.Keyword+"%",
						p.StatusIn[0],
						p.StatusIn[1],
						p.VisibilityIn[0],
						p.ExtensionIn[0],
						p.ExtensionIn[1],
						p.SizeGte,
						p.SizeLte,
						p.UploadDateGte,
						p.UploadDateLte,
					).
					WillReturnRows(searchRows)

				findMetaStmt := regexp.QuoteMeta("SELECT file_id, `key`, value FROM `file_meta` WHERE `file_meta`.`file_id` = ?")
				findMetaRows := sqlmock.NewRows([]string{
					"file_id", "key", "value",
				}).
					AddRow("id-1", "user_id", "123").
					AddRow("id-1", "feature", "profile")
				dbClient.
					ExpectQuery(findMetaStmt).
					WithArgs("id-1").
					WillReturnRows(findMetaRows)

				dbClient.
					ExpectQuery(countStmt).
					WithArgs(
						"%"+p.Keyword+"%",
						p.StatusIn[0],
						p.StatusIn[1],
						p.VisibilityIn[0],
						p.ExtensionIn[0],
						p.ExtensionIn[1],
						p.SizeGte,
						p.SizeLte,
						p.UploadDateGte,
						p.UploadDateLte,
					).
					WillReturnError(fmt.Errorf("network error"))

				res, err := fileRepo.SearchFile(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("there is one file", func() {
			It("should return result", func() {
				searchRows := sqlmock.NewRows([]string{
					"id", "slug", "name", "mimetype",
					"extension", "size", "visibility", "status",
					"uploaded_at", "created_at", "updated_at", "deleted_at",
				}).AddRow(
					"id-1", "dolphin-22.jpg", "Dolphin 22", "image/jpeg",
					"jpg", 1025, "public", "available",
					currentTs.UnixMilli(), currentTs.UnixMilli(),
					currentTs.UnixMilli(), nil,
				)
				dbClient.
					ExpectQuery(searchStmt).
					WithArgs(
						"%"+p.Keyword+"%",
						p.StatusIn[0],
						p.StatusIn[1],
						p.VisibilityIn[0],
						p.ExtensionIn[0],
						p.ExtensionIn[1],
						p.SizeGte,
						p.SizeLte,
						p.UploadDateGte,
						p.UploadDateLte,
					).
					WillReturnRows(searchRows)

				findMetaStmt := regexp.QuoteMeta("SELECT file_id, `key`, value FROM `file_meta` WHERE `file_meta`.`file_id` = ?")
				findMetaRows := sqlmock.NewRows([]string{
					"file_id", "key", "value",
				}).
					AddRow("id-1", "user_id", "123").
					AddRow("id-1", "feature", "profile")
				dbClient.
					ExpectQuery(findMetaStmt).
					WithArgs("id-1").
					WillReturnRows(findMetaRows)

				countRows := sqlmock.
					NewRows([]string{"count(*)"}).
					AddRow(1)
				dbClient.
					ExpectQuery(countStmt).
					WithArgs(
						"%"+p.Keyword+"%",
						p.StatusIn[0],
						p.StatusIn[1],
						p.VisibilityIn[0],
						p.ExtensionIn[0],
						p.ExtensionIn[1],
						p.SizeGte,
						p.SizeLte,
						p.UploadDateGte,
						p.UploadDateLte,
					).
					WillReturnRows(countRows)

				res, err := fileRepo.SearchFile(ctx, p)

				r := &repository.SearchFileResult{
					Summary: repository.SearchFileSummary{
						TotalItems: 1,
					},
					Items: []repository.SearchFileItem{
						{
							Id:         "id-1",
							Slug:       "dolphin-22.jpg",
							Mimetype:   "image/jpeg",
							Extension:  "jpg",
							Size:       1025,
							Name:       "Dolphin 22",
							Visibility: "public",
							Status:     "available",
							UploadedAt: time.UnixMilli(currentTs.UnixMilli()).UTC(),
							CreatedAt:  time.UnixMilli(currentTs.UnixMilli()).UTC(),
							UpdatedAt:  typeconv.Time(time.UnixMilli(currentTs.UnixMilli()).UTC()),
							DeletedAt:  nil,
							Meta: map[string]string{
								"feature": "profile",
								"user_id": "123",
							},
						},
					},
				}
				Expect(res).To(Equal(r))
				Expect(err).To(BeNil())
			})
		})

		When("there are some files", func() {
			It("should return result", func() {
				dbClient.
					ExpectQuery(searchStmt).
					WithArgs(
						"%"+p.Keyword+"%",
						p.StatusIn[0],
						p.StatusIn[1],
						p.VisibilityIn[0],
						p.ExtensionIn[0],
						p.ExtensionIn[1],
						p.SizeGte,
						p.SizeLte,
						p.UploadDateGte,
						p.UploadDateLte,
					).
					WillReturnRows(searchRows)

				findMetaRows := sqlmock.NewRows([]string{
					"file_id", "key", "value",
				}).
					AddRow("id-1", "user_id", "123").
					AddRow("id-1", "feature", "profile").
					AddRow("id-2", "user_id", "456").
					AddRow("id-2", "feature", "background")
				dbClient.
					ExpectQuery(findMetaStmt).
					WithArgs("id-1", "id-2").
					WillReturnRows(findMetaRows)

				dbClient.
					ExpectQuery(countStmt).
					WithArgs(
						"%"+p.Keyword+"%",
						p.StatusIn[0],
						p.StatusIn[1],
						p.VisibilityIn[0],
						p.ExtensionIn[0],
						p.ExtensionIn[1],
						p.SizeGte,
						p.SizeLte,
						p.UploadDateGte,
						p.UploadDateLte,
					).
					WillReturnRows(countRows)

				res, err := fileRepo.SearchFile(ctx, p)

				Expect(res).To(Equal(r))
				Expect(err).To(BeNil())
			})
		})

		DescribeTable("search using custom sort", func(sort string, query string) {
			searchStmt := regexp.QuoteMeta(strings.TrimSpace(`
				SELECT id, slug, name, mimetype, extension, size, visibility, status, uploaded_at, created_at, updated_at, deleted_at
				FROM ` + "`file`" + ` 
				WHERE name LIKE ?
				AND status IN (?,?)
				AND visibility IN (?)
				AND extension IN (?,?)
				AND size >= ?
				AND size <= ?
				AND uploaded_at >= ?
				AND uploaded_at <= ?
				ORDER BY ` + query + `
				LIMIT 24
				OFFSET 48
			`))
			dbClient.
				ExpectQuery(searchStmt).
				WithArgs(
					"%"+p.Keyword+"%",
					p.StatusIn[0],
					p.StatusIn[1],
					p.VisibilityIn[0],
					p.ExtensionIn[0],
					p.ExtensionIn[1],
					p.SizeGte,
					p.SizeLte,
					p.UploadDateGte,
					p.UploadDateLte,
				).
				WillReturnRows(searchRows)

			findMetaRows := sqlmock.NewRows([]string{
				"file_id", "key", "value",
			}).
				AddRow("id-1", "user_id", "123").
				AddRow("id-1", "feature", "profile").
				AddRow("id-2", "user_id", "456").
				AddRow("id-2", "feature", "background")
			dbClient.
				ExpectQuery(findMetaStmt).
				WithArgs("id-1", "id-2").
				WillReturnRows(findMetaRows)

			dbClient.
				ExpectQuery(countStmt).
				WithArgs(
					"%"+p.Keyword+"%",
					p.StatusIn[0],
					p.StatusIn[1],
					p.VisibilityIn[0],
					p.ExtensionIn[0],
					p.ExtensionIn[1],
					p.SizeGte,
					p.SizeLte,
					p.UploadDateGte,
					p.UploadDateLte,
				).
				WillReturnRows(countRows)

			p := repository.SearchFileParam{
				Limit:         24,
				Offset:        48,
				Keyword:       "sa",
				StatusIn:      []string{"available", "deleted"},
				VisibilityIn:  []string{"public"},
				ExtensionIn:   []string{"jpg", "png"},
				SizeGte:       1024,
				SizeLte:       2048,
				UploadDateGte: 1669638670000,
				UploadDateLte: 1669638670476,
				Sort:          sort,
			}
			res, err := fileRepo.SearchFile(ctx, p)

			Expect(res).To(Equal(r))
			Expect(err).To(BeNil())
		},
			Entry("should query using uploaded_at desc", "latest_upload", "uploaded_at DESC"),
			Entry("should query using uploaded_at asc", "newest_upload", "uploaded_at ASC"),
			Entry("should query using size desc", "highest_size", "size DESC"),
			Entry("should query using size asc", "lowest_size", "size ASC"),
		)
	})

	Context("UpdateFile function", Label("unit"), func() {

		var (
			ctx        context.Context
			currentTs  time.Time
			dbClient   sqlmock.Sqlmock
			fileRepo   repository.File
			p          repository.UpdateFileParam
			r          *repository.UpdateFileResult
			updateStmt string
			checkStmt  string
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

			gormClient, err := gorm.Open(gorm_mysql.New(gorm_mysql.Config{
				Conn:                      db,
				SkipInitializeWithVersion: true,
			}), &gorm.Config{
				DisableAutomaticPing: true,
			})
			if err != nil {
				AbortSuite("failed create gorm client: " + err.Error())
			}
			fileRepo = mysql.NewFile(mysql.FileParam{
				GormClient: gormClient,
			})

			p = repository.UpdateFileParam{
				Id:        "id",
				Status:    typeconv.String("deleting"),
				UpdatedAt: currentTs,
			}
			r = &repository.UpdateFileResult{
				Id:         "id",
				Slug:       "slug",
				Name:       "name",
				Mimetype:   "mimetype",
				Extension:  "extension",
				Size:       123,
				Visibility: "public",
				Status:     "deleting",
				CreatedAt:  time.UnixMilli(currentTs.UnixMilli()).UTC(),
				UpdatedAt:  time.UnixMilli(currentTs.UnixMilli()).UTC(),
				UploadedAt: time.UnixMilli(currentTs.UnixMilli()).UTC(),
				DeletedAt:  typeconv.Time(time.UnixMilli(currentTs.UnixMilli()).UTC()),
			}
			updateStmt = regexp.QuoteMeta("UPDATE `file` SET `status`=?,`updated_at`=? WHERE id = ?")
			checkStmt = regexp.QuoteMeta("SELECT id, slug, name, mimetype, extension, size, visibility, status, created_at, updated_at, uploaded_at, deleted_at FROM `file` WHERE id = ? ORDER BY `file`.`id` LIMIT 1")
			checkRows = sqlmock.NewRows([]string{
				"id", "slug", "name", "mimetype",
				"extension", "size", "visibility", "status",
				"created_at", "updated_at",
				"uploaded_at", "deleted_at",
			}).AddRow(
				"id", "slug", "name", "mimetype",
				"extension", 123, "public", "deleting",
				currentTs.UnixMilli(), currentTs.UnixMilli(),
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

				res, err := fileRepo.UpdateFile(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("begin error")))
			})
		})

		When("failed rollback during update file", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectExec(updateStmt).
					WithArgs(
						p.Status,
						p.UpdatedAt.UnixMilli(),
						p.Id,
					).
					WillReturnError(fmt.Errorf("network error"))

				dbClient.
					ExpectRollback().
					WillReturnError(fmt.Errorf("rollback error"))

				res, err := fileRepo.UpdateFile(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("rollback error")))
			})
		})

		When("failed update file", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectExec(updateStmt).
					WithArgs(
						p.Status,
						p.UpdatedAt.UnixMilli(),
						p.Id,
					).
					WillReturnError(fmt.Errorf("network error"))

				dbClient.
					ExpectRollback()

				res, err := fileRepo.UpdateFile(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("failed rollback during check update result", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectExec(updateStmt).
					WithArgs(
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

				res, err := fileRepo.UpdateFile(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("rollback error")))
			})
		})

		When("failed check update result", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectExec(updateStmt).
					WithArgs(
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

				res, err := fileRepo.UpdateFile(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("failed commit trx", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectExec(updateStmt).
					WithArgs(
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

				res, err := fileRepo.UpdateFile(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("commit error")))
			})
		})

		When("success update file", func() {
			It("should return result", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectExec(updateStmt).
					WithArgs(
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

				res, err := fileRepo.UpdateFile(ctx, p)

				Expect(res).To(Equal(r))
				Expect(err).To(BeNil())
			})
		})
	})

	Context("SearchLocation function", Label("unit"), func() {

		var (
			ctx        context.Context
			dbClient   sqlmock.Sqlmock
			fileRepo   repository.File
			p          repository.SearchLocationParam
			r          *repository.SearchLocationResult
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
			db, dbClient, err = sqlmock.New()
			if err != nil {
				AbortSuite("failed create db mock: " + err.Error())
			}

			gormClient, err := gorm.Open(gorm_mysql.New(gorm_mysql.Config{
				Conn:                      db,
				SkipInitializeWithVersion: true,
			}), &gorm.Config{
				DisableAutomaticPing: true,
			})
			if err != nil {
				AbortSuite("failed create gorm client: " + err.Error())
			}
			fileRepo = mysql.NewFile(mysql.FileParam{
				GormClient: gormClient,
			})

			p = repository.SearchLocationParam{
				Limit:    5,
				Statuses: []string{"pending"},
			}
			r = &repository.SearchLocationResult{
				Summary: repository.SearchLocationSummary{
					TotalItems: 2,
				},
				Items: []repository.SearchLocationItem{
					{
						Id:       "id1",
						FileId:   "f1",
						BarrelId: "b1",
						Priority: 2,
						Status:   "pending",
					},
					{
						Id:       "id2",
						FileId:   "f2",
						BarrelId: "b2",
						Priority: 2,
						Status:   "pending",
					},
				},
			}
			searchStmt = regexp.QuoteMeta(strings.TrimSpace(`
				SELECT id, file_id, barrel_id, priority, status
				FROM ` + "`file_location`" + ` 
				WHERE status IN (?)
				ORDER BY created_at ASC
				LIMIT 5
			`))
			countStmt = regexp.QuoteMeta(strings.TrimSpace(`
				SELECT count(*)
				FROM ` + "`file_location`" + `
				WHERE status IN (?)
				LIMIT 5
			`))
			searchRows = sqlmock.NewRows([]string{
				"id", "file_id", "barrel_id",
				"priority", "status",
			}).AddRow(
				"id1", "f1", "b1",
				2, "pending",
			).AddRow(
				"id2", "f2", "b2",
				2, "pending",
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

		When("failed search location", func() {
			It("should return error", func() {
				dbClient.
					ExpectQuery(searchStmt).
					WithArgs(
						p.Statuses[0],
					).
					WillReturnError(fmt.Errorf("network error"))

				res, err := fileRepo.SearchLocation(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("there are no location", func() {
			It("should return empty result", func() {
				dbClient.
					ExpectQuery(searchStmt).
					WithArgs(
						p.Statuses[0],
					).
					WillReturnError(gorm.ErrRecordNotFound)

				res, err := fileRepo.SearchLocation(ctx, p)

				r := &repository.SearchLocationResult{
					Summary: repository.SearchLocationSummary{
						TotalItems: 0,
					},
					Items: []repository.SearchLocationItem{},
				}
				Expect(res).To(Equal(r))
				Expect(err).To(BeNil())
			})
		})

		When("failed count search location", func() {
			It("should return result", func() {
				searchRows = sqlmock.NewRows([]string{
					"id", "file_id", "barrel_id",
					"priority", "status",
				}).AddRow(
					"id1", "f1", "b1",
					2, "pending",
				)
				dbClient.
					ExpectQuery(searchStmt).
					WithArgs(
						p.Statuses[0],
					).
					WillReturnRows(searchRows)

				dbClient.
					ExpectQuery(countStmt).
					WithArgs(
						p.Statuses[0],
					).
					WillReturnError(fmt.Errorf("network error"))

				res, err := fileRepo.SearchLocation(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("there is one location", func() {
			It("should return result", func() {
				searchRows = sqlmock.NewRows([]string{
					"id", "file_id", "barrel_id",
					"priority", "status",
				}).AddRow(
					"id1", "f1", "b1",
					2, "pending",
				)
				dbClient.
					ExpectQuery(searchStmt).
					WithArgs(
						p.Statuses[0],
					).
					WillReturnRows(searchRows)

				countRows := sqlmock.
					NewRows([]string{"count(*)"}).
					AddRow(1)
				dbClient.
					ExpectQuery(countStmt).
					WithArgs(
						p.Statuses[0],
					).
					WillReturnRows(countRows)

				res, err := fileRepo.SearchLocation(ctx, p)

				r := &repository.SearchLocationResult{
					Summary: repository.SearchLocationSummary{
						TotalItems: 1,
					},
					Items: []repository.SearchLocationItem{
						{
							Id:       "id1",
							FileId:   "f1",
							BarrelId: "b1",
							Priority: 2,
							Status:   "pending",
						},
					},
				}
				Expect(res).To(Equal(r))
				Expect(err).To(BeNil())
			})
		})

		When("there are some locations", func() {
			It("should return result", func() {
				dbClient.
					ExpectQuery(searchStmt).
					WithArgs(
						p.Statuses[0],
					).
					WillReturnRows(searchRows)

				dbClient.
					ExpectQuery(countStmt).
					WithArgs(
						p.Statuses[0],
					).
					WillReturnRows(countRows)

				res, err := fileRepo.SearchLocation(ctx, p)

				Expect(res).To(Equal(r))
				Expect(err).To(BeNil())
			})
		})
	})

	Context("UpdateLocationByIds function", Label("unit"), func() {

		var (
			ctx        context.Context
			currentTs  time.Time
			dbClient   sqlmock.Sqlmock
			fileRepo   repository.File
			p          repository.UpdateLocationByIdsParam
			r          *repository.UpdateLocationByIdsResult
			updateStmt string
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

			gormClient, err := gorm.Open(gorm_mysql.New(gorm_mysql.Config{
				Conn:                      db,
				SkipInitializeWithVersion: true,
			}), &gorm.Config{
				DisableAutomaticPing: true,
			})
			if err != nil {
				AbortSuite("failed create gorm client: " + err.Error())
			}
			fileRepo = mysql.NewFile(mysql.FileParam{
				GormClient: gormClient,
			})

			p = repository.UpdateLocationByIdsParam{
				Ids:       []string{"i1", "i2", "i3"},
				Status:    typeconv.String("uploading"),
				UpdatedAt: currentTs,
			}
			r = &repository.UpdateLocationByIdsResult{
				TotalUpdated: 3,
			}
			updateStmt = regexp.QuoteMeta("UPDATE `file_location` SET `status`=?,`updated_at`=? WHERE id IN (?,?,?)")
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

				res, err := fileRepo.UpdateLocationByIds(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("begin error")))
			})
		})

		When("failed rollback during update location", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectExec(updateStmt).
					WithArgs(
						p.Status,
						p.UpdatedAt.UnixMilli(),
						p.Ids[0],
						p.Ids[1],
						p.Ids[2],
					).
					WillReturnError(fmt.Errorf("network error"))

				dbClient.
					ExpectRollback().
					WillReturnError(fmt.Errorf("rollback error"))

				res, err := fileRepo.UpdateLocationByIds(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("rollback error")))
			})
		})

		When("failed update location", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectExec(updateStmt).
					WithArgs(
						p.Status,
						p.UpdatedAt.UnixMilli(),
						p.Ids[0],
						p.Ids[1],
						p.Ids[2],
					).
					WillReturnError(fmt.Errorf("network error"))

				dbClient.
					ExpectRollback()

				res, err := fileRepo.UpdateLocationByIds(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("network error")))
			})
		})

		When("failed commit trx", func() {
			It("should return error", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectExec(updateStmt).
					WithArgs(
						p.Status,
						p.UpdatedAt.UnixMilli(),
						p.Ids[0],
						p.Ids[1],
						p.Ids[2],
					).
					WillReturnResult(sqlmock.NewResult(3, 3))

				dbClient.
					ExpectCommit().
					WillReturnError(fmt.Errorf("commit error"))

				res, err := fileRepo.UpdateLocationByIds(ctx, p)

				Expect(res).To(BeNil())
				Expect(err).To(Equal(fmt.Errorf("commit error")))
			})
		})

		When("success update location", func() {
			It("should return result", func() {
				dbClient.
					ExpectBegin()

				dbClient.
					ExpectExec(updateStmt).
					WithArgs(
						p.Status,
						p.UpdatedAt.UnixMilli(),
						p.Ids[0],
						p.Ids[1],
						p.Ids[2],
					).
					WillReturnResult(sqlmock.NewResult(3, 3))

				dbClient.
					ExpectCommit()

				res, err := fileRepo.UpdateLocationByIds(ctx, p)

				Expect(res).To(Equal(r))
				Expect(err).To(BeNil())
			})
		})

		When("success update all location data", func() {
			It("should return result", func() {
				p := repository.UpdateLocationByIdsParam{
					Ids:        []string{"i1", "i2", "i3"},
					Status:     typeconv.String("uploading"),
					ExternalId: typeconv.String("id"),
					UploadedAt: typeconv.Time(currentTs),
					UpdatedAt:  currentTs,
				}

				dbClient.
					ExpectBegin()

				updateStmt := regexp.QuoteMeta("UPDATE `file_location` SET `external_id`=?,`status`=?,`updated_at`=?,`uploaded_at`=? WHERE id IN (?,?,?)")
				dbClient.
					ExpectExec(updateStmt).
					WithArgs(
						p.ExternalId,
						p.Status,
						p.UpdatedAt.UnixMilli(),
						p.UploadedAt.UnixMilli(),
						p.Ids[0],
						p.Ids[1],
						p.Ids[2],
					).
					WillReturnResult(sqlmock.NewResult(3, 3))

				dbClient.
					ExpectCommit()

				res, err := fileRepo.UpdateLocationByIds(ctx, p)

				Expect(res).To(Equal(r))
				Expect(err).To(BeNil())
			})
		})

	})
})
