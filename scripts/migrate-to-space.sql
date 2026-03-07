-- 数据迁移：创建"Demo Space"Space，将所有用户和 Bot 加入
-- 执行前确认：space_id 唯一，执行一次即可
-- 用法：mysql -u root dmwork < scripts/migrate-to-space.sql

-- 生成 space_id（固定值，便于引用）
SET @space_id = 'minglue_default';
SET @space_name = 'Demo Space';
SET @now = NOW();

-- 1. 创建 Space（如果不存在）
INSERT IGNORE INTO `space` (space_id, name, status, created_at, updated_at)
VALUES (@space_id, @space_name, 1, @now, @now);

-- 2. 将所有用户加入该 Space（排除已存在的）
INSERT IGNORE INTO `space_member` (space_id, uid, role, status, created_at, updated_at)
SELECT @space_id, uid, 0, 1, @now, @now
FROM `user`
WHERE status = 1;

-- 3. 设置第一个注册的用户为 Owner（role=2）
UPDATE `space_member`
SET role = 2
WHERE space_id = @space_id
AND uid = (SELECT uid FROM `user` WHERE status = 1 ORDER BY created_at ASC LIMIT 1);

-- 4. 将所有 Robot 也加入 Space（robot 表里的 uid）
INSERT IGNORE INTO `space_member` (space_id, uid, role, status, created_at, updated_at)
SELECT @space_id, uid, 0, 1, @now, @now
FROM `robot`
WHERE status = 1;

-- 验证
SELECT 'Space created' AS step, COUNT(*) AS count FROM `space` WHERE space_id = @space_id
UNION ALL
SELECT 'Members added', COUNT(*) FROM `space_member` WHERE space_id = @space_id AND status = 1;
