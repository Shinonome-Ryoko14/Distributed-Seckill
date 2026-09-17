-- 0001_init.up.sql
-- 任务卡（里程碑 A · W1）
-- 目标：建立基础三表 users / products / seckill_activities（字段清单见计划 §2.1）。
-- 约束：
--   * 引擎 InnoDB，字符集 utf8mb4 / utf8mb4_unicode_ci（与 compose 服务端一致）。
--   * users.username 必须唯一（登录标识）。
--   * seckill_activities.seckill_stock 是库存唯一事实源，只有 order 服务写它（ADR-0002）。
-- 验收：
--   * make migrate 成功；三表均能 SHOW CREATE TABLE 出预期结构。
--   * SHOW INDEX FROM users 能看到 username 的 UNIQUE 索引。
--   * make migrate-down 能干净回滚（不报 "Unknown table"）。
--
-- 方向级提示：主键统一 BIGINT UNSIGNED AUTO_INCREMENT；时间用带默认值的 TIMESTAMP/DATETIME。
-- 金额类型、能否为 NULL、索引/外键的选择属你的设计考点，请自行决策并留有依据。
--
-- 字段清单（列名仅作规格，DDL 类型与约束由你补全）：
--   users(id, username, password_hash, created_at, updated_at)
--   products(id, name, description, price, status, created_at, updated_at)
--   seckill_activities(id, product_id, seckill_stock, seckill_price,
--                      start_time, end_time, status, created_at, updated_at)

CREATE TABLE users(
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,

  username VARCHAR(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_as_cs NOT NULL,
  -- 大小写敏感判定，给该列指定排序规则。

  password_hash VARCHAR(255) NOT NULL,
  -- bcrypt输出固定60已足够, 但这里预留255，为以后做优化留空间
  
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY(id),
  UNIQUE KEY uk_users_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE products(
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,

  name VARCHAR(64) NOT NULL,

  description VARCHAR(255) NOT NULL,

  price INT UNSIGNED NOT NULL COMMENT '单位：分',

  status TINYINT NOT NULL,
  -- 关于为什么选择TINYINT：扩展性强，与状态码兼容性好，
  -- 可读性稍差但可利用COMMENT和代码层克服

  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE seckill_activities(
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,

  product_id BIGINT UNSIGNED NOT NULL,

  seckill_stock INT UNSIGNED NOT NULL,

  seckill_price INT UNSIGNED NOT NULL,

  start_time TIMESTAMP NOT NULL,

  end_time TIMESTAMP NOT NULL,

  status TINYINT NOT NULL,

  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY(id),
  CONSTRAINT seckill_activities_product FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE RESTRICT,
  KEY idx_product_id (product_id),
  KEY idx_status_start_time(status, start_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
