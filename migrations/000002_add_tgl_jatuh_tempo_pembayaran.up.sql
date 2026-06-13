-- Migration: Add tgl_jatuh_tempo_pembayaran column to langganan table
-- This column stores the explicit payment due date (Kotak 3) separately from 
-- the usage period end date (tgl_jatuh_tempo / Kotak 2).
-- When generating invoices, this field is used as the invoice's tgl_jatuh_tempo.

ALTER TABLE langganan ADD COLUMN tgl_jatuh_tempo_pembayaran DATE NULL AFTER tgl_jatuh_tempo;

-- Backfill: Set existing records' tgl_jatuh_tempo_pembayaran to match tgl_jatuh_tempo
-- so that existing data behaves the same as before.
UPDATE langganan SET tgl_jatuh_tempo_pembayaran = tgl_jatuh_tempo WHERE tgl_jatuh_tempo IS NOT NULL;

-- Add index for performance
CREATE INDEX idx_langganan_tgl_jt_pembayaran ON langganan(tgl_jatuh_tempo_pembayaran);
