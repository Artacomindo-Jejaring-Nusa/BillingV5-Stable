-- Cleanup and Consolidation Script for Permissions
-- This script merges redundant permissions and renames them to match the system standards.

-- 1. Standardize existing names (fixing typos or alternate names)
UPDATE permissions SET name = 'view_data_teknis' WHERE name = 'view_teknis';
UPDATE permissions SET name = 'create_data_teknis' WHERE name = 'create_teknis';
UPDATE permissions SET name = 'edit_data_teknis' WHERE name = 'edit_teknis';
UPDATE permissions SET name = 'delete_data_teknis' WHERE name = 'delete_teknis';

UPDATE permissions SET name = 'view_brand_&_paket' WHERE name = 'view_paket' OR name = 'view_brand_paket';
UPDATE permissions SET name = 'create_brand_&_paket' WHERE name = 'create_paket' OR name = 'create_brand_paket';
UPDATE permissions SET name = 'edit_brand_&_paket' WHERE name = 'edit_paket' OR name = 'edit_brand_paket';
UPDATE permissions SET name = 'delete_brand_&_paket' WHERE name = 'delete_paket' OR name = 'delete_brand_paket';

UPDATE permissions SET name = 'view_reports_revenue' WHERE name = 'view_reports' OR name = 'view_laporan_pendapatan';
UPDATE permissions SET name = 'create_reports_revenue' WHERE name = 'create_reports' OR name = 'create_laporan_pendapatan';
UPDATE permissions SET name = 'edit_reports_revenue' WHERE name = 'edit_reports' OR name = 'edit_laporan_pendapatan';
UPDATE permissions SET name = 'delete_reports_revenue' WHERE name = 'delete_reports' OR name = 'delete_laporan_pendapatan';

UPDATE permissions SET name = 'view_simulasi_harga' WHERE name = 'view_simulasi';
UPDATE permissions SET name = 'create_simulasi_harga' WHERE name = 'create_simulasi';
UPDATE permissions SET name = 'edit_simulasi_harga' WHERE name = 'edit_simulasi';
UPDATE permissions SET name = 'delete_simulasi_harga' WHERE name = 'delete_simulasi';

UPDATE permissions SET name = 'view_mikrotik_servers' WHERE name = 'view_servers';
UPDATE permissions SET name = 'create_mikrotik_servers' WHERE name = 'create_servers';
UPDATE permissions SET name = 'edit_mikrotik_servers' WHERE name = 'edit_servers';
UPDATE permissions SET name = 'delete_mikrotik_servers' WHERE name = 'delete_servers';

UPDATE permissions SET name = 'view_odp_management' WHERE name = 'view_odp';
UPDATE permissions SET name = 'create_odp_management' WHERE name = 'create_odp';
UPDATE permissions SET name = 'edit_odp_management' WHERE name = 'edit_odp';
UPDATE permissions SET name = 'delete_odp_management' WHERE name = 'delete_odp';

UPDATE permissions SET name = 'manage_sk' WHERE name = 'create_kelola_s&k' OR name = 'edit_kelola_s&k';

-- 2. Consolidate Duplicates (Handle the cases where multiple records might now have the same name)
-- We find duplicates and update role_has_permissions to use only one ID, then delete the other.
CREATE TEMPORARY TABLE temp_perms_to_delete AS
SELECT p2.id as delete_id, p1.id as keep_id
FROM permissions p1
JOIN permissions p2 ON p1.name = p2.name AND p1.id < p2.id;

UPDATE IGNORE role_has_permissions rhp
JOIN temp_perms_to_delete tptd ON rhp.permission_id = tptd.delete_id
SET rhp.permission_id = tptd.keep_id;

DELETE FROM role_has_permissions WHERE permission_id IN (SELECT delete_id FROM temp_perms_to_delete);
DELETE FROM permissions WHERE id IN (SELECT delete_id FROM temp_perms_to_delete);

DROP TEMPORARY TABLE temp_perms_to_delete;

-- 3. Cleanup unused/garbage permissions
DELETE FROM permissions WHERE name IN ('create_s&k', 'view_s&k', 'edit_s&k', 'delete_s&k', 'create_sk', 'edit_sk', 'delete_sk');
-- Note: keep view_sk if used, but seeder will ensure standardized ones exist.
