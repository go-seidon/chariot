
CREATE TABLE IF NOT EXISTS `file` (
  `id` VARCHAR(128) NOT NULL,
  `slug` VARCHAR(512) NOT NULL,
  `name` VARCHAR(4096) NOT NULL,
  `mimetype` VARCHAR(256) NOT NULL,
  `extension` VARCHAR(128) NOT NULL,
  `size` BIGINT NOT NULL,
  `visibility` VARCHAR(32) NOT NULL,
  `status` VARCHAR(32) NOT NULL,
  `uploaded_at` BIGINT NOT NULL,
  `created_at` BIGINT NOT NULL,
  `updated_at` BIGINT NOT NULL,
  `deleted_at` BIGINT NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE uk_slug(`slug`)
) 
DEFAULT CHARACTER SET utf8mb4
COLLATE utf8mb4_unicode_ci
ENGINE = InnoDB;

