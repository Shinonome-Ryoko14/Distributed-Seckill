-- 0001_init.down.sql
-- 目标：回滚 0001_init.up.sql，使 schema 回到迁移前状态。
-- 验收：make migrate-down 后三张表均不存在。
--
-- 方向级提示：DROP 顺序与建表顺序相反，注意表间引用关系。

DROP TABLE IF EXISTS seckill_activities;

DROP TABLE IF EXISTS products;

DROP TABLE IF EXISTS users;
