-- 清理旧的测试数据
DELETE FROM logs WHERE type = 5 AND content LIKE '%429 error simulation%';

-- 清理旧的统计数据
DELETE FROM rate_limit_429_stats;