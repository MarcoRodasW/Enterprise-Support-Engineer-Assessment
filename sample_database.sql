CREATE DATABASE IF NOT EXISTS wikimedia_prod;
USE wikimedia_prod;

CREATE TABLE users (
  id INT AUTO_INCREMENT PRIMARY KEY,
  username VARCHAR(50) NOT NULL,
  email VARCHAR(100) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  last_login DATETIME,
  is_active BOOLEAN DEFAULT 1,
  permission_level ENUM('read','write','admin') DEFAULT 'read'
);

CREATE TABLE api_keys (
  id INT AUTO_INCREMENT PRIMARY KEY,
  user_id INT NOT NULL,
  api_key VARCHAR(64) UNIQUE NOT NULL,
  rate_limit INT DEFAULT 100,
  calls_made INT DEFAULT 0,
  is_valid BOOLEAN DEFAULT 1,
  expires_at DATE,
  last_reset DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE audit_logs (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  api_key VARCHAR(64) NOT NULL,
  endpoint VARCHAR(255) NOT NULL,
  response_code SMALLINT NOT NULL,
  response_time_ms INT NOT NULL,
  timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  error_message TEXT
);

-- Sample data
INSERT INTO users (username, email, permission_level) VALUES
('data_importer', 'importer@example.com', 'write'),
('api_consumer', 'consumer@example.com', 'read');

INSERT INTO api_keys (user_id, api_key, rate_limit, expires_at) VALUES
(1, 'key_ABC123', 5, '2024-12-31'),  -- Low limit for testing
(2, 'key_DEF456', 5, '2024-12-31');

INSERT INTO audit_logs (api_key, endpoint, response_code, response_time_ms, error_message) VALUES
('key_ABC123', '/api/export', 200, 120, NULL),
('key_ABC123', '/api/export', 200, 110, NULL),
('key_DEF456', '/api/audit', 200, 85, NULL);

-- FIXME 3: Create GDPR-compliant masked view
-- CREATE VIEW masked_audit_logs AS
-- SELECT 
--   id,
--   CONCAT(LEFT(api_key, 6), '...') AS masked_key,
--   endpoint,
--   response_code,
--   response_time_ms,
--   timestamp
-- FROM audit_logs;