-- Migration: Standardize invoice and subscription status values
-- Date: 2026-06-04
-- Description: Simplify status to 3 values each:
--   Invoice:   Belum Bayar, Expired, Lunas
--   Langganan: Aktif, Suspended, Berhenti

-- ============================================
-- 1. INVOICE STATUS MIGRATION
-- ============================================

-- "Kadaluarsa" → "Expired"
UPDATE invoices
SET status_invoice = 'Expired', updated_at = NOW()
WHERE status_invoice = 'Kadaluarsa';

-- "Belum Dibayar" → "Belum Bayar" (jaga-jaga kalau ada data lama)
UPDATE invoices
SET status_invoice = 'Belum Bayar', updated_at = NOW()
WHERE status_invoice = 'Belum Dibayar';

-- Juga update di tabel archive kalau ada
UPDATE invoice_archives
SET status_invoice = 'Expired', updated_at = NOW()
WHERE status_invoice = 'Kadaluarsa';

UPDATE invoice_archives
SET status_invoice = 'Belum Bayar', updated_at = NOW()
WHERE status_invoice = 'Belum Dibayar';

-- ============================================
-- 2. VERIFICATION QUERY (jalankan setelah migration)
-- ============================================
-- Pastikan tidak ada status lama yang tersisa:
--
-- SELECT status_invoice, COUNT(*) FROM invoices GROUP BY status_invoice;
-- SELECT status_invoice, COUNT(*) FROM invoice_archives GROUP BY status_invoice;
--
-- Hasil yang diharapkan hanya: Belum Bayar, Expired, Lunas
--
-- SELECT status, COUNT(*) FROM langganan GROUP BY status;
--
-- Hasil yang diharapkan hanya: Aktif, Suspended, Berhenti
