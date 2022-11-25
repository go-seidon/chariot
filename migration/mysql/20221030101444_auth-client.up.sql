
CREATE TABLE IF NOT EXISTS `auth_client` (
  `id` VARCHAR(128) NOT NULL,
  `client_id` VARCHAR(256) NOT NULL,
  `client_secret` TEXT NOT NULL,
  `name` VARCHAR(128) NOT NULL,
  `type` VARCHAR(32) NOT NULL,
  `status` VARCHAR(16) NOT NULL,
  `created_at` BIGINT NOT NULL,
  `updated_at` BIGINT NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE uk_client_id(`client_id`)
) 
DEFAULT CHARACTER SET utf8mb4
COLLATE utf8mb4_unicode_ci
ENGINE = InnoDB;
