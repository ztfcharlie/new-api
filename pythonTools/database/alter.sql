
-- 2. 修改现有表结构
-- channels 表
ALTER TABLE `channels` 
ADD COLUMN `settings` longtext,
ADD COLUMN `channel_info` json DEFAULT NULL;

-- logs 表
ALTER TABLE `logs` MODIFY COLUMN `is_stream` tinyint(1) DEFAULT NULL,ADD COLUMN `ip` varchar(191) DEFAULT '';


-- redemptions 表
ALTER TABLE `redemptions` ADD COLUMN `expired_time` bigint DEFAULT NULL;

-- tokens 表
ALTER TABLE `tokens` MODIFY COLUMN `unlimited_quota` tinyint(1) DEFAULT NULL,MODIFY COLUMN `model_limits_enabled` tinyint(1) DEFAULT NULL;

-- top_ups 表
ALTER TABLE `top_ups` 
ADD COLUMN `complete_time` bigint DEFAULT NULL,
MODIFY COLUMN `trade_no` varchar(255) DEFAULT NULL,
ADD UNIQUE KEY `trade_no` (`trade_no`),
ADD KEY `idx_top_ups_trade_no` (`trade_no`);

-- users 表
ALTER TABLE `users` 
ADD COLUMN `remark` varchar(255) DEFAULT NULL,
ADD COLUMN `stripe_customer` varchar(64) DEFAULT NULL,
ADD KEY `idx_users_stripe_customer` (`stripe_customer`);
