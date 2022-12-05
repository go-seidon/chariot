
CREATE TABLE IF NOT EXISTS `file_location` (
  `id` VARCHAR(128) NOT NULL,
  `file_id` VARCHAR(128) NOT NULL,
  `barrel_id` VARCHAR(128) NOT NULL,
  `external_id` VARCHAR(256) NULL DEFAULT NULL,
  `priority` INT NOT NULL,
  `status` VARCHAR(32) NOT NULL,
  `uploaded_at` BIGINT NULL DEFAULT NULL,
  `created_at` BIGINT NOT NULL,
  `updated_at` BIGINT NOT NULL,
  `deleted_at` BIGINT NULL DEFAULT NULL,
  PRIMARY KEY (`id`)
) 
DEFAULT CHARACTER SET utf8mb4
COLLATE utf8mb4_unicode_ci
ENGINE = InnoDB;
