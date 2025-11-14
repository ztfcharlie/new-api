-- Migration from v0.4 to v0.5
-- Modify task_id column to support longer task IDs (VARCHAR(256))

ALTER TABLE `tasks` CHANGE `task_id` `task_id` VARCHAR(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL;