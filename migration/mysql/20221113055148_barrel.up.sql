
CREATE TABLE IF NOT EXISTS `barrel` (
  `id` VARCHAR(128) NOT NULL,
  `code` VARCHAR(256) NOT NULL,
  `name` VARCHAR(128) NOT NULL,
  `provider` VARCHAR(32) NOT NULL,
  `status` VARCHAR(16) NOT NULL,
  `created_at` BIGINT NOT NULL,
  `updated_at` BIGINT NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE uk_code(`code`)
) 
DEFAULT CHARACTER SET utf8mb4
COLLATE utf8mb4_unicode_ci
ENGINE = InnoDB;

