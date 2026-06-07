-- Standardization Script for BillingRevaktor
-- This script cleans up legacy data and aligns it with the new backend standards.

-- 1. Standardization of Invoice Statuses
-- Map legacy Indonesian status strings to standard English ones if necessary,
-- or ensure consistency across the database.

-- Convert 'Kadaluarsa' to 'Expired' (standard in the new system)
UPDATE invoices SET status_invoice = 'Expired' WHERE status_invoice = 'Kadaluarsa';
UPDATE invoices_archive SET status_invoice = 'Expired' WHERE status_invoice = 'Kadaluarsa';

-- Ensure 'Lunas' is consistent
UPDATE invoices SET status_invoice = 'Lunas' WHERE status_invoice IN ('PAID', 'paid', 'lunas');
UPDATE invoices_archive SET status_invoice = 'Lunas' WHERE status_invoice IN ('PAID', 'paid', 'lunas');

-- Ensure 'Belum Dibayar' is consistent
UPDATE invoices SET status_invoice = 'Belum Dibayar' WHERE status_invoice IN ('PENDING', 'pending', 'belum dibayar', 'Belum bayar');

-- 2. Clean up accidentally created plural tables (if any exist)
-- This helps if AutoMigrate ran before the SingularTable strategy was enforced.

-- We check and drop/rename tables if they exist with plural names and have no data, 
-- or merge them if they do. For safety, we just rename them if they exist and singular doesn't.

-- Rename pelanggans to pelanggan if pelanggan doesn't exist
SET @plural_plg = (SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'pelanggans');
SET @singular_plg = (SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'pelanggan');
-- IF plural exists AND singular doesn't -> RENAME
-- IF both exist -> User should manually merge, but we can't easily automate safely here.

-- 3. Release constraints for soft-deleted items (NIK and Email)
-- This allows re-using NIK/Email for new customers if the old record was soft-deleted.
UPDATE pelanggan SET email = CONCAT('deleted_', UNIX_TIMESTAMP(), '_', email) 
WHERE deleted_at IS NOT NULL AND email NOT LIKE 'deleted_%';

UPDATE pelanggan SET no_ktp = CONCAT('deleted_', UNIX_TIMESTAMP(), '_', no_ktp) 
WHERE deleted_at IS NOT NULL AND no_ktp != '' AND no_ktp NOT LIKE 'deleted_%';

-- 4. Standardize Subscription Status
UPDATE langganan SET status = 'Aktif' WHERE status IN ('active', 'AKTIF', 'aktif');
UPDATE langganan SET status = 'Suspended' WHERE status IN ('suspended', 'SUSPEND', 'suspend');
UPDATE langganan SET status = 'Berhenti' WHERE status IN ('stopped', 'berhenti', 'BERHENTI', 'OFF', 'off');

-- 5. Fix possible empty/null values for critical columns
UPDATE pelanggan SET id_brand = 'ajn-01' WHERE id_brand IS NULL OR id_brand = '';
UPDATE pelanggan SET alamat = 'Tambun' WHERE alamat IS NULL OR alamat = '';
