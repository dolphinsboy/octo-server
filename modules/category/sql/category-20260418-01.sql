-- +migrate Up

UPDATE `group_category` SET `name` = '默认分组' WHERE `is_default` = 1 AND `name` = '未分类';

-- +migrate Down
-- NOTE: 仅回退名称仍为 '默认分组' 的行。通过 DM_DEFAULT_CATEGORY_NAME 自定义的名称不会被回退。

UPDATE `group_category` SET `name` = '未分类' WHERE `is_default` = 1 AND `name` = '默认分组';
