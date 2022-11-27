package mysql_test

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
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
						BarrelId:   "h1",
						ExternalId: typeconv.String("e1"),
						Priority:   1,
						Status:     "available",
						CreatedAt:  currentTs,
						UploadedAt: &currentTs,
					},
					{
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
			createLocationStmt = regexp.QuoteMeta("INSERT INTO `file_location` (`file_id`,`barrel_id`,`external_id`,`priority`,`status`,`uploaded_at`,`created_at`,`updated_at`) VALUES (?,?,?,?,?,?,?,?),(?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE `file_id`=VALUES(`file_id`)")
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
			findStmt = regexp.QuoteMeta("SELECT id, slug, name, mimetype, extension, size, visibility, status, created_at, updated_at, deleted_at, uploaded_at FROM `file` WHERE id = ? ORDER BY `file`.`id` LIMIT 1")
			findMetaStmt = regexp.QuoteMeta("SELECT file_id, `key`, value FROM `file_meta` WHERE `file_meta`.`file_id` = ?")
			findLocationStmt = regexp.QuoteMeta("SELECT file_id, barrel_id, external_id, priority, status, created_at, updated_at, uploaded_at FROM `file_location` WHERE `file_location`.`file_id` = ?")
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
				findStmt := regexp.QuoteMeta("SELECT id, slug, name, mimetype, extension, size, visibility, status, created_at, updated_at, deleted_at, uploaded_at FROM `file` WHERE slug = ? ORDER BY `file`.`id` LIMIT 1")
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
	})
})
