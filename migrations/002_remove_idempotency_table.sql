-- Migration to remove idempotency_keys table since we're now using Redis
-- File: 002_remove_idempotency_table.sql
-- Drop indexes first
DROP INDEX IF EXISTS idx_idempotency_student_id;

DROP INDEX IF EXISTS idx_idempotency_expires_at;

-- Drop the idempotency_keys table
DROP TABLE IF EXISTS idempotency_keys;