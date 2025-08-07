-- Migration: For Task#1 Add last_reset field to api_keys
ALTER TABLE api_keys ADD COLUMN last_reset DATETIME DEFAULT CURRENT_TIMESTAMP;
