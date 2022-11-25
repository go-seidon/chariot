
ALTER TABLE `file_location` DROP FOREIGN KEY `ibfk_fl_file_id`;

ALTER TABLE `file_location` DROP FOREIGN KEY `ibfk_fl_barrel_id`;

ALTER TABLE `file_location` DROP INDEX `idx_file_id`;

ALTER TABLE `file_location` DROP INDEX `idx_barrel_id`;

