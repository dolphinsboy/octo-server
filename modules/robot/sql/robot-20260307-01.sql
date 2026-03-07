-- +migrate Up

-- access_mode: 0=需要审批（默认） 1=自动通过（公共AI） 2=禁止申请（私有）
-- 存量数据默认1（自动通过），保持向后兼容
ALTER TABLE `robot` ADD COLUMN `access_mode` TINYINT NOT NULL DEFAULT 1 COMMENT 'AI访问模式: 0=需审批 1=自动通过 2=禁止申请';

-- 机器人访问申请表
CREATE TABLE IF NOT EXISTS `robot_apply` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `uid` VARCHAR(40) NOT NULL COMMENT '申请人UID',
  `robot_uid` VARCHAR(40) NOT NULL COMMENT '目标AI的UID',
  `owner_uid` VARCHAR(40) NOT NULL COMMENT 'AI Owner的UID（冗余字段，加速查询）',
  `remark` VARCHAR(200) NOT NULL DEFAULT '' COMMENT '申请理由',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态: 0=待处理 1=通过 2=拒绝',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uid_robot_pending` (`uid`, `robot_uid`, `status`),
  KEY `idx_owner_status` (`owner_uid`, `status`),
  KEY `idx_robot_status` (`robot_uid`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='机器人访问申请表';
