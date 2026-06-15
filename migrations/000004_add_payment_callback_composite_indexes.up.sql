-- Migration: Add composite indexes to payment_callback_logs for performance optimization
-- Composite Index 1: (xendit_id, status) -> idx_xendit_status
-- Composite Index 2: (external_id, xendit_id) -> idx_external_xendit

CREATE INDEX idx_xendit_status ON payment_callback_logs(xendit_id, status);
CREATE INDEX idx_external_xendit ON payment_callback_logs(external_id, xendit_id);
