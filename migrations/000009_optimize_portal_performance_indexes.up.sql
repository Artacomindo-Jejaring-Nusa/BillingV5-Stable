-- Migration: Optimize performance with database indexes for customer lookups and portal traffic
-- Tables affected: pelanggan, langganan, invoices, users, data_teknis

-- 1. Table: pelanggan
CREATE INDEX idx_pelanggan_no_telp ON pelanggan(no_telp);
CREATE INDEX idx_pelanggan_nama ON pelanggan(nama);
CREATE INDEX idx_pelanggan_no_ktp ON pelanggan(no_ktp);
CREATE INDEX idx_pelanggan_brand_default ON pelanggan(brand_default);

-- 2. Table: langganan
CREATE INDEX idx_langganan_pelanggan_status ON langganan(pelanggan_id, status);
CREATE INDEX idx_langganan_metode_pembayaran ON langganan(metode_pembayaran);

-- 3. Table: invoices
CREATE INDEX idx_invoices_status_invoice ON invoices(status_invoice);
CREATE INDEX idx_invoices_pelanggan_status ON invoices(pelanggan_id, status_invoice);
CREATE INDEX idx_invoices_no_telp ON invoices(no_telp);
CREATE INDEX idx_invoices_email ON invoices(email);
CREATE INDEX idx_invoices_xendit_external_id ON invoices(xendit_external_id);

-- 4. Table: users
CREATE INDEX idx_users_phone_no ON users(phone_no);
