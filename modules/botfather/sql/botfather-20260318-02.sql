-- +migrate Up
ALTER TABLE `user_api_key` ADD COLUMN `space_id` varchar(40) NOT NULL DEFAULT '' COMMENT '绑定的Space ID' AFTER `api_key`;
