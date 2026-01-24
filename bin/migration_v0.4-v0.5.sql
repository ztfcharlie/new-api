ALTER TABLE request_logs ADD COLUMN azure_request_id VARCHAR(100) DEFAULT '';
ALTER TABLE request_logs ADD COLUMN reject_flage BIGINT DEFAULT NULL;
