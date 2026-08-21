-- Revert Migration: Remove portal performance optimization indexes

DROP INDEX idx_pelanggan_no_telp ON pelanggan;
DROP INDEX idx_pelanggan_nama ON pelanggan;
DROP INDEX idx_pelanggan_no_ktp ON pelanggan;
DROP INDEX idx_pelanggan_brand_default ON pelanggan;

DROP INDEX idx_langganan_pelanggan_status ON langganan;
DROP INDEX idx_langganan_metode_pembayaran ON langganan;

DROP INDEX idx_invoices_status_invoice ON invoices;
DROP INDEX idx_invoices_pelanggan_status ON invoices;
DROP INDEX idx_invoices_no_telp ON invoices;
DROP INDEX idx_invoices_email ON invoices;
DROP INDEX idx_invoices_xendit_external_id ON invoices;

DROP INDEX idx_users_phone_no ON users;
