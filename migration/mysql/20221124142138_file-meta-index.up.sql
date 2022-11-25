
ALTER TABLE `file_meta` ADD INDEX idx_file_id(`file_id`);

ALTER TABLE `file_meta` 
  ADD CONSTRAINT `ibfk_fm_file_id` 
  FOREIGN KEY (`file_id`) REFERENCES `file`(`id`) 
  ON DELETE CASCADE ON UPDATE CASCADE;
