
ALTER TABLE `file_location` ADD INDEX idx_file_id(`file_id`);

ALTER TABLE `file_location` ADD INDEX idx_barrel_id(`barrel_id`);

ALTER TABLE `file_location` 
  ADD CONSTRAINT `ibfk_fl_file_id` 
  FOREIGN KEY (`file_id`) REFERENCES `file`(`id`) 
  ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE `file_location` 
  ADD CONSTRAINT `ibfk_fl_barrel_id` 
  FOREIGN KEY (`barrel_id`) REFERENCES `barrel`(`id`) 
  ON DELETE RESTRICT ON UPDATE CASCADE;
